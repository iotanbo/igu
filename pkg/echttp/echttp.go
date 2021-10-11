// Package echttp defines status and error codes of the HTTP error group.
package echttp

import (
	"fmt"

	"github.com/iotanbo/igu/pkg/ecdef"
)

const (
	// This interim response indicates that everything so far is OK and that
	// the client should continue the request, or ignore the response
	// if the request is already finished.
	Continue_100 ecdef.ErrCode = ecdef.ErrCode(iota + ecdef.HTTP_RANGE_BEGIN)
)

// ECToString(...) returns a string describing an echttp error code.
func ECToString(errCode ecdef.ErrCode) string {
	r := ""
	switch errCode {
	case Continue_100:
		r = "100 continue"
	default:
		r = fmt.Sprintf("HTTP status code %d", errCode)
	}
	return r
}
