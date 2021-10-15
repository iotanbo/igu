package sh

// https://github.com/cosiner/argv

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/iotanbo/igu/pkg/ec"
	//lint:ignore ST1001 - for concise error handling.
	. "github.com/iotanbo/igu/pkg/errs"
)

// ExecuteCmd executes a shell command and returns stdout and stderr outputs
// as strings. Optional timeout in milliseconds can be specified.
//
// Returned errors:
//	ec.NoError // completed successfully
//	ec.ProcessExit // sub-process exited with non-zero error code
//	ec.TimedOut // timeout occurred
//	ec.PermissionDenied
// Other errors may be returned for other situations.
//
// Usage example:
//	cmd := "sleep"
//	args := []string{"5"}
//	stdout, stderr, e := ExecuteCmd(cmd, args, 200)
func ExecuteCmd(cmd string, args []string, timeout ...int64) (string, string, Err) {
	// Based on https://stackoverflow.com/a/43246464/3824328
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if len(timeout) > 0 {
		// time.Duration units are nanoseconds, so multiply by 1000000
		calcTimeout := time.Duration(timeout[0] * 1000000)
		ctx, cancel := context.WithTimeout(context.Background(),
			calcTimeout)
		defer cancel()

		c := exec.CommandContext(ctx, cmd, args...)
		c.Stdout = &stdout
		c.Stderr = &stderr
		if err := c.Run(); err != nil {
			errMsg := fmt.Sprintf("%v", err)
			if strings.Contains(errMsg, "signal: killed") {
				return stdout.String(), stderr.String(),
					Err{Code: ec.TimedOut, Cause: err}
			}
			return stdout.String(), stderr.String(),
				FromError(err)
		}
		return stdout.String(), stderr.String(), NoError
	} else {
		c := exec.Command(cmd, args...)
		c.Stdout = &stdout
		c.Stderr = &stderr
		if err := c.Run(); err != nil {
			return stdout.String(), stderr.String(), FromError(err)
		}
		return stdout.String(), stderr.String(), NoError
	}
}

// ExecuteLine executes command that is a single line, e.g. `cp file1 file2`.
// Returns stdout and stderr outputs as strings.
// Optional timeout in milliseconds can be specified.
// This function requires `bash` to be present in the system.
//
// Returned errors:
//	ec.NoError // completed successfully
//	ec.Syntax // cmd contains syntax errors
//	ec.ProcessExit // sub-process exited with non-zero error code
//	ec.TimedOut // timeout occurred
//	ec.PermissionDenied
// Other errors may be returned for other situations.
//
// Usage example:
//	cmdLine := "sleep 5"
//	stdout, stderr, e := ExecuteLine(cmdLine, 200)
func ExecuteLine(cmdLine string, timeout ...int64) (string, string, Err) {
	cmd := "bash"
	args := []string{
		"-c",
		cmdLine,
	}
	return ExecuteCmd(cmd, args, timeout...)

	// Old implementation
	// argSets, err := argv.Argv(cmdLine, func(backquoted string) (string, error) {
	// 	return backquoted, nil
	// }, nil)
	// if err != nil {
	// 	return "", "", Err{Code: ec.Syntax}
	// }

	// cmdArray := argSets[0]
	// if len(cmdArray) == 0 {
	// 	return "", "", NoError
	// }

	// cmd := cmdArray[0]
	// args := []string{}
	// if len(cmdArray) >= 2 {
	// 	args = cmdArray[1:]
	// }
	// return ExecuteCmd(cmd, args, timeout...)
}
