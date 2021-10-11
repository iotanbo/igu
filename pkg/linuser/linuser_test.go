package linuser

import (
	"fmt"
	"runtime"
	"testing"

	//"github.com/iotanbo/igu/pkg/ec"
	"github.com/iotanbo/igu/pkg/ec"
	"github.com/stretchr/testify/require"
)

// Type aliases to improve readability
var printf = fmt.Printf
var expect = require.True

func TestGroupExists(t *testing.T) {
	if runtime.GOOS != "windows" { // unix
		printf("-- Running TestGroupExists on ...nix.\n")

		// Group 'sudo' should exist
		exists, e := GroupExists("tty")
		expect(t, e.None(), "GroupExists('tty') returned error: %v.", e)
		expect(t, exists, "Expected GroupExists('tty')==true, got false.")

		// Group 'notExistingGroup' should not exist
		exists, e = GroupExists("notExistingGroup")
		expect(t, e.None(), "GroupExists('notExistingGroup') returned error: %v.", e)
		expect(t, !exists, "Expected GroupExists('notExistingGroup')==false, got true.")

	} else { // windows
		printf("-- Skipping TestGroupExists on Windows.\n")
	}
}
func TestUserExists(t *testing.T) {
	if runtime.GOOS != "windows" { // unix
		printf("-- Running TestUserExists on ...nix.\n")

		// User 'root' should exist
		exists, e := UserExists("root")
		expect(t, e.None(), "UserExists('root') returned error: %v.", e)
		expect(t, exists, "Expected UserExists('root')==true, got false.")

		// User 'notExistingUser' should not exist
		exists, e = UserExists("notExistingUser")
		expect(t, e.None(), "UserExists('notExistingUser') returned error: %v.", e)
		expect(t, !exists, "Expected UserExists('notExistingUser')==false, got true.")

	} else { // windows
		printf("-- Skipping TestUserExists on Windows.\n")
	}
}

// This test must only be run manually with 'sudo' privileges:
// sudo -s
// cd pkg/linuser
// /usr/local/go/bin/go test -run CreateGroup
func TestCreateGroup(t *testing.T) {
	if runtime.GOOS != "windows" { // unix
		printf("-- Running TestCreateGroup on ...nix.\n")

		// Create 'dummy_test_group'
		e := CreateGroup("dummy_test_group", true, 5005)
		expect(t, e.None(), "CreateGroup('dummy_test_group') returned error: %v.", e)

		// Creating 'sudo' group should return ec.AlreadyExists
		e = CreateGroup("sudo", true)
		expect(t, e.Code == ec.AlreadyExists, "CreateGroup('sudo') expected: ec.AlreadyExists, got: %v.", e)

	} else { // windows
		printf("-- Skipping TestCreateGroup on Windows.\n")
	}
}

// This test must only be run manually with 'sudo' privileges,
// immediately after TestCreateGroup:
// cd pkg/linuser
// /usr/local/go/bin/go test -run DeleteGroup
func TestDeleteGroup(t *testing.T) {
	if runtime.GOOS != "windows" { // unix
		printf("-- Running TestDeleteGroup on ...nix.\n")

		// Delete 'dummy_test_group'
		e := DeleteGroup("dummy_test_group", true)
		expect(t, e.None(), "DeleteGroup('dummy_test_group') returned error: %v.", e)

		// Trying to delete 'nonExistingGroup' should return ec.NotFound
		e = DeleteGroup("nonExistingGroup", true)
		expect(t, e.Code == ec.NotFound, "DeleteGroup('nonExistingGroup') expected: ec.NotFound, got: %v.", e)
	} else { // windows
		printf("-- Skipping TestDeleteGroup on Windows.\n")
	}
}

// This test must only be run manually with 'sudo' privileges:
// cd pkg/linuser
// /usr/local/go/bin/go test -run CreateUser
func TestCreateUser(t *testing.T) {
	if runtime.GOOS != "windows" { // unix
		printf("-- Running TestCreateUser on ...nix.\n")

		// Create 'testuser' in 'testuser'
		ud := UserDescriptor{
			UserName:  "testuser",
			UserId:    5006,
			GroupName: "testuser",
			GroupId:   7006,
			Home:      "/home/testuser",
			Shell:     "/bin/bash",
			Password:  "test",
		}
		e := CreateUser(ud, true)
		expect(t, e.None(), "CreateUser('testuser') returned error: %v.", e)

		// Creating same user again should return ec.AlreadyExists
		e = CreateUser(ud, true)
		expect(t, e.Code == ec.AlreadyExists, "CreateUser('testuser') expected: ec.AlreadyExists, got: %v.", e)

	} else { // windows
		printf("-- Skipping TestCreateUser on Windows.\n")
	}
}

// This test must only be run manually with 'sudo' privileges,
// immediately after TestCreateUser:
// cd pkg/linuser
// /usr/local/go/bin/go test -run DeleteUser
func TestDeleteUser(t *testing.T) {
	if runtime.GOOS != "windows" { // unix
		printf("-- Running TestDeleteUser on ...nix.\n")

		// Delete 'testuser'
		e := DeleteUser("testuser", true, true)
		expect(t, e.None(), "DeleteUser('testuser') returned error: %v.", e)

		// Trying to delete 'testuser' once more should return ec.NotFound
		e = DeleteUser("nonExistingUser", true, true)
		expect(t, e.Code == ec.NotFound, "DeleteUser('testuser') expected: ec.NotFound, got: %v.", e)
	} else { // windows
		printf("-- Skipping TestDeleteUser on Windows.\n")
	}
}

// This test must only be run manually with 'sudo' privileges:
// /usr/local/go/bin/go test -run TestAddUserToGroup
func TestAddUserToGroup(t *testing.T) {
	// Create two dummy users
	const userA = "dummyUser"
	const userB = "dummyUserB"

	userExists, e := UserExists(userA)
	expect(t, e.None(), "UserExists(%s): expected NoError, got %v.", userA, e)
	if !userExists {
		ud := UserDescriptor{
			UserName:  userA,
			UserId:    55550,
			GroupName: userA,
			GroupId:   55550,
		}
		e = CreateUser(ud, true)
		expect(t, e.None(), "CreateUser(%s): expected NoError, got %v.", userA, e)
	}

	userExists, e = UserExists(userB)
	expect(t, e.None(), "UserExists(%s): expected NoError, got %v.", userB, e)
	if !userExists {
		udb := UserDescriptor{
			UserName:  userB,
			UserId:    55551,
			GroupName: userB,
			GroupId:   55551,
		}
		e = CreateUser(udb, true)
		expect(t, e.None(), "CreateUser(%s): expected NoError, got %v.", userB, e)
	}

	// Adding existing user to existing group returns NoError
	e = AddUserToGroup(userA, userB, true)
	expect(t, e.None(), "AddUserToGroup(%s, %s): expected NoError, got %v.",
		userA, userB, e)

	// Adding not-existing user to existing group returns ec.NotFound
	e = AddUserToGroup("notExists", userB, true)
	expect(t, e.Code == ec.NotFound, "AddUserToGroup(notExists, %s): expected ec.NotFound, got %v.", userB, e)

	// Adding existing user to not-existing group returns ec.NotFound
	e = AddUserToGroup(userA, "groupNotExists", true)
	expect(t, e.Code == ec.NotFound, "AddUserToGroup(%s, groupNotExists): expected ec.NotFound, got %v.", userA, e)

	// IsUserInGroup tests
	// When user and group match, IsUserInGroup returns (true, NoError)
	inGroup, e := IsUserInGroup(userA, userA, true)
	expect(t, e.None(), "IsUserInGroup(%s, %s): expected NoError, got %v.",
		userA, userA, e)
	expect(t, inGroup, "IsUserInGroup(%s, %s): expected inGroup==true, got %v.", userA, userA, inGroup)

	// When user is in group, IsUserInGroup returns (true, NoError)
	inGroup, e = IsUserInGroup(userA, userB, true)
	expect(t, e.None(), "IsUserInGroup(%s, %s): expected NoError, got %v.", userA, userB, e)
	expect(t, inGroup, "IsUserInGroup(%s, %s): expected inGroup==true, got %v.", userA, userB, inGroup)

	// When user is not in group, IsUserInGroup returns (false, NoError)
	inGroup, e = IsUserInGroup(userB, userA, true)
	expect(t, e.None(), "IsUserInGroup(%s, %s): expected NoError, got %v.",
		userB, userA, e)
	expect(t, !inGroup, "IsUserInGroup(%s, %s): expected inGroup==false, got %v.",
		userB, userA, inGroup)

	// When user does not exist, IsUserInGroup returns (false, ec.NotFound)
	_, e = IsUserInGroup("notExists", userB, true)
	expect(t, e.Code == ec.NotFound, "IsUserInGroup(notExists, %s): expected ec.NotFound, got %v.", userB, e)

	// When group does not exist, IsUserInGroup returns (false, ec.NotFound)
	_, e = IsUserInGroup(userA, "notExists", true)
	expect(t, e.Code == ec.NotFound, "IsUserInGroup(%s, notExists): expected ec.NotFound, got %v.", userA, e)

	// RemoveUserFromGroup tests
	// When removing user from group it does not belong to, returns ec.NothingDone
	e = RemoveUserFromGroup(userB, userA)
	expect(t, e.Code == ec.NothingDone,
		"RemoveUserFromGroup(%s, %s): expected ec.NothingDone, got %v.", userB, userA, e)

	// When removing user from group it belongs to, returns NoError
	e = RemoveUserFromGroup(userA, userB)
	expect(t, e.None(), "RemoveUserFromGroup(%s, %s): expected NoError, got %v.",
		userA, userB, e)

	// Delete users
	e = DeleteUser(userA, true, true)
	expect(t, e.None(), "DeleteUser(%s, true, true): expected NoError, got %v.",
		userA, e)

	e = DeleteUser(userB, true, true)
	expect(t, e.None(), "DeleteUser(%s, true, true): expected NoError, got %v.",
		userB, e)
}
