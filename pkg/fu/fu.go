package fu

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	otiai10 "github.com/otiai10/copy"

	"github.com/iotanbo/igu/pkg/ec"

	//lint:ignore ST1001 - for clear and concise error handling.
	. "github.com/iotanbo/igu/pkg/errs"
)

// External resources:
// Symlinks/hardlinks
// https://stackoverflow.com/a/31889712/3824328

// Copy tree:
// implementation by docker (linux (possible unix)-only)
// https://github.com/moby/moby/blob/master/daemon/graphdriver/copy/copy.go

// using stdlib, unix-only (+ docker implementation)
// https://stackoverflow.com/a/56314145/3824328

// + external lib ready to use, nice interface and a lot of options
// https://github.com/otiai10/copy

// FsItemType is used for enumeration
// of file system item types used in IGU library.
type FsItemType int32

// SymlinkCopyMode defines action to be performed when copying symlinks.
type SymlinkCopyMode int32

const (
	// SYMLINK_SHALLOW creates a new symlink pointing to dest.
	SYMLINK_SHALLOW SymlinkCopyMode = iota
	// SYMLINK_DEEP creates hard-copy of contents.
	SYMLINK_DEEP
	// SYMLINK_UNMODIFIED copies symlink as is, not modifying it.
	SYMLINK_UNMODIFIED
)

// DestOverwriteMode defines action to be performed
// when copying or moving files and destination already exists.
type DestOverwriteMode int32

const (
	// NO_OVERWRITE forbids overwriting destination.
	NO_OVERWRITE DestOverwriteMode = iota
	// MERGE allows merging source with destination.
	MERGE
	// OVERWRITE allows overwriting destination.
	OVERWRITE
)

// CopyOptions specifies options to be applied when copying or moving file or directory.
type CopyOptions struct {
	// SymlinkMode defines how to copy symlinks, must be one of
	// [SYMLINK_SHALLOW (default), SYMLINK_DEEP, SYMLINK_UNMODIFIED].
	SymlinkMode SymlinkCopyMode

	// OverwriteMode defines action to be done if destination already exists:
	// [NO_OVERWRITE (default), MERGE, OVERWRITE].
	OverwriteMode DestOverwriteMode

	// Skip defines if a file system item
	// has to be skipped while copying or moving.
	// If this function returns true, the item will be skipped.
	Skip func(src string) (bool, error)

	// AddPermission adds specified permission to each item.
	AddPermission os.FileMode

	// Sync file after copy.
	// Useful in case when file must be on the disk
	// (in case crash happens, for example),
	// at the expense of some performance penalty.
	Sync bool

	// Preserve the atime and the mtime of the entries.
	// On linux we can preserve only up to 1 millisecond accuracy.
	PreserveTimes bool

	// The byte size of the buffer to use for copying files.
	// If zero, the internal default buffer of 32KB is used.
	// See https://golang.org/pkg/io/#CopyBuffer for more information.
	CopyBufferSize uint
}

const (
	// File system item type is unknown.
	TYPE_UNKNOWN FsItemType = iota
	// Item is a regular file.
	TYPE_FILE
	// Item is a directory.
	TYPE_DIR
	// Item is a symlink (...nix-only type).
	TYPE_SYMLINK
	// Item is a hardlink (...nix-only type).
	// Hardlinks are two or more files that point to the same contents.
	// Changing the contents of one will change the others,
	// deleting/renaming one will not affect the others.
	// Only hardlinks to files are permitted.
	TYPE_HARDLINK
	// Item is a named pipe (...nix-only type).
	TYPE_NAMED_PIPE
)

func (t FsItemType) String() string {
	switch t {
	case TYPE_UNKNOWN:
		return "TYPE_UNKNOWN"
	case TYPE_FILE:
		return "TYPE_FILE"
	case TYPE_DIR:
		return "TYPE_DIR"
	case TYPE_SYMLINK:
		return "TYPE_SYMLINK"
	case TYPE_HARDLINK:
		return "TYPE_HARDLINK"
	case TYPE_NAMED_PIPE:
		return "TYPE_NAMED_PIPE"
	default:
		panic(fmt.Sprintf("unknown FsItemType: %d", t))
	}
}

// PathExists returns (true, FsItemType , NoError) if specified path exists,
// (false, TYPE_UNKNOWN, NoError) if it does not exist,
// and (false, TYPE_UNKNOWN, Err) if error occurred.
// Returned errors:
//	ec.PermissionDenied;
// 	ec.TimedOut;
//	...or other less common errors.
func PathExists(path string) (bool, FsItemType, Err) {
	t, e := GetItemType(path)
	if e.Some() {
		if e.Eq(ec.NotFound) {
			return false, TYPE_UNKNOWN, NoError
		}
		return false, TYPE_UNKNOWN, e
	}
	return true, t, NoError
}

// PathExistsTypeMatches returns (true, NoError) only if path exists
// and matches tExpected or one of the optional additionalTypes,
// (false, NoError) only if path does not exist.
// Otherwise returns errors:
//	ec.Type // path exists but doesn't match tExpected;
//	ec.PermissionDenied;
// 	ec.TimedOut;
//	...or other less common errors.
func PathExistsTypeMatches(path string, tExpected FsItemType,
	additionalTypes ...FsItemType) (bool, Err) {
	exists, t, e := PathExists(path)
	if e.Some() || !exists {
		return false, e
	}
	if t == tExpected {
		return true, NoError
	}
	for _, t_ex := range additionalTypes {
		if t == t_ex {
			return true, NoError
		}
	}
	return false, Err{Code: ec.Type, Msg: t.String()}
}

// FileExists returns (true, NoError) if specified path exists and is of type
// [TYPE_FILE, TYPE_SYMLINK, TYPE_HARDLINK, TYPE_NAMED_PIPE],
// or (false, NoError) if path doesn't exist.
// Otherwise:
//	ec.Type // path exists but is a directory, Msg field will contain actual type as string;
//	ec.PermissionDenied;
// 	ec.TimedOut;
//	...or other less common errors.
func FileExists(path string) (bool, Err) {
	return PathExistsTypeMatches(path, TYPE_FILE, TYPE_SYMLINK, TYPE_HARDLINK)
}

// DirExists returns (true, NoError) if specified path exists and is a dir,
// or (false, NoError) if path doesn't exist.
// Otherwise:
//	ec.Type // path exists but is not a dir, Msg field will contain actual type as string;
//	ec.PermissionDenied;
// 	ec.TimedOut;
//	...or other less common errors.
func DirExists(path string) (bool, Err) {
	return PathExistsTypeMatches(path, TYPE_DIR)
}

// SymlinkExists returns (true, NoError) only if specified path exists
// and is a symlink, or (false, NoError) if path doesn't exist.
// Returned errors:
//	ec.Type // path exists but is not a symlink, Msg field will contain actual type as string;
//	ec.PermissionDenied;
// 	ec.TimedOut;
//	...or other less common errors.
func SymlinkExists(path string) (bool, Err) {
	return PathExistsTypeMatches(path, TYPE_SYMLINK)
}

// HardlinkExists returns (true, NoError) only if specified path exists
// and is a hardlink, or (false, NoError) if path doesn't exist.
// Returned errors:
//	ec.Type // path exists but is not a hardlink, Msg field will contain actual type as string;
//	ec.PermissionDenied;
// 	ec.TimedOut;
//	...or other less common errors.
func HardlinkExists(path string) (bool, Err) {
	return PathExistsTypeMatches(path, TYPE_HARDLINK)
}

// NamedPipeExists returns (true, NoError) only if specified path exists
// and is a named pipe, or (false, NoError) if path doesn't exist.
// Returned errors:
//	ec.Type // path exists but is not a named pipe, Msg field will contain actual type as string;
//	ec.PermissionDenied;
// 	ec.TimedOut;
//	...or other less common errors.
func NamedPipeExists(path string) (bool, Err) {
	return PathExistsTypeMatches(path, TYPE_NAMED_PIPE)
}

// Translates CopyOptions into underlying implementation copy options
// which are somewhat cumbersome to be used directly.
func translateCopyOptions(o CopyOptions) otiai10.Options {
	var r = otiai10.Options{}

	// OnSymlink
	switch o.SymlinkMode {
	case SYMLINK_SHALLOW:
		r.OnSymlink = func(src string) otiai10.SymlinkAction {
			return otiai10.Shallow
		}
	case SYMLINK_DEEP:
		r.OnSymlink = func(src string) otiai10.SymlinkAction {
			return otiai10.Deep
		}
	case SYMLINK_UNMODIFIED:
		r.OnSymlink = func(src string) otiai10.SymlinkAction {
			return otiai10.Skip
		}
	}
	// OnDirExists
	switch o.OverwriteMode {
	case NO_OVERWRITE:
		r.OnDirExists = func(src, dest string) otiai10.DirExistsAction {
			return otiai10.Untouchable
		}
	case MERGE:
		r.OnDirExists = func(src, dest string) otiai10.DirExistsAction {
			return otiai10.Merge
		}
	case OVERWRITE:
		r.OnDirExists = func(src, dest string) otiai10.DirExistsAction {
			return otiai10.Replace
		}
	}

	// The rest of fields can be copied directly
	r.Skip = o.Skip
	if r.Skip == nil {
		r.Skip = func(string) (bool, error) {
			return false, nil // Don't skip
		}
	}
	r.AddPermission = o.AddPermission
	r.Sync = o.Sync
	r.PreserveTimes = o.PreserveTimes
	r.CopyBufferSize = o.CopyBufferSize
	return r
}

// Copy is a "swiss army knife" function that copies any kind of
// file system item (file, dir, symlink etc.) into dest using the options.
// The default options are: symlink shallow copy, no overwrite dest, no skip,
// no additional permissions, no sync, not preserve times,
// use default 32KB buffer. Returns NoError if success. Otherwise:
//	ec.NotFound // src not exists
//	ec.AlreadyExists // dest exists and OverwriteMode is NO_OVERWRITE
//	ec.Type // dest exists and has type different from src
//	ec.PermissionDenied
//	ec.TimedOut
//	...or other less common errors.
//
// Usage example:
//	// Copy using default options
//	e := Copy("/src", "/dest")
//	// Allow dest overwriting
//	e = Copy("/src", "/dest", CopyOptions{
//			OverwriteMode: OVERWRITE})
//	// Skip items that contain "temp" in their path
//	e = Copy("/src", "/dest", CopyOptions{
//			Skip: func(src string) (bool, error) {
//				if strings.Contains(src, "temp") {
//					return true, nil
//				}
//				return false, nil
//			},
//		})
func Copy(src, dest string, options ...CopyOptions) Err {
	var o CopyOptions
	if len(options) > 0 {
		o = options[0]
	} else {
		o = CopyOptions{}
	}
	// Check if dest exists and has same type as src
	srcExists, srcType, e := PathExists(src)
	if e.Some() {
		return e
	}
	if !srcExists {
		return Err{Code: ec.NotFound}
	}
	destExists, destType, e := PathExists(dest)
	if e.Some() {
		return e
	}
	if destExists {
		if srcType == destType {
			if o.OverwriteMode == NO_OVERWRITE {
				return Err{Code: ec.AlreadyExists, Msg: dest}
			}
		} else {
			// Dest already exists but its type doesn't match src
			return Err{Code: ec.Type}
		}
	}
	r := translateCopyOptions(o)
	err := otiai10.Copy(src, dest, r)
	if err != nil {
		return FromError(err)
	}
	return NoError
}

// CreateBinFile creates a binary file at specified path
// with specified contents.
// Contents is immediately flushed to permanent storage.
// The overwrite parameter allows file overwriting.
// Returns NoError if success. Otherwise:
//	ec.AlreadyExists // file already exists and overwrite is false
//	ec.Type // path already exists but is not a regular file
//	ec.PermissionDenied
//	ec.TimedOut
//	...or other less common errors.
func CreateBinFile(path string, contents []byte, overwrite bool) Err {
	// Check if path already exists
	exists, e := FileExists(path)
	if e.Some() {
		return e
	}
	if exists {
		if !overwrite {
			return Err{Code: ec.AlreadyExists, Msg: path}
		}
		// Delete old file
		err := os.Remove(path)
		if err != nil {
			return FromError(err)
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return FromError(err)
	}
	defer f.Close()
	_, err = f.Write(contents)
	if err != nil {
		return FromError(err)
	}
	err = f.Sync()
	if err != nil {
		return FromError(err)
	}
	return NoError
}

// CreateTextFile creates a text file at specified path
// with specified contents. If parent directory doesn't exist,
// it will be created.
// Contents is immediately flushed to permanent storage.
// The overwrite parameter allows file overwriting.
// Returns NoError if success. Otherwise:
//	ec.AlreadyExists // file already exists and overwrite is false
//	ec.Type // path already exists but is not a regular file
//	ec.PermissionDenied
//	ec.TimedOut
//	...or other less common errors.
func CreateTextFile(path, contents string, overwrite bool) Err {
	// Check if path already exists
	exists, e := FileExists(path)
	if e.Some() {
		return e
	}
	if exists {
		if !overwrite {
			return Err{Code: ec.AlreadyExists, Msg: path}
		}
		// Delete old file
		err := os.Remove(path)
		if err != nil {
			return FromError(err)
		}
	} else {
		// Create directory if not exists
		d := filepath.Dir(path)
		if err := os.MkdirAll(d, 0755); err != nil {
			e := FromError(err)
			e.Msg = "can't create directory: " + d
			return e
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return FromError(err)
	}
	defer f.Close()
	_, err = f.WriteString(contents)
	if err != nil {
		return FromError(err)
	}
	err = f.Sync()
	if err != nil {
		return FromError(err)
	}
	return NoError
}

// ReadBinFile reads the whole file into a slice of bytes.
// Returns (contents, NoError) if success.
// Otherwise:
//	ec.NotFound // path not exists
//	ec.Type // path already exists but is a directory
//	ec.PermissionDenied
//	ec.TimedOut
//	...or other less common errors.
func ReadBinFile(path string) ([]byte, Err) {
	var result []byte
	// Check if path already exists
	exists, t, e := PathExists(path)
	if e.Some() {
		return result, e
	}
	if !exists {
		return result, Err{Code: ec.NotFound, Msg: path}
	}
	if t == TYPE_DIR {
		return result, Err{Code: ec.Type, Msg: "TYPE_DIR"}
	}
	result, err := ioutil.ReadFile(path)
	if err != nil {
		return result, FromError(err)
	}
	return result, NoError
}

// ReadTextFile reads the whole text file into a string.
// Returns (contents, NoError) if success.
// Otherwise:
//	ec.NotFound // path not exists
//	ec.Type // path already exists but is a directory
//	ec.PermissionDenied
//	ec.TimedOut
//	...or other less common errors.
// TODO: check how it deals with encoding!
func ReadTextFile(path string) (string, Err) {
	result, e := ReadBinFile(path)
	if e.Some() {
		return "", e
	}
	return string(result), NoError
}

// ReadLines reads the text file into a slice of strings.
// Each string represents a separate line.
// Note: unix (\n) and windows (\r\n) line separators are supported
// on any platform, but old mac line separators (\r) are not supported
// and are treated as a single line.
// Returns (contents, NoError) if success.
// Otherwise:
//	ec.NotFound // path not exists
//	ec.Type // path already exists but is not a regular file
//	ec.Other with wrapped bufio.ErrTooLong // line exceeds 64KB
//	ec.PermissionDenied
//	ec.TimedOut
//	...or other less common errors.
// TODO: check how it deals with encoding!
func ReadLines(path string) ([]string, Err) {
	var result = []string{}
	// Check if path already exists
	exists, ty, e := PathExists(path)
	if e.Some() {
		return result, e
	}
	if !exists {
		return result, Err{Code: ec.NotFound, Msg: path}
	}
	if ty == TYPE_DIR {
		return result, Err{Code: ec.Type, Msg: "TYPE_DIR"}
	}

	file, err := os.Open(path)
	if err != nil {
		return result, FromError(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// See https://stackoverflow.com/a/16615559/3824328
	// it is possible to resize scanner's capacity for lines over 64K,
	// but it is not done here
	for scanner.Scan() {
		result = append(result, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return result, FromError(err)
	}
	return result, NoError
}
