// Package ecdb defines error codes of the database error group
// that are used in IGU library.
package ecdb

import (
	"fmt"

	"github.com/iotanbo/igu/pkg/ecdef"
)

const (
	// Error is a generic database error that does not provide extra details.
	Error ecdef.ErrCode = ecdef.ErrCode(iota + ecdef.DB_RANGE_BEGIN)
)

// ECToString(...) returns a string describing an ECDB error code.
func ECToString(errCode ecdef.ErrCode) string {
	r := ""
	switch errCode {
	case Error:
		r = "database error"
	default:
		r = fmt.Sprintf("unknown ECDB error code (%d)", errCode)
	}
	return r
}
