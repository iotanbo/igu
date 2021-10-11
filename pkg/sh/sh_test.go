package sh

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/iotanbo/igu/pkg/ec"
	"github.com/stretchr/testify/require"
)

// Type aliases to improve readability
var printf = fmt.Printf
var expect = require.True

func TestExecuteCmd(t *testing.T) {
	if runtime.GOOS != "windows" { // unix
		// Test without timeout
		printf("-- Running tests on ...nix.\n")
		cmd := "ls"
		args := []string{"-la"}
		stdout, stderr, e := ExecuteCmd(cmd, args)
		expect(t, e.None(), "* expected e.None(), got: "+
			"'%s', '%s', '%v'\n", stdout, stderr, e)
		expect(t, len(stderr) == 0)

		// Test with timeout
		cmd = "sleep"
		args = []string{"5"}
		stdout, stderr, e = ExecuteCmd(cmd, args, 200)
		expect(t, e.Code == ec.TimedOut, "* expected ec.TimedOut, got: "+
			"'%s', '%s', '%v'\n", stdout, stderr, e)

		// TODO: Test with timeout that is not exceeded
	} else { // windows
		printf("-- Running tests on Windows.\n")
	}

}

func TestExecuteLine(t *testing.T) {
	if runtime.GOOS != "windows" { // unix
		// Test without timeout
		printf("-- Running tests on ...nix.\n")
		cmdLine := " ls -la"
		stdout, stderr, e := ExecuteLine(cmdLine)
		expect(t, e.None(), "* expected e.None(), got: "+
			"'%s', '%s', '%v'\n", stdout, stderr, e)
		expect(t, len(stderr) == 0)

		// Test with timeout
		cmdLine = "sleep 5"
		stdout, stderr, e = ExecuteLine(cmdLine, 200)
		expect(t, e.Code == ec.TimedOut, "* expected ec.TimedOut, got: "+
			"'%s', '%s', '%v'\n", stdout, stderr, e)

		// TODO: Test with timeout that is not exceeded
	} else { // windows
		printf("-- Running tests on Windows.\n")
	}

}
