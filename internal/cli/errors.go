package cli

import (
	"errors"
	"fmt"
	"log"
)

// HandleErr handles any CLI error
// TODO(REALMC-7340): build this out
func HandleErr(err error) {
	if err == nil {
		return
	}
	log.Fatal(err)
}

// New creates a new CLI error
func New(message string) Err {
	return Err{message: message}
}

// NewWrapped creates a new CLI error with the wrapped cause's details
// hidden from the resulting error message
func NewWrapped(message string, err error) Err {
	return Err{message: message, cause: err}
}

// NewPrivileged creates a new CLI error with the wrapped cause's details
// exposed in the resulting error message
func NewPrivileged(message string, err error) PrivilegedErr {
	return PrivilegedErr{NewWrapped(message, err)}
}

// Err is a CLI error
type Err struct {
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
	if cause := errors.Unwrap(err); cause != nil {
		return findRootCause(cause)
	}
	return err
}
