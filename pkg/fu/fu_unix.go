//go:build !windows
// +build !windows

package fu

import (
	//"fmt"
	//"io/ioutil"
	"os"
	//"path/filepath"
	"syscall"

	"github.com/iotanbo/igu/pkg/ec"
	//lint:ignore ST1001 - for concise error handling.
	. "github.com/iotanbo/igu/pkg/errs"
)

// GetItemType (unix version) returns the type of the file system item,
// one of [TYPE_FILE, TYPE_DIR, TYPE_SYMLINK, TYPE_HARDLINK, TYPE_NAMED_PIPE] and NoError if success.
// Otherwise returns TYPE_UNKNOWN and one of the following errors:
//	ec.NotFound // path does not exist or is invalid.
// 	ec.PermissionDenied;
// 	ec.TimedOut;
//	...or other less common errors.
// TODO: implement TYPE_NAMED_PIPE
func GetItemType(path string) (FsItemType, Err) {
	if path == "" {
		return TYPE_UNKNOWN, Err{Code: ec.NotFound}
	}
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return TYPE_UNKNOWN, Err{Code: ec.NotFound, Cause: err}
		} else if os.IsPermission(err) {
			return TYPE_UNKNOWN, Err{Code: ec.PermissionDenied, Cause: err}
		} else if os.IsTimeout(err) {
			return TYPE_UNKNOWN, Err{Code: ec.TimedOut, Cause: err}
		} else {
			return TYPE_UNKNOWN, Err{Code: ec.Other, Cause: err}
		}
	}
	if info.IsDir() {
		return TYPE_DIR, NoError
	} else {
		// https://github.com/docker/docker/blob/master/pkg/archive/archive_unix.go
		// in 'func setHeaderForSpecialDevice()'
		s, ok := info.Sys().(*syscall.Stat_t)
		if !ok {
			return TYPE_UNKNOWN,
				Err{Code: ec.Other,
					Msg: "failed to convert info.Sys() value to syscall.Stat_t"}
		}
		// True if the file is a symlink.
		if info.Mode()&os.ModeSymlink != 0 {
			return TYPE_SYMLINK, NoError
		}
		// The index number of this file's inode:
		//inode := uint64(s.Ino)
		// Total number of files/hardlinks connected to this file's inode:
		nlink := uint32(s.Nlink)
		if nlink > 1 {
			// There is no distinction between a file and a hardlink;
			// if two or more files point to same inode, they can be
			// considered hardlinks.
			return TYPE_HARDLINK, NoError
		}
		return TYPE_FILE, NoError
	}
}
