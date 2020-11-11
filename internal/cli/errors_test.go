package cli

import (
	"errors"
	"testing"

	u "github.com/10gen/realm-cli/internal/utils/test"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestErr(t *testing.T) {
	t.Run("New should return a new error with no cause", func(t *testing.T) {
		err := New("failed to time travel")

		u.MustMatch(t, cmp.Diff(err.Error(), "failed to time travel"))
		u.MustMatch(t, cmp.Diff(nil, err.Unwrap()))
		u.MustMatch(t, cmp.Diff(err.String(), err.Error()))
	})

	t.Run("NewWrapped should return a new error with a wrapped cause and omit the details", func(t *testing.T) {
		cause := rootCause()
		err := NewWrapped("failed to time travel", cause)

		u.MustMatch(t, cmp.Diff(err.Error(), "failed to time travel"))
		u.MustMatch(t, cmp.Diff(err.Unwrap(), cause, cmpopts.EquateErrors()))
		u.MustMatch(t, cmp.Diff(err.String(), "failed to time travel: unable to reach precisely 88 MPH"))

		t.Run("NewWrapped should return an error with a deeply nested wrapped cause and omit the details", func(t *testing.T) {
			wrappedErr := NewWrapped("failed to save yourself", err)

			u.MustMatch(t, cmp.Diff(wrappedErr.Error(), "failed to save yourself"))
			u.MustMatch(t, cmp.Diff(wrappedErr.Unwrap(), cause, cmpopts.EquateErrors()))
			u.MustMatch(t, cmp.Diff(wrappedErr.String(), "failed to save yourself: failed to time travel: unable to reach precisely 88 MPH"))
		})
	})

	t.Run("NewPrivilegedor should return a new error with a wrapped cause and include the details", func(t *testing.T) {
		cause := rootCause()
		err := NewPrivileged("failed to time travel", cause)

		u.MustMatch(t, cmp.Diff(err.Error(), "failed to time travel: unable to reach precisely 88 MPH"))
		u.MustMatch(t, cmp.Diff(err.Unwrap(), cause, cmpopts.EquateErrors()))
		u.MustMatch(t, cmp.Diff(err.String(), err.Error()))

		t.Run("NewPrivilegedor should return an error with a deeply nested wrapped cause", func(t *testing.T) {
			wrappedErr := NewPrivileged("failed to save yourself", err)

			u.MustMatch(t, cmp.Diff(wrappedErr.Error(), "failed to save yourself: unable to reach precisely 88 MPH"))
			u.MustMatch(t, cmp.Diff(wrappedErr.Unwrap(), cause, cmpopts.EquateErrors()))
			u.MustMatch(t, cmp.Diff(wrappedErr.String(), "failed to save yourself: failed to time travel: unable to reach precisely 88 MPH"))
		})
	})
}

func rootCause() error {
	return errors.New("unable to reach precisely 88 MPH")
}
