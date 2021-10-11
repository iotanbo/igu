// Package ecauth defines error codes of the authentication / authorization error group
// that are used in IGU library.
package ecauth

import (
	"fmt"

	"github.com/iotanbo/igu/pkg/ecdef"
)

const (
	// Authentication/Authorization failure that does not provide extra details.
	Failed ecdef.ErrCode = ecdef.ErrCode(iota + ecdef.AUTH_RANGE_BEGIN)

	// Bad credentials.
	Credentials

	// Proposed authentication method not supported.
	UnsupportedMethod
)

// ECToString(...) returns a string describing an ecauth error code.
func ECToString(errCode ecdef.ErrCode) string {
	r := ""
	switch errCode {
	case Failed:
		r = "authentication failed"
	case Credentials:
		r = "bad credentials"
	case UnsupportedMethod:
		r = "unsupported authentication method"
	default:
		r = fmt.Sprintf("unknown ecauth error code (%d)", errCode)
	}
	return r
}
