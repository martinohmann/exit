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

func wrapErr(err error) error {
	return fmt.Errorf("wrapped: %w", err)
}

func TestExit(t *testing.T) {
	for _, testCase := range []struct {
		name string
		err  error
		code int
	}{
		{name: "no error", code: CodeOK},
		{name: "untyped error", err: errUntyped, code: CodeErr},
		{name: "ExitError", err: Error(127, errUntyped), code: 127},
		{name: "wrapped ExitError", err: wrapErr(Error(127, errUntyped)), code: 127},
		{name: "nil error wrapped in ExitError", err: Error(127, nil), code: CodeOK},
		{name: "flag.ErrHelp", err: flag.ErrHelp, code: CodeHelpErr},
		{name: "wrapped flag.Help", err: wrapErr(flag.ErrHelp), code: CodeHelpErr},
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

func TestError(t *testing.T) {
	err := Error(CodeOSErr, nil)
	if err != nil {
		t.Errorf("got %#v, want nil", err)
	}

	err = Error(CodeOSErr, errors.New("the-error"))
	if exitErr, ok := err.(ExitError); !ok {
		t.Errorf("got %#v, want ExitError", err)
	} else if code := exitErr.ExitCode(); code != CodeOSErr {
		t.Errorf("got ExitError with code %d, want %d", code, CodeOSErr)
	} else if err.Error() != "the-error" {
		t.Errorf("got msg %q, want %q", err.Error(), "the-error")
	}

	origErr := &os.PathError{Err: errors.New("the-error")}

	err = Error(CodeErr, origErr)
	if wrappedErr := errors.Unwrap(err); wrappedErr != origErr {
		t.Errorf("errors.Unwrap(ExitError), got: %#v, want: %#v",
			wrappedErr, origErr)
	}
}

func TestErrorf(t *testing.T) {
	err := Errorf(CodeOSErr, "error: %s", "some-arg")
	if exitErr, ok := err.(ExitError); !ok {
		t.Errorf("got %#v, want ExitError", err)
	} else if code := exitErr.ExitCode(); code != CodeOSErr {
		t.Errorf("got ExitError with code %d, want %d", code, CodeOSErr)
	} else if err.Error() != "error: some-arg" {
		t.Errorf("got msg %q, want %q", err.Error(), "error: some-arg")
	}
}

func TestErrorp(t *testing.T) {
	var err error
	Errorp(CodeUsage, &err)

	if err != nil {
		t.Errorf("got %#v, want nil", err)
	}

	err = errors.New("error")

	Errorp(CodeUsage, &err)
	if err == nil {
		t.Error("got nil, want ExitError")
	} else if exitErr, ok := err.(ExitError); !ok {
		t.Errorf("got %#v, want ExitError", err)
	} else if code := exitErr.ExitCode(); code != CodeUsage {
		t.Errorf("got ExitError with code %d, want %d", code, CodeUsage)
	}
}

func TestSetErrorHandler(t *testing.T) {
	SetErrorHandler(func(err error) (code int, handled bool) {
		if err == nil {
			t.Error("error handler called with nil error")
		}

		var exitErr ExitError

		if errors.As(err, &exitErr) {
			// for testing purposes just add 1 to the existing exit code.
			return exitErr.ExitCode() + 1, true
		}

		return 0, false
	})
	defer SetErrorHandler(nil)

	for _, testCase := range []struct {
		name string
		err  error
		code int
	}{
		{name: "no error", code: CodeOK},
		{name: "untyped error", err: errUntyped, code: CodeErr},
		{name: "ExitError", err: Error(127, errUntyped), code: 128},
		{name: "wrapped ExitError", err: wrapErr(Error(127, errUntyped)), code: 128},
		{name: "nil error wrapped in ExitError", err: Error(127, nil), code: CodeOK},
		{name: "flag.ErrHelp", err: flag.ErrHelp, code: CodeHelpErr},
		{name: "wrapped flag.Help", err: wrapErr(flag.ErrHelp), code: CodeHelpErr},
		{name: "exec.ExitError", err: execExitError(10), code: 11},
		{name: "wrapped exec.ExitError", err: wrapErr(execExitError(3)), code: 4},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			if got := Code(testCase.err); got != testCase.code {
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
