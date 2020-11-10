package cli

import (
	"fmt"
)

// NewErr creates a new CLI error
func NewErr(message string) Err {
	return Err{message: message}
}

// NewErrw creates a new CLI error with the wrapped cause's details
// hidden from the resulting error message
func NewErrw(message string, err error) Err {
	return Err{message: message, cause: err}
}

// NewPrivilegedErr creates a new CLI error with the wrapped cause's details
// exposed in the resulting error message
func NewPrivilegedErr(message string, err error) PrivilegedErr {
	return PrivilegedErr{NewErrw(message, err)}
}

// WrappedErr is an error that can be unwrapped
type WrappedErr interface {
	Unwrap() error
}

// Err is a CLI error
type Err struct {
	WrappedErr
	message string
	cause   error
}

func (err Err) Error() string { return err.message }

// Unwrap unwraps the first non-CLI error as the root cause
func (err Err) Unwrap() error { return findRootCause(err.cause) }

func (err Err) String() string {
	if err.cause == nil {
		return err.message
	}

	var cause string
	switch c := err.cause.(type) {
	case Err:
		cause = c.String()
	default:
		cause = c.Error()
	}
	return fmt.Sprintf("%s: %s", err.message, cause)
}

// PrivilegedErr is a privileged CLI error
type PrivilegedErr struct {
	Err
}

func (err PrivilegedErr) Error() string {
	return fmt.Sprintf("%s: %s", err.message, err.Unwrap().Error())
}

func findRootCause(err error) error {
	switch e := err.(type) {
	case WrappedErr:
		return findRootCause(e.Unwrap())
	default:
		return e
	}
}
