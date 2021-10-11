// Package ec (Error Codes) defines error codes of the basic error group
// that are used in IGU library.
package ec

import (
	"fmt"

	"github.com/iotanbo/igu/pkg/ecdef"
)

const (
	// Zero error code, means no error.
	NoError ecdef.ErrCode = iota
	// Entity not found.
	NotFound
	// Operation lacked the necessary privileges to complete.
	PermissionDenied
	// Operation failed because a pipe was closed.
	BrokenPipe
	// Entity already exists, often a file.
	AlreadyExists
	// Entity already closed, often a file.
	AlreadyClosed
	// Operation needs to block to complete, but the blocking was requested to not occur.
	WouldBlock
	// Parameter was incorrect.
	InvalidInput
	// Invalid data source, e.g. file’s contents expected to be UTF-8 but is not.
	InvalidData
	// Operation timed out, causing it to be canceled.
	TimedOut
	// Call to write returned Ok(0), no more data can be written at the moment.
	WriteZero
	// Operation was interrupted (and typically can be retried).
	Interrupted
	// Other is the error code to be returned from a function
	// to signify an error that was not expected by normal flow control;
	// If the error is critical, use panic instead.
	Other
	// “end of file” was reached prematurely.
	UnexpectedEof
	// Unsupported on current platform.
	Unsupported
	// Assertion error.
	Assertion
	// Specified index is out of range.
	Index
	// Specified key not exists or is invalid.
	Key
	// Memory corruption or allocation failure.
	Memory
	// Entity not implemented.
	NotImplemented
	// Generic recursion error (not allowed, endless recursion etc.).
	Recursion
	// Generic syntax error.
	Syntax
	// Type is invalid.
	Type
	// Value is invalid.
	Value
	// Dummy error code used for testing and benchmarking.
	Dummy

	//* Other error codes that are staged for inclusion
	// An unexpected in current context error, user has a choice to either investigate
	// the cause and recover or terminate the program.
	// Unexpected

	// Windows-specific error when unix symlink is found on disk.
	SymlinksNotSupported

	// Process exited with error.
	ProcessExit

	// Result already achieved or job can't be done
	NothingDone
)

// ECToString() returns error code description as a string.
func ECToString(errCode ecdef.ErrCode) string {
	r := ""
	switch errCode {
	case NoError:
		r = "ec.NoError"
	case NotFound:
		r = "ec.NotFound"
	case PermissionDenied:
		r = "ec.PermissionDenied"
	// Operation failed because a pipe was closed.
	case BrokenPipe:
		r = "ec.BrokenPipe"
	// Entity already exists, often a file.
	case AlreadyExists:
		r = "ec.AlreadyExists"
		// Entity already closed, often a file.
	case AlreadyClosed:
		r = "ec.AlreadyClosed"
	// Operation needs to block to complete, but the blocking was requested to not occur.
	case WouldBlock:
		r = "ec.WouldBlock (operation needs to block to complete)"
	// Parameter was incorrect.
	case InvalidInput:
		r = "ec.InvalidInput (parameter was incorrect)"
	// Invalid data source, e.g. file’s contents expected to be UTF-8 but is not.
	case InvalidData:
		r = "ec.InvalidData (invalid source data)"
	// Operation timed out, causing it to be canceled.
	case TimedOut:
		r = "ec.TimedOut"
	// Call to write returned Ok(0), no more data can be written at the moment.
	case WriteZero:
		r = "ec.WriteZero (no more data can be written at the moment)"
	// Operation was interrupted (and typically can be retried).
	case Interrupted:
		r = "ec.Interrupted"
	case Other:
		r = "ec.Other (other error)"
	// “end of file” was reached prematurely.
	case UnexpectedEof:
		r = "ec.UnexpectedEof (EOF reached prematurely)"
	// Unsupported on current platform.
	case Unsupported:
		r = "ec.Unsupported"
	// Generic assertion error.
	case Assertion:
		r = "ec.Assertion (assertion error)"
	// Specified index is out of range.
	case Index:
		r = "ec.Index (index error)"
	// Specified key not exists or is invalid.
	case Key:
		r = "ec.Key (key error)"
	// Memory allocation failure or corruption.
	case Memory:
		r = "ec.Memory (memory error)"
	// Entity not implemented.
	case NotImplemented:
		r = "ec.NotImplemented"
	// Generic recursion error (not allowed, endless recursion etc.)
	case Recursion:
		r = "ec.Recursion (recursion error)"
	// Generic syntax error.
	case Syntax:
		r = "ec.Syntax (syntax error)"
	case Type:
		r = "ec.Type (type error)"
	case Value:
		r = "ec.Value (value error)"
	case Dummy:
		r = "ec.Dummy (dummy error for testing purposes)"
	case SymlinksNotSupported:
		r = "ec.SymlinksNotSupported (Windows?)"
	case ProcessExit:
		r = "ec.ProcessExit (process exited with error)"
	case NothingDone:
		r = "ec.DoneNothing (result already achieved or job can't be done)"
	default:
		r = fmt.Sprintf("unknown basic error code (%d)", errCode)
	}
	return r
}
