package exit

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"testing"
)

var errUntyped = errors.New("error")

func wrapErr(err error) error { return fmt.Errorf("wrapped: %w", err) }

func TestExit(t *testing.T) {
	for _, testCase := range []struct {
		name string
		err  error
		code int
	}{
		{name: "no error", code: 0},
		{name: "untyped error", err: errUntyped, code: 1},
		{name: "exit error", err: Error(127, errUntyped), code: 127},
		{name: "wrapped exit error", err: wrapErr(Error(127, errUntyped)), code: 127},
		{name: "nil exit error", err: Error(127, nil), code: 0},
		{name: "flag help", err: flag.ErrHelp, code: 2},
		{name: "wrapped flag help", err: wrapErr(flag.ErrHelp), code: 2},
		{name: "exec.ExitError", err: execExitError(10), code: 10},
		{name: "wrapped exec.ExitError", err: wrapErr(execExitError(3)), code: 3},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			var got int

			osExit = func(code int) { got = code }
			defer func() { osExit = os.Exit }()

			Exit(testCase.err)

			if got != testCase.code {
				t.Errorf("got %d, want %d", got, testCase.code)
			}
		})
	}
}

// TestProcessExitCodeHelper is a helper to produce *exec.ExitError with a user
// defined exit code in unit tests.
func TestProcessExitCodeHelper(t *testing.T) {
	if os.Getenv("GO_PROCESS_EXIT_CODE_HELPER") != "1" {
		return
	}

	var code int = 1

	if len(os.Args) > 1 {
		// Starting with go1.16 calling os.Exit(0) from test cases causes tests
		// to fail so we ensure that the test helper only calls os.Exit with
		// non-zero exit codes. The last argument of our test helper is
		// considered to be the desired exit code for the test.
		arg, err := strconv.Atoi(os.Args[len(os.Args)-1])
		if err == nil && arg != 0 {
			code = arg
		}
	}

	os.Exit(code)
}

// execExitError produces an *exec.ExitError with the desired code. Must not be
// run in init funcs or to directly initialize package level variables or `go
// test` will hang and crash your machine.
func execExitError(code int) error {
	args := []string{"-test.run=TestProcessExitCodeHelper", "--", strconv.Itoa(code)}
	cmd := exec.Command(os.Args[0], args...)
	cmd.Env = []string{"GO_PROCESS_EXIT_CODE_HELPER=1"}
	return cmd.Run()
}