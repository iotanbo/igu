package linuser

import (
	"fmt"
	"strings"

	"github.com/iotanbo/igu/pkg/ec"
	"github.com/iotanbo/igu/pkg/fu"
	"github.com/iotanbo/igu/pkg/sh"

	//"github.com/iotanbo/igu/pkg/sh"
	//"github.com/iotanbo/igu/pkg/ec"
	//lint:ignore ST1001 - for clear and concise error handling.
	. "github.com/iotanbo/igu/pkg/errs"
)

// UserDescriptor describes user to be created with CreateUser function.
type UserDescriptor struct {
	UserName    string
	UserId      int // generated by OS if 0
	GroupName   string
	GroupId     int // generated by OS if 0
	Password    string
	Home        string // e.g. `/home/myUser`
	Shell       string // `/usr/sbin/nologin` will be used by default
	Description string
	IsSystem    bool // set to true if this is a system user (id < 1000)
}

func run(cmd string, verbose bool) Err {
	// shell := "bash"
	// args := []string{
	// 	"-c",
	// 	cmd,
	// }
	stdout, stderr, e := sh.ExecuteLine(cmd, 10000)
	if verbose {
		if len(stdout) > 0 {
			fmt.Printf("'%s'\n", stdout)
		}
	}
	if e.Some() {
		fmt.Printf("'%s'\n", stderr)
		errLowerCase := strings.ToLower(stderr)
		if strings.Contains(errLowerCase, "permission denied") {
			return Err{Code: ec.PermissionDenied, Msg: stderr}
		}
		if strings.Contains(errLowerCase, "already exists") {
			return Err{Code: ec.AlreadyExists, Msg: stderr}
		}
		// Returned by gpasswd when user is not a member of the group
		if strings.Contains(errLowerCase, "not a member") {
			return Err{Code: ec.NothingDone, Msg: stderr}
		}
		//
		e.Msg = stderr
		return e
	}
	return NoError
}

func GroupExists(groupName string) (bool, Err) {
	contents, e := fu.ReadLines("/etc/group")
	if e.Some() {
		return false, FromError(e)
	}
	prefix := fmt.Sprintf("%s:", groupName)
	for _, line := range contents {
		if strings.HasPrefix(line, prefix) {
			return true, NoError
		}
	}
	return false, NoError
}

func UserExists(userName string) (bool, Err) {
	contents, e := fu.ReadLines("/etc/passwd")
	if e.Some() {
		return false, FromError(e)
	}
	prefix := fmt.Sprintf("%s:", userName)
	for _, line := range contents {
		if strings.HasPrefix(line, prefix) {
			return true, NoError
		}
	}
	return false, NoError
}

// CreateGroup creates a linux user group.
// If groupId not specified, default system ID will be created.
// `verbose` parameter allows printing additional info and stdout.
//
// Returned errors:
//	ec.NoError // completed successfully
//	ec.AlreadyExists // group already exists
//	ec.ProcessExit // sub-process exited with non-zero error code
//	ec.TimedOut // execution took longer than 10 seconds
//	ec.PermissionDenied
// Other errors may be returned for other situations.
//
// Usage example:
//	e := CreateGroup("newUserGroup", false, 1001)
//	if e.Some() { /* handle errors */ }
func CreateGroup(groupName string, verbose bool, groupId ...int) Err {
	gid := 0
	if len(groupId) > 0 {
		gid = groupId[0]
	}
	alreadyExists, e := GroupExists(groupName)
	if e.Some() {
		return e
	}
	if alreadyExists {
		return Err{Code: ec.AlreadyExists}
	}
	groupIdParams := ""
	if gid != 0 {
		groupIdParams = fmt.Sprintf("-g %d", gid)
	}
	cmd := fmt.Sprintf("groupadd %s %s", groupIdParams, groupName)
	if verbose {
		fmt.Printf("* CreateGroup: executing command '%s'\n", cmd)
	}
	return run(cmd, verbose)
}

// DeleteGroup deletes a linux user group.
// It will delete the group even if it is the primary group of a user.
// `verbose` parameter allows printing additional info and stdout.
//
// Returned errors:
//	ec.NoError // completed successfully
//	ec.NotFound // group not found
//	ec.ProcessExit // sub-process exited with non-zero error code
//	ec.TimedOut // execution took longer than 10 seconds
//	ec.PermissionDenied
// Other errors may be returned for other situations.
//
// Usage example:
//	e := DeleteGroup("userGroupName", false)
//	if e.Some() { /* handle errors */ }
func DeleteGroup(groupName string, verbose bool) Err {
	alreadyExists, e := GroupExists(groupName)
	if e.Some() {
		return e
	}
	if !alreadyExists {
		return Err{Code: ec.NotFound}
	}
	cmd := fmt.Sprintf("groupdel -f %s", groupName)
	if verbose {
		fmt.Printf("* DeleteGroup: executing command '%s'\n", cmd)
	}
	return run(cmd, verbose)
}

// CreateUser creates a linux user defined by UserDescriptor.
// System users differ only by UID (up to 1000 on Ubuntu)
// and some system preferences like login screen.
// Sometimes the default-created user ID and group ID may differ,
// in order to ensure thir equality, pass the IDs explicitely.
// If `Shell` not specified, `/usr/sbin/nologin` will be used.
// If specified group does not exist, it will be created.
//
// Args:
//	ud // UserDescriptor.
//	verbose // verbose mode.
//
// Returned errors:
//	ec.NoError // completed successfully
//	ec.AlreadyExists // user already exists
//	ec.ProcessExit // sub-process exited with non-zero error code
//	ec.TimedOut // execution took longer than 10 seconds
//	ec.PermissionDenied
// Other errors may be returned for other situations.
//
// Usage example:
//	ud := UserDescriptor {
//		UserName:  "myUser",
//		GroupName: "myUser",
//		Home:      "/home/myUser",
//		Shell:     "/bin/bash",
//		Password:  "test",
//	}
//	e := CreateUser(ud, false)
//	if e.Some() { /* handle errors */ }
func CreateUser(
	ud UserDescriptor,
	verbose bool) Err {
	userExists, e := UserExists(ud.UserName)
	if e.Some() {
		return e
	}
	if userExists {
		return Err{Code: ec.AlreadyExists}
	}

	// https://linux.die.net/man/8/useradd

	homeDirParams := "--no-create-home"
	if len(ud.Home) > 0 {
		homeDirParams = fmt.Sprintf("--create-home --home %s", ud.Home)
	}
	userParams := ""
	if ud.UserId > 0 {
		userParams = fmt.Sprintf("--uid %d", ud.UserId)
	}
	sysUserParams := ""
	if ud.IsSystem {
		sysUserParams = "--system"
	}
	shellParams := "--shell /usr/sbin/nologin"
	if len(ud.Shell) > 0 {
		shellParams = fmt.Sprintf("--shell %s", ud.Shell)
	}
	descParams := ""
	if len(ud.Description) > 0 {
		descParams = fmt.Sprintf("--comment %s", ud.Description)
	}
	groupParams := "--no-user-group"
	if len(ud.GroupName) > 0 {
		groupExists, e := GroupExists(ud.GroupName)
		if e.Some() {
			return e
		}
		if !groupExists {
			e = CreateGroup(ud.GroupName, verbose, ud.GroupId)
			if e.Some() {
				return e
			}
		}
		groupParams = fmt.Sprintf("-g %s", ud.GroupName)
	}

	cmd := fmt.Sprintf("useradd --no-log-init %s %s %s %s %s %s %s", sysUserParams, descParams, groupParams, homeDirParams,
		shellParams, userParams, ud.UserName)

	if verbose {
		fmt.Printf("* CreateUser: executing command '%s'\n", cmd)
	}
	e = run(cmd, verbose)
	if e.Some() {
		return e
	}
	// Set password if specified
	if len(ud.Password) > 0 {
		cmd = fmt.Sprintf("echo %s:%s | chpasswd", ud.UserName, ud.Password)
		if verbose {
			fmt.Printf("* CreateUser: setting user password.\n")
		}
		return run(cmd, verbose)
	}
	return NoError
}

// DeleteUser deletes a linux user, its primary group (only if there are
// no more members in it and it has the same name as user (GID does not matter)),
// and optionally the user home directory.
// Before deletion, the user will be removed from any group it belongs to.
// In case there are other users in this user's primary group,
// the group will not be deleted, you have to do it manually.
// `deleteHomeDir` allows to delete user's home directory.
// `verbose` allows printing additional info and stdout.
//
// Returned errors:
//	ec.NoError // completed successfully
//	ec.NotFound // user not found
//	ec.ProcessExit // sub-process exited with non-zero error code
//	ec.TimedOut // execution took longer than 10 seconds
//	ec.PermissionDenied
// Other errors may be returned for other situations.
//
// Usage example:
//	e := DeleteUser("userName", true, false)
//	if e.Some() { /* handle errors */ }
func DeleteUser(userName string, deleteHomeDir bool, verbose bool) Err {
	alreadyExists, e := UserExists(userName)
	if e.Some() {
		return e
	}
	if !alreadyExists {
		return Err{Code: ec.NotFound}
	}
	removeHomeDirParams := ""
	if deleteHomeDir {
		// -r: remove home directory, -f: remove files owned by other users
		removeHomeDirParams = "-rf"
	}
	cmd := fmt.Sprintf("userdel %s %s", removeHomeDirParams, userName)
	if verbose {
		fmt.Printf("* DeleteUser: executing '%s'\n", cmd)
	}
	return run(cmd, verbose)
}

// Returns NoError if both user and group exist, ec.NotFound with
// corresponding message if at least one does not exist,
// or other errors for other situations.
func ensureUserAndGroupExist(userName string, groupName string) Err {
	userExists, e := UserExists(userName)
	if e.Some() {
		return e
	}
	if !userExists {
		return Err{
			Code: ec.NotFound,
			Msg:  fmt.Sprintf("user %s does not exist", userName),
		}
	}
	groupExists, e := GroupExists(groupName)
	if e.Some() {
		return e
	}
	if !groupExists {
		return Err{
			Code: ec.NotFound,
			Msg:  fmt.Sprintf("group %s does not exist", groupName),
		}
	}
	return NoError
}

// AddUserToGroup adds existing user to an existing group.
// Both user and group must exist, otherwise `ec.NotFound` error is returned.
// If user is already in the group, NoError is returned.
func AddUserToGroup(userName string, groupName string, verbose ...bool) Err {
	e := ensureUserAndGroupExist(userName, groupName)
	if e.Some() {
		return e
	}
	cmd := fmt.Sprintf("usermod -a -G %s %s", groupName, userName)
	v := false
	if len(verbose) > 0 {
		v = verbose[0]
	}
	if v {
		fmt.Printf("* AddUserToGroup: executing '%s'\n", cmd)
	}
	return run(cmd, v)
}

// RemoveUserFromGroup removes user from group.
// Both user and group must exist, otherwise `ec.NotFound` error is returned.
// If user is not a member of the group, `ec.NothingDone` is returned.
func RemoveUserFromGroup(userName string, groupName string, verbose ...bool) Err {
	e := ensureUserAndGroupExist(userName, groupName)
	if e.Some() {
		return e
	}
	cmd := fmt.Sprintf("gpasswd --delete %s %s", userName, groupName)
	v := false
	if len(verbose) > 0 {
		v = verbose[0]
	}
	if v {
		fmt.Printf("* RemoveUserFromGroup: executing '%s'\n", cmd)
	}
	return run(cmd, v)
}

// IsUserInGroup returns (true, NoError) if both user and group exist
// and the user is in the group.
func IsUserInGroup(userName string, groupName string, verbose ...bool) (bool, Err) {
	e := ensureUserAndGroupExist(userName, groupName)
	if e.Some() {
		return false, e
	}

	contents, e := fu.ReadLines("/etc/group")
	if e.Some() {
		return false, FromError(e)
	}
	gPrefix := fmt.Sprintf("%s:", groupName)
	// Deal with the case when user name
	// matches part of another user's name.
	// To distinguish them, use comma, space or colon
	// to ensure that the full name is found.
	uSuffixes := []string{
		fmt.Sprintf(",%s", userName),
		fmt.Sprintf(":%s", userName),
	}
	for _, line := range contents {
		if strings.HasPrefix(line, gPrefix) {
			// Handle case when groupName equals to userName
			if userName == groupName {
				return true, NoError
			}
			// groupName and userName are different
			for _, s := range uSuffixes {
				if strings.HasSuffix(line, s) {
					return true, NoError
				}
			}
			if strings.Contains(line, fmt.Sprintf(":%s,", userName)) {
				return true, NoError
			}
		}
	}
	return false, NoError
}