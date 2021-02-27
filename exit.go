// Package exit provides a helper to choose a reasonable exit code based on
// errors.
//
// Errors can be passed to Exit to exit the program with a meaningful exit code
// based on the kind and type of error:
//
//   exit.Exit(err)
//
// Alternatively the exit code for an error can be computed via Status and used
// later:
//
//   code := exit.Status(err)
//
//   os.Exit(code)
//
// For more control over the exit code, errors can be wrapped in an ExitError
// which passes on the desired exit code:
//
//   func foo() error {
//     err := someOperation()
//
//     // Wraps err in an ExitError with code 127 if err is non-nil.
//     return exit.Error(127, err)
//   }
//
//   func main() {
//     exit.Exit(foo())  // May produce exit code 0 on success, 127 on error.
//   }
//
// Errors can also be wrapped in a defer statement via Errorp:
//
//   func foo() (err error) {
//     defer exit.Errorp(127, &err)
//     err = someOperation()
//     if err != nil {
//       return
//     }
//
//     err = someOtherOperation()
//     return
//   }
//
package exit

import (
	"errors"
	"flag"
	"os"

	"github.com/spf13/pflag"
)

// ExitError is an error that can signal the desired exit code. It is
// implemented by the standard library's *exec.ExitError for example.
type ExitError interface {
	error
	ExitCode() int
}

// Error wraps err with an ExitError that returns given code. If err is nil it
// is returned as is.
func Error(code int, err error) error {
	if err == nil {
		return nil
	}

	return &exitError{err, code}
}

// Errorp wraps the pointed-to error with an ExitError and sets err to the new
// value. If the value of err a nil it is not wrapped. Can be used in defer
// statements to wrap errors before returning them.
//
// Example:
//
//   defer exit.Errorp(127, &err)
//
// See Error for more information.
func Errorp(code int, err *error) {
	*err = Error(code, *err)
}

type exitError struct {
	error
	code int
}

func (e *exitError) ExitCode() int { return e.code }

// Status picks a suitable exit code for err. If err is nil the returned code
// is 0. Otherwise it attempts to provide a meaningful exit code for err.
//
// Uses the standard library's errors.Is and errors.As functions to also
// inspect wrapped errors.
//
// If an error implements ExitError (e.g. *exec.ExitError) the value
// returned by err.ExitCode() will be returned.
//
// If err contains flag.ErrHelp or github.com/spf13/pflag.ErrHelp, the exit
// code will be 2.
//
// All other errors produce exit code 1.
func Status(err error) int {
	var exitErr ExitError

	switch {
	case err == nil:
		return 0
	case errors.Is(err, flag.ErrHelp) || errors.Is(err, pflag.ErrHelp):
		return 2
	case errors.As(err, &exitErr):
		return exitErr.ExitCode()
	default:
		return 1
	}
}

// Overridden in tests.
var osExit = os.Exit

// Exit is a convenience alternative for os.Exit. Calls os.Exit with the exit
// code obtained from err. If err is nil this is equivalent to os.Exit(0). See
// documentation of Status for possible exit codes.
func Exit(err error) {
	osExit(Status(err))
}
