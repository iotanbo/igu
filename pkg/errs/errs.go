package errs

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"reflect"

	"github.com/iotanbo/igu/pkg/ec"
	"github.com/iotanbo/igu/pkg/ecauth"
	"github.com/iotanbo/igu/pkg/ecdb"
	"github.com/iotanbo/igu/pkg/ecdef"
	"github.com/iotanbo/igu/pkg/ecfs"
	"github.com/iotanbo/igu/pkg/echttp"
	"github.com/iotanbo/igu/pkg/ecmath"
	"github.com/iotanbo/igu/pkg/ecnet"
	"github.com/iotanbo/igu/pkg/ecsys"
)

// NoError is same as Err{}. It's a convenience global variable that makes it easy to return
// no-error values from functions and improve readability. Must not be modified by user.
var NoError = Err{}

// AppECToString is a function object that converts custom app-specific
// error codes defined by user to strings.
// By default, this is just a stub that prints error codes.
// If there is a need to use custom error codes within an application
// and get more informative messages,
// assign a custom function that converts your custom error codes to strings
// to this variable at program start (e.g. in the main init() function).
// Example is provided in documentation.
var AppECToString = func(code ecdef.ErrCode) string {
	return fmt.Sprintf("undefined app-specific error (%d)", code)
}

// Err is a simple yet effective error type that implements error interface
// and is used by Iotanbo Go Utils (IGU) library.
// For better performance it is recommended to return it by value.
// Usage examples:
//	e0 := Err{} // no error
//	e1 := Err{Code: ec.Dummy}
//	e2 := Err{Code: ec.Other, Msg: "other error from dummy", Cause: e1}
type Err struct {
	// Error code
	Code ecdef.ErrCode
	// Optional error message.
	Msg string
	// The error that caused this error (optional, typically nil).
	Cause error
}

// IsBasic() checks that e.Code belongs to the basic group of error codes
// defined in the ec package.
func (e *Err) IsBasic() bool {
	return e.Code >= ecdef.BASIC_RANGE_BEGIN && e.Code <= ecdef.BASIC_RANGE_END
}

// IsAuth() checks that e.Code belongs to the auth group of error codes
// defined in the ecauth package.
func (e *Err) IsAuth() bool {
	return e.Code >= ecdef.AUTH_RANGE_BEGIN && e.Code <= ecdef.AUTH_RANGE_END
}

// IsDB() checks that e.Code belongs to the database group of error codes
// defined in the ecdb package.
func (e *Err) IsDB() bool {
	return e.Code >= ecdef.DB_RANGE_BEGIN && e.Code <= ecdef.DB_RANGE_END
}

// IsMath() checks that e.Code belongs to the math group of error codes
// defined in the ecmath package.
func (e *Err) IsMath() bool {
	return e.Code >= ecdef.MATH_RANGE_BEGIN && e.Code <= ecdef.MATH_RANGE_END
}

// IsNet() checks that e.Code belongs to the network group of error codes
// defined in the ecnet package.
func (e *Err) IsNet() bool {
	return e.Code >= ecdef.NET_RANGE_BEGIN && e.Code <= ecdef.NET_RANGE_END
}

// IsSys() checks that e.Code belongs to the system group of error codes
// defined in the ecsys package.
func (e *Err) IsSys() bool {
	return e.Code >= ecdef.SYS_RANGE_BEGIN && e.Code <= ecdef.SYS_RANGE_END
}

// IsHttp() checks that e.Code belongs to the http group of error codes
// defined in the echttp package.
func (e *Err) IsHTTP() bool {
	return e.Code >= ecdef.HTTP_RANGE_BEGIN && e.Code <= ecdef.HTTP_RANGE_END
}

// IsFS() checks that e.Code belongs to the file system group of error codes
// defined in the ecfs package.
func (e *Err) IsFS() bool {
	return e.Code >= ecdef.FS_RANGE_BEGIN && e.Code <= ecdef.FS_RANGE_END
}

// IsApp() checks that e.Code belongs to the app-specific group of error codes
// optionally defined by user.
func (e *Err) IsApp() bool {
	return e.Code >= ecdef.APP_RANGE_BEGIN && e.Code <= ecdef.APP_RANGE_END
}

// errToString(...) returns (only) current error's message as string.
func errToString(e *Err) string {
	var r string
	if e.IsBasic() {
		r = ec.ECToString(e.Code)
	} else if e.IsFS() {
		r = ecfs.ECToString(e.Code)
	} else if e.IsSys() {
		r = ecsys.ECToString(e.Code)
	} else if e.IsMath() {
		r = ecmath.ECToString(e.Code)
	} else if e.IsNet() {
		r = ecnet.ECToString(e.Code)
	} else if e.IsApp() {
		r = AppECToString(e.Code)
	} else if e.IsAuth() {
		r = ecauth.ECToString(e.Code)
	} else if e.IsDB() {
		r = ecdb.ECToString(e.Code)
	} else if e.IsHTTP() {
		r = echttp.ECToString(e.Code)
	} else {
		r = fmt.Sprintf("unknown error code (%d)", e.Code)
	}
	if len(e.Msg) == 0 {
		return r
	}
	return r + fmt.Sprintf(" %s", e.Msg)
}

// Error() is the implementation of error interface for type Err,
// the resulting string also includes the descriptions of all wrapped errors (if any).
func (e Err) Error() string {
	r := errToString(&e)
	cause := e.Cause
	// Include messages from the wrapped errors
	for {
		if cause == nil {
			break
		}
		casted, ok := AsErr(cause)
		if ok {
			r += fmt.Sprintf(": %s", errToString(&casted))
		} else {
			r += fmt.Sprintf(": %s", cause)
		}
		cause = errors.Unwrap(cause)
	}
	return r
}

// Some() returns true in case of error, i.e. when e.Code has value other than ec.NoError.
func (e *Err) Some() bool { return e.Code != ec.NoError }

// None() returns true if e.Code is equal to ec.NoError.
func (e *Err) None() bool { return e.Code == ec.NoError }

// Eq() returns true if error code of this Err object equals to specified value.
// e.Msg and e.Cause fields are ignored.
func (e *Err) Eq(errCode ecdef.ErrCode) bool { return e.Code == errCode }

// Unwrap() returns the wrapped error if any or nil otherwise.
// Intended for convenient traversing the error chain.
func (e Err) Unwrap() error { return e.Cause }

func compareErrorInterfaces(current error, target error) bool {
	targetType := reflect.TypeOf(target)
	targetValue := reflect.ValueOf(target)

	currentType := reflect.TypeOf(current)
	if currentType == targetType {
		// Try comparing directly
		if currentType.Comparable() {
			currentValue := reflect.ValueOf(current)
			return currentValue == targetValue
			// fmt.Printf("ret false - types are same but values differ (%s, %s)\n",
			// 	currentValue, targetValue)
			// return false
		}
		// fmt.Printf("ret false - types same but  not comparable (%s, %s)\n",
		// 	currentType, targetType)
		return false
	}
	// fmt.Printf("ret false - types are different (%s, %s)\n",
	// 	currentType, targetType)
	return false
}

// Is() compares e and all its wrapped errors to target.
// It returns true if e (or any of its wrapped errors) equals to target.
// It can be used directly as a method or indirectly by errors.Is() function.
// Note that there is a much more efficient Eq() method to compare two Err object's codes.
//	e := Err{Code: ec.Dummy, Cause: Err{Code: ec.NotFound, Cause: fs.ErrNotExist}}
//	fmt.Println(errors.Is(e, fs.ErrNotExists)) // true
func (e Err) Is(target error) bool {
	if target == nil {
		return false
	}
	var targetAsErr Err
	targetOfTypeErr := false
	// Special case if target is of type Err
	if castedTarget, ok := target.(Err); ok {
		targetAsErr = castedTarget
		targetOfTypeErr = true
		if e.Eq(targetAsErr.Code) {
			return true
		}
	}
	var current error = e.Cause
	for {
		if current == nil {
			return false
		}
		if targetOfTypeErr {
			// check if current is also of type Err
			if castedCurrent, ok := current.(Err); ok {
				if castedCurrent.Eq(targetAsErr.Code) {
					return true
				}
			}
		} else {
			if compareErrorInterfaces(current, target) {
				return true
			}
		}
		// Try to Unwrap() current cause
		i, ok := current.(interface{ Unwrap() error })
		if ok { // the error has Unwrap() method
			current = i.Unwrap()
			// fmt.Printf("successfully unwrapped current to (%s)\n",
			// 	current)
		} else { // can't unwrap any more
			// fmt.Println("can't unwrap any more, returning false")
			return false
		}
	}
}

// Is() old implementation that doesn't check wrapped errors.
// func (e Err) Is(target error) bool {
// 	// Check that target is of type Err
// 	targetErr, ok := target.(Err)
// 	if ok {
// 		// Check that Code fields match
// 		return targetErr.Code == e.Code
// 	}
// 	return false
// }

// AsErr() tries to cast e to type Err.
// If succeeds, it returns e as type Err and true,
// otherwise a default-constructed Err object and false.
// This function needs zero memory allocations,
// benchmarking shows execution time of 15ns on a vingate i7.
func AsErr(e error) (Err, bool) {
	if e == nil {
		return Err{}, false
	}
	castedToErr, ok := e.(Err)
	if ok {
		return castedToErr, true
	}
	return castedToErr, false
}

// FromError converts standard error interface type to Err
// by searching for best possible match.
// If there is no match, the error code will be set to ec.Other.
func FromError(e error) Err {
	var code = ec.Other

	// Check if it is PathError
	if err, ok := e.(*os.PathError); ok {
		//fmt.Printf("--- debug FromStdError() - PathError: Op: '%s', Path: '%s',"+
		//	"Err: '%v'\n", err.Op, err.Path, err.Err)

		// Timeout is treaded specially in PathError
		if err.Timeout() {
			code = ec.TimedOut
		} else {
			es := err.Err.Error()
			switch es {
			// This seems to be MACOS-specific error
			case "read-only file system":
				code = ec.PermissionDenied
			// Following are defined in package oserror
			case "invalid argument":
				code = ec.InvalidInput
			case "permission denied":
				code = ec.PermissionDenied
			case "file already exists":
				code = ec.AlreadyExists
			case "file does not exist":
				code = ec.NotFound
			case "no such file or directory":
				code = ec.NotFound
			case "file already closed":
				code = ec.AlreadyClosed
			// Windows10-specific PermissionDenied error message
			case "Access is denied.":
				code = ec.PermissionDenied
			// Windows10-specific error when it encounters a unix symlink
			case "The name of the file cannot be resolved by the system.":
				code = ec.SymlinksNotSupported
			// Windows10-specific NotFound error message
			case "The system cannot find the file specified.":
				code = ec.NotFound
			}

		}
	} else if _, ok := e.(*os.LinkError); ok {
		// TODO: remove this debug
		//fmt.Printf("--- debug FromStdError() - LinkError: Op: '%s', Old: '%s', New: '%s', Err: '%v'\n", le.Op, le.Old, le.New, le.Error())
	} else if _, ok := e.(*os.SyscallError); ok {
		// TODO: remove this debug
		//fmt.Printf("--- debug FromStdError() - SyscallError: Syscall: '%s', Err: '%v'\n", se.Syscall, se.Err)
		// Timeout is treaded specially in PathError
		if err.Timeout() {
			code = ec.TimedOut
		} else {
			// es := err.Err.Error()
			// switch es {
			// }
			code = ec.Other
		}
	} else if _, ok := e.(*exec.ExitError); ok {
		code = ec.ProcessExit
	}

	return Err{Code: code, Cause: e}
}
