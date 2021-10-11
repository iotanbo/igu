// Package ecmath defines error codes of the math error group
// that are used in IGU library.
package ecmath

import (
	"fmt"

	"github.com/iotanbo/igu/pkg/ecdef"
)

const (
	// Error is a generic math-related error that does not provide extra details.
	Error ecdef.ErrCode = ecdef.ErrCode(iota + ecdef.MATH_RANGE_BEGIN)
	// Floating point operation error.
	FloatingPoint
	// Overflow error.
	Overflow
	// Division by zero.
	ZeroDivision
)

// ECToString(...) returns a string describing an ecmath error code.
func ECToString(errCode ecdef.ErrCode) string {
	r := ""
	switch errCode {
	case Error:
		r = "math error"
		// Generic floating point error.
	case FloatingPoint:
		r = "floating point error"
	// Generic overflow error.
	case Overflow:
		r = "overflow"
	// Division by zero.
	case ZeroDivision:
		r = "division by zero"
	default:
		r = fmt.Sprintf("unknown ecmath error code (%d)", errCode)
	}
	return r
}
