// Package ecdef (Error Code Definitions) defines ErrCode type along with error code range constants used in IGU library, these definitions are separated into a standalone package
// to eliminate cyclic imports.
package ecdef

// Error Code type
type ErrCode int32

// Constants that define error code ranges

// The beginning of basic error code range.
const BASIC_RANGE_BEGIN ErrCode = 0

// The end of basic error code range.
const BASIC_RANGE_END ErrCode = 99

// The beginning of HTTP error code range.
const HTTP_RANGE_BEGIN ErrCode = 100

// The end of HTTP error code range.
const HTTP_RANGE_END ErrCode = 599

// The beginning of file system error code range.
const FS_RANGE_BEGIN ErrCode = 1000000

// The end of file system error code range.
const FS_RANGE_END ErrCode = 1000299

// The beginning of authentication/authorization error code range.
const AUTH_RANGE_BEGIN ErrCode = 1000300

// The end of authentication/authorization error code range.
const AUTH_RANGE_END ErrCode = 1000599

// The beginning of network error code range.
const NET_RANGE_BEGIN ErrCode = 1000600

// The end of network error code range.
const NET_RANGE_END ErrCode = 1000899

// The beginning of database-related error code range.
const DB_RANGE_BEGIN ErrCode = 1000900

// The end of database-related error code range.
const DB_RANGE_END ErrCode = 1001199

// The beginning of database-related error code range.
const MATH_RANGE_BEGIN ErrCode = 1001200

// The end of database-related error code range.
const MATH_RANGE_END ErrCode = 1001499

// The beginning of system error code range.
const SYS_RANGE_BEGIN ErrCode = 1001500

// The end of system error code range.
const SYS_RANGE_END ErrCode = 1001799

// The beginning of app-specific error code range.
const APP_RANGE_BEGIN ErrCode = 1000000000

// The end of app-specific error code range.
const APP_RANGE_END ErrCode = 1000099999
