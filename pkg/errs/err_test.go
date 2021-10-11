package errs

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"testing"

	"github.com/iotanbo/igu/pkg/ec"

	"github.com/iotanbo/igu/pkg/ecdef"
	"github.com/stretchr/testify/require"
)

func ExampleAsErr() {
	var err error = Err{Code: ec.Dummy}
	if e, ok := AsErr(err); ok {
		fmt.Println(e)
	}
	err = fmt.Errorf("not of type Err")
	if e, ok := AsErr(err); ok {
		fmt.Println(e)
	} else {
		fmt.Println("Second err is not of type Err.")
	}
	// Output:
	// dummy error
	// Second err is not of type Err.
}

func TestErr_Is(t *testing.T) {
	e := NoError
	// Check that modifying the fields of `e` does not affect the NoError variable
	e.Msg = "test message"
	require.True(t, NoError.Msg == "", "NoError.Msg changed after assignment")

	require.True(t, errors.Is(e, NoError), "errors.Is(e, NoError) must return true")

	require.False(t, errors.Is(e, fmt.Errorf("")),
		"comparing with type other than Err must return false")
	require.False(t, errors.Is(e, nil),
		"comparing with nil must return false")

	// Check wrapped errors
	e1 := Err{Code: ec.Dummy, Cause: fmt.Errorf("cause0")}
	require.True(t, errors.Is(e1, Err{Code: ec.Dummy}),
		"errors.Is(e, ec.Dummy) must return true")

	// If method Is() not implemented for a concret error type,
	// errors.Is() returns false.
	require.False(t, errors.Is(fmt.Errorf("cause0"), fmt.Errorf("cause0")),
		"errors.Is(fmt.Errorf(\"cause0\"), fmt.Errorf(\"cause0\")) expected to return false")

	// Two wrapped errors of different types
	e2 := Err{Code: ec.Dummy, Cause: Err{Code: ec.NotFound, Cause: fs.ErrNotExist}}
	fmt.Printf("e2: %v\n", e2)
	require.True(t, errors.Is(e2, Err{Code: ec.Dummy}),
		"errors.Is(e2, Err{Code: ec.Dummy}) must return true")
	require.True(t, errors.Is(e2, Err{Code: ec.NotFound}),
		"errors.Is(e2, Err{Code: ec.NotFound}) must return true")
	require.True(t, errors.Is(e2, fs.ErrNotExist),
		"errors.Is(e2, fs.ErrNotExist) must return true")
	// Same as previous, but is called as a method
	require.True(t, e2.Is(fs.ErrNotExist),
		"e2.Is(fs.ErrNotExist) must return true")

}

//* Test that error code ranges do not overlap and have capacity >= 99 values.
func TestECRanges(t *testing.T) {

	actions := map[string]func(e *Err) bool{
		"basic": func(e *Err) bool { return e.IsBasic() },
		"auth":  func(e *Err) bool { return e.IsAuth() },
		"db":    func(e *Err) bool { return e.IsDB() },
		"fs":    func(e *Err) bool { return e.IsFS() },
		"http":  func(e *Err) bool { return e.IsHTTP() },
		"math":  func(e *Err) bool { return e.IsMath() },
		"net":   func(e *Err) bool { return e.IsNet() },
		"sys":   func(e *Err) bool { return e.IsSys() },
		"app":   func(e *Err) bool { return e.IsApp() },
	}
	var testData = []struct {
		rangeName  string
		start, end Err
	}{
		{
			"basic",
			Err{Code: ecdef.BASIC_RANGE_BEGIN},
			Err{Code: ecdef.BASIC_RANGE_END},
		},
		{
			"auth",
			Err{Code: ecdef.AUTH_RANGE_BEGIN},
			Err{Code: ecdef.AUTH_RANGE_END},
		},
		{
			"db",
			Err{Code: ecdef.DB_RANGE_BEGIN},
			Err{Code: ecdef.DB_RANGE_END},
		},
		{
			"fs",
			Err{Code: ecdef.FS_RANGE_BEGIN},
			Err{Code: ecdef.FS_RANGE_END},
		},
		{
			"http",
			Err{Code: ecdef.HTTP_RANGE_BEGIN},
			Err{Code: ecdef.HTTP_RANGE_END},
		},
		{
			"math",
			Err{Code: ecdef.MATH_RANGE_BEGIN},
			Err{Code: ecdef.MATH_RANGE_END},
		},
		{
			"net",
			Err{Code: ecdef.NET_RANGE_BEGIN},
			Err{Code: ecdef.NET_RANGE_END},
		},
		{
			"sys",
			Err{Code: ecdef.SYS_RANGE_BEGIN},
			Err{Code: ecdef.SYS_RANGE_END},
		},
		{
			"app",
			Err{Code: ecdef.APP_RANGE_BEGIN},
			Err{Code: ecdef.APP_RANGE_END},
		},
	}

	for _, sample := range testData {
		// Check that range has enough capacity
		diff := int32(sample.end.Code) - int32(sample.start.Code)
		require.True(t, diff >= 99, "range '%s' has capacity less than 99 (%d)",
			sample.rangeName, diff)
		// Check that ranges do not overlap
		for name, action := range actions {
			resultStart := action(&sample.start)
			resultEnd := action(&sample.end)
			if name == sample.rangeName {
				require.True(t, resultStart, "range start for '%s' must be true", name)
				require.True(t, resultEnd, "range end for '%s' must be true", name)
			} else {
				require.False(t, resultStart,
					"range '%s' start overlaps with range '%s'", name, sample.rangeName)
				require.False(t, resultStart,
					"range '%s' end overlaps with range '%s'", name, sample.rangeName)
			}
		}
	}

}

func TestFromError(t *testing.T) {
	// Create file without privileges must return ec.PermissionDenied
	_, err := os.Create("/dummy.txt")
	e := FromError(err)
	require.True(t, e.Eq(ec.PermissionDenied),
		"os.Create('/dummy.txt') did NOT return ec.PermissionDenied but %v", e)

	// Open non-existing file must return ec.NotFound
	_, err = os.Open("/dummy.txt")
	e = FromError(err)
	require.True(t, e.Eq(ec.NotFound),
		"os.Create('/dummy.txt') did NOT return ec.NotFound but %v", e)

	// Attempt to overwrite existing file must return ec.AlreadyExists
	// TODO

	//fmt.Printf("error Open(/dummy.txt): %v\n", err)
}

// --- BENCHMARKS
var dummyErr = Err{Code: ec.Dummy}

// Benchmarking shows that returning plain Err is cheap,
// there is no memory allocations and it takes about 0.27..2.4ns;
//! Returning pointer (*Err) instead of Err increases execution time
//! to 48 ns and makes one allocation.
//? compiler cheats (over-optimizes)?
func generateDummyError() Err {
	return dummyErr // Err{Code: ec.Dummy}
}

func BenchmarkPlainErrorCreation(b *testing.B) {
	//var sum int32 = 1
	for i := 0; i < b.N; i++ {
		e := generateDummyError()
		if e.Code != ec.Dummy {
			log.Fatalf("Expected ec.Dummy, received %v", e)
		}
		//sum += int32(e.Code << (i % 32))
	}
	// log.Printf("BenchmarkPlainErrorCreation dummy sum: %d\n", sum)
}

// Benchmarking shows that returning plain Err with a short message
// takes same time of about 2.5 ns, no allocations.
func generateErrorWithMessage() Err {
	return Err{Code: ec.Dummy, Msg: "Short dummy description"}
}

// 2.3ns/op, 0 alloc
func BenchmarkErrorWithMessageCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		e := generateErrorWithMessage()
		if e.Code != ec.Dummy {
			log.Fatalf("Expected ec.Dummy, received %v", e)
		}
	}
}

var dummyCause1 = Err{Code: ec.Dummy, Msg: "dummyCause1"}
var dummyCause2 = Err{Code: ec.Dummy, Msg: "dummyCause2", Cause: dummyCause1}
var dummyCause3 = Err{Code: ec.Dummy, Msg: "dummyCause3", Cause: dummyCause2}
var dummyCause4 = Err{Code: ec.Dummy, Msg: "dummyCause4", Cause: dummyCause3}

// 59ns/op, 1 alloc
func generateErrorWithMessageAndCause() Err {
	return Err{Code: ec.Dummy, Msg: "Short dummy description",
		Cause: dummyCause1}
}

// 59ns/op, 1 alloc
func BenchmarkErrorWithMessageAndCauseCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		e := generateErrorWithMessageAndCause()
		if e.Code != ec.Dummy {
			log.Fatalf("Expected ec.Dummy, received %v", e)
		}
	}
}

func generateErrorWithMessageAndCauseChain() Err {
	return Err{Code: ec.Dummy, Msg: "Short dummy description",
		Cause: dummyCause4}
}

// 59ns/op, 1 alloc
func BenchmarkErrorWithMessageAndCauseChainCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		e := generateErrorWithMessageAndCauseChain()
		if e.Code != ec.Dummy {
			log.Fatalf("Expected ec.Dummy, received %v", e)
		}
	}
}

// 14ns/op, 0 alloc
func BenchmarkAsErr(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var err error = Err{Code: ec.Dummy, Msg: "Dummy message"}
		e, ok := AsErr(err)
		if !ok {
			log.Fatalf("BenchmarkAsErr: AsErr() to succeed, received %T", e)
		}
		if e.Code != ec.Dummy {
			log.Fatalf("BenchmarkAsErr: subErr.Code != ec.Dummy, received %v", e.Code)
		}
	}
}

// 0.57ns/op, 0 alloc
func BenchmarkErrNone(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if dummyErr.None() {
			log.Fatalln("Expected _dummyErr.None() return false")
		}
	}
}

// 0.57ns/op, 0 alloc
func BenchmarkErrIsBasic(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if !dummyErr.IsBasic() {
			log.Fatalln("Expected _dummyErr.IsBasic() return true")
		}
	}
}
