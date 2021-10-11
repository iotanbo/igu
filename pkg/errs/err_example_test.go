// Demonstrates basic error handling with errs package
// and usage of app-specific error codes.
package errs_test

import (
	"fmt"

	"github.com/iotanbo/igu/pkg/ec"
	"github.com/iotanbo/igu/pkg/ecdef"

	. "github.com/iotanbo/igu/pkg/errs"
)

// Demonstrates how to define some app-specific error codes.
// This is usually done in a separate file so that
// other project files can import it.
const (
	ECAppMyError ecdef.ErrCode = ecdef.ErrCode(iota + ecdef.APP_RANGE_BEGIN)
	ECAppOther
)

func init() {
	// Demonstrates how to re-define function that converts
	// app-specific error codes to strings.
	AppECToString = func(code ecdef.ErrCode) string {
		switch code {
		case ECAppMyError:
			return "my app error"
		case ECAppOther:
			return "my other app error"
		default:
			return fmt.Sprintf("unknown app-specific error code (%d)", code)
		}
	}
}

// Demonstrates how to return a value and no error.
func doJobWithoutErrors() (int, Err) { return 42, NoError }

// Demonstrates how to return a value and an error.
func generateDummyError() (int, Err) {
	return 0, Err{Code: ec.Dummy, Msg: "for testing purposes"}
}

func Example() {
	// Recommendation: use name `e` for variables of type Err
	// and `err` for interface type `error` to distinguish them.
	val, e := doJobWithoutErrors()
	if e.Some() { // check if there is an error
		panic("error was not expected")
	}
	if val != 42 {
		panic(fmt.Sprintf("The answer expected to be 42, got %d!\n", val))
	}

	val, e = generateDummyError()
	if e.None() {
		panic("dummy error expected")
	}
	// Check error group if required.
	if !e.IsBasic() {
		panic("the error was expected to belong to basic group")
	}
	// Check that e is a dummy error.
	if e.Eq(ec.Dummy) {
		fmt.Println("Got dummy error (as expected).")
	}

	// Demonstrates how to create an app-specific error.
	appErr := Err{Code: ECAppMyError, Cause: e}
	fmt.Printf("%v", appErr)

	// Output:
	// Got dummy error (as expected).
	// my app error: dummy error for testing purposes
}
