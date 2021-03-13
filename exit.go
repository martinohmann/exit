// Package exit wrap errors with exit code information and conversely produces
// meaningful exit codes based on errors.
//
// Errors can be passed to Exit to exit the program with a meaningful exit code
// based on the kind and type of error:
//
//   exit.Exit(err)
//
// Alternatively the exit code for an error can be computed via Code and used
// later:
//
//   code := exit.Code(err)
//
//   os.Exit(code)
//
// For more control over the exit code, errors can be wrapped in an ExitError
// which passes on the desired exit code:
//
//   func foo() error {
//     err := someOperation()
//
//     // Wraps err in an ExitError with code 74 if err is non-nil.
//     return exit.Error(exit.CodeIOErr, err)
//   }
//
//   func main() {
//     // Produces exit code 0 on success, 74 on error.
//     exit.Exit(foo())
//   }
//
// Directly construct an error of type ExitError with a specific exit code or
// wrap an existing error with more context:
//
//   err1 := exit.Errorf(code, "failed to do the thing")
//   err2 := exit.Errorf(code, "failed to do the other thing: %w", err)
//
// Errors can also be wrapped in a defer statement via Errorp:
//
//   func foo() (err error) {
//     defer exit.Errorp(exit.CodeNoPerm, &err)
//     err = someOperation()
//     if err != nil {
//       return
//     }
//
//     err = someOtherOperation()
//     return
//   }
//
//   func main() {
//     // Produces exit code 0 on success, 77 on error.
//     exit.Exit(foo())
//   }
//
// Mapping of errors to exit code can also be controlled by setting a custom
// error handler via SetErrorHandler:
//
//   func main() {
//     exit.SetErrorHandler(func(err error) (code int, handled bool) {
//       var customErr CustomError
//       if errors.As(err, &customErr) {
//         return exit.CodeUsage, true
//       }
//       return 0, false
//     })
//
//     // Produces exit code 64 if someOperation returns CustomError.
//     exit.Exit(someOperation())
//   }
//
package exit

import (
	"errors"
	"flag"
	"fmt"
	"os"
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

// Errorf creates a new error and wraps it into an ExitError with given exit
// code. Format and args are used to build the wrapped error via fmt.Errorf.
func Errorf(code int, format string, args ...interface{}) error {
	return &exitError{fmt.Errorf(format, args...), code}
}

// Errorp wraps the pointed-to error with an ExitError and sets err to the new
// value. If the value of err a nil it is not wrapped. Can be used in defer
// statements to wrap errors before returning them.
//
// Example:
//
//   defer exit.Errorp(exit.CodeOSErr, &err)
//
// See Error for more information.
func Errorp(code int, err *error) {
	*err = Error(code, *err)
}

type exitError struct {
	error
	code int
}

func (e *exitError) Unwrap() error { return e.error }

func (e *exitError) ExitCode() int { return e.code }

// Code picks a suitable exit code for err. If err is nil the returned code
// is 0. Otherwise it attempts to provide a meaningful exit code for err.
//
// If a custom error handler func was set via SetErrorHandler and it is
// non-nil, this func is executed first to determine a suitable exit code if
// err is non-nil. Otherwise it proceeds to determine the exit code by the
// builtin rules below.
//
// Uses the standard library's errors.Is and errors.As functions to also
// inspect wrapped errors.
//
// If an error implements ExitError (e.g. *exec.ExitError) the value
// returned by err.ExitCode() will be returned.
//
// If err contains flag.ErrHelp the exit code will be 2.
//
// All other errors produce exit code 1.
func Code(err error) int {
	if err != nil && errorHandlerFn != nil {
		if code, handled := errorHandlerFn(err); handled {
			return code
		}
	}

	var exitErr ExitError

	switch {
	case err == nil:
		return CodeOK
	case errors.Is(err, flag.ErrHelp):
		return CodeHelpErr
	case errors.As(err, &exitErr):
		return exitErr.ExitCode()
	default:
		return CodeErr
	}
}

var (
	// Overridden in tests.
	osExit = os.Exit

	errorHandlerFn ErrorHandlerFunc
)

// ErrorHandlerFunc may provide an exit code for err. If it determined a
// suitable exit code for err it should signal this by setting the second
// return value to true.
type ErrorHandlerFunc func(err error) (code int, handled bool)

// SetErrorHandler sets a custom error handler. The error handler is called
// when Code or Exit are invoked with a non-nil error. If fn does not signal
// that it handled an error by returning true as its second return value the
// exit code is determined using the builtin rules.
//
// Calling SetErrorHandler is not goroutine-safe. Should be called early in
// main.
//
// See Code for more information.
func SetErrorHandler(fn ErrorHandlerFunc) {
	errorHandlerFn = fn
}

// Exit is a convenience alternative for os.Exit. Calls os.Exit with the exit
// code obtained from err. If err is nil this is equivalent to os.Exit(0).
//
// See Code for possible exit codes.
func Exit(err error) {
	osExit(Code(err))
}
