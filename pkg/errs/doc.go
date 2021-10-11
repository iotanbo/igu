// Package errs implements simple yet effective error type Err
// that is compatible with standard error interface and
// is intended to be used with Iotanbo Go Utils (IGU) library.
// For brevity, this package can be imported the following way:
//	import (
//		//lint:ignore ST1001 - dot import to improve readability.
//		. "github.com/iotanbo/igu/pkg/errs"
//	)
// Import without dot can be used as well.
//
// Features:
//   * implements a light-weight Err object that is created, returned
//     from a function and checked in just 0.5... 2.5ns as benchmarks show;
//   * error codes are efficiently compared with each other;
//   * error codes are grouped into categories for better context help;
//   * error wrapping supported;
//   * standard error interface supported;
//   * support for custom, application-unique errors specified by user;
package errs
