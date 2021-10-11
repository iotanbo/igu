// Package ecnet defines error codes of the network error group
// that are used in IGU library.
package ecnet

import (
	"fmt"

	"github.com/iotanbo/igu/pkg/ecdef"
)

const (
	// Error is a generic network-related error that does not provide extra details.
	Error ecdef.ErrCode = ecdef.ErrCode(iota + ecdef.NET_RANGE_BEGIN)
	// Connection was refused by the remote server.
	ConnectionRefused
	// Connection was reset by the remote server.
	ConnectionReset
	// Connection was aborted (terminated) by the remote server.
	ConnectionAborted
	// Network operation failed because it was not connected yet.
	NotConnected
	// Socket address is already in use elsewhere.
	AddrInUse
	// Nonexistent network interface was requested or the address is not local.
	AddrNotAvailable
)

// ECToString(...) returns a string describing an ecnet error code.
func ECToString(errCode ecdef.ErrCode) string {
	r := ""
	switch errCode {
	case Error:
		r = "network error"
		// Connection was refused by the remote server.
	case ConnectionRefused:
		r = "connection refused"
	// Connection was reset by the remote server.
	case ConnectionReset:
		r = "connection reset"
	// Connection was aborted (terminated) by the remote server.
	case ConnectionAborted:
		r = "connection aborted"
	// Network operation failed because it was not connected yet.
	case NotConnected:
		r = "not connected"
	// Socket address is already in use elsewhere.
	case AddrInUse:
		r = "address in use"
	// Nonexistent network interface was requested or the address is not local.
	case AddrNotAvailable:
		r = "address not available"
	default:
		r = fmt.Sprintf("unknown ecnet error code (%d)", errCode)
	}
	return r
}
