// Package ecfs defines error codes of the file system error group
// that are used in IGU library.
package ecfs

import (
	"fmt"

	"github.com/iotanbo/igu/pkg/ecdef"
)

const (
	// Error is a generic file system error that does not provide extra details.
	Error ecdef.ErrCode = ecdef.ErrCode(iota + ecdef.FS_RANGE_BEGIN)
	// Entity is not a file.
	NotAFile
	// Entity is not a directory.
	NotADir
	// Entity is not a symlink.
	NotASymlink
	// Entity is not a hardlink.
	NotAHardlink
	// The file is corrupt.
	FileCorrupt
	// The file is too large.
	FileTooLarge
	// The path is invalid.
	InvalidPath
)

// ECToString(...) returns a string describing an ecfs error code.
func ECToString(errCode ecdef.ErrCode) string {
	r := ""
	switch errCode {
	case Error:
		r = "file system error"
	case NotAFile:
		r = "not a file"
	case NotADir:
		r = "not a directory"
	case NotASymlink:
		r = "not a symlink"
	case NotAHardlink:
		r = "not a hardlink"
	case FileCorrupt:
		r = "file is corrupt"
	case FileTooLarge:
		r = "file is too large"
	case InvalidPath:
		r = "invalid path"
	default:
		r = fmt.Sprintf("unknown ecfs error code (%d)", errCode)
	}
	return r
}
