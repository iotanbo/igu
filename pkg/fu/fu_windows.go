package fu

// This file contains code exclusive for windows.

import (
	"os"

	"github.com/iotanbo/igu/pkg/ec"

	//lint:ignore ST1001 - for clear and concise error handling.
	. "github.com/iotanbo/igu/pkg/errs"
)

// GetItemType() - windows version - returns the type of the file system item,
// one of [TYPE_FILE, TYPE_DIR, TYPE_SYMLINK] and NoError if success.
// Otherwise returns TYPE_UNKNOWN and one of the following errors:
//	ec.NotFound if path does not exist or is invalid;
// 	ec.PermissionDenied;
// 	ec.TimedOut;
// 	ec.Other if other error(s) occurred.
func GetItemType(path string) (FsItemType, Err) {
	if path == "" {
		return TYPE_UNKNOWN, Err{Code: ec.NotFound}
	}
	info, err := os.Stat(path)
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
		if info.Mode()&os.ModeSymlink != 0 {
			return TYPE_SYMLINK, NoError
		}
		return TYPE_FILE, NoError
	}
}
