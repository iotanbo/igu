// Package ecsys defines error codes of the system error group
// that are used in IGU library.
package ecsys

import (
	"fmt"

	"github.com/iotanbo/igu/pkg/ecdef"
)

const (
	// Error is a generic system error that does not provide extra details.
	Error ecdef.ErrCode = ecdef.ErrCode(iota + ecdef.SYS_RANGE_BEGIN)
	// System exit.
	SystemExit
	// Interrupted from keyboard.
	KeyboardInterrupt
	// Generic arithmetic error.
)

// ECToString(...) returns a string describing an ecsys error code.
func ECToString(errCode ecdef.ErrCode) string {
	r := ""
	switch errCode {
	case Error:
		r = "system error"
	case SystemExit:
		r = "system exit"
	case KeyboardInterrupt:
		r = "keyboard interrupt"
	default:
		r = fmt.Sprintf("unknown ecsys error code (%d)", errCode)
	}
	return r
}
