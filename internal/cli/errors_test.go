package cli

import (
	"errors"
	"testing"

	"github.com/10gen/realm-cli/internal/utils/test/so"
)

func TestErr(t *testing.T) {
	t.Run("NewErr should return a new error with no cause", func(t *testing.T) {
		err := NewErr("failed to time travel")

		so.So(t, err.Error(), so.ShouldEqual, "failed to time travel")
		so.So(t, err.Unwrap(), so.ShouldBeNil)
		so.So(t, err.String(), so.ShouldEqual, err.Error())
	})

	t.Run("NewErrw should return a new error with a wrapped cause and omit the details", func(t *testing.T) {
		cause := rootCause()
		err := NewErrw("failed to time travel", cause)

		so.So(t, err.Error(), so.ShouldEqual, "failed to time travel")
		so.So(t, err.Unwrap(), so.ShouldEqual, cause)
		so.So(t, err.String(), so.ShouldEqual, "failed to time travel: unable to reach precisely 88 MPH")

		t.Run("NewErrw should return an error with a deeply nested wrapped cause and omit the details", func(t *testing.T) {
			wrappedErr := NewErrw("failed to save yourself", err)

			so.So(t, wrappedErr.Error(), so.ShouldEqual, "failed to save yourself")
			so.So(t, wrappedErr.Unwrap(), so.ShouldEqual, cause)
			so.So(t, wrappedErr.String(), so.ShouldEqual, "failed to save yourself: failed to time travel: unable to reach precisely 88 MPH")
		})
	})

	t.Run("NewPrivilegedError should return a new error with a wrapped cause and include the details", func(t *testing.T) {
		cause := rootCause()
		err := NewPrivilegedErr("failed to time travel", cause)

		so.So(t, err.Error(), so.ShouldEqual, "failed to time travel: unable to reach precisely 88 MPH")
		so.So(t, err.Unwrap(), so.ShouldEqual, cause)
		so.So(t, err.String(), so.ShouldEqual, err.Error())

		t.Run("NewPrivilegedError should return an error with a deeply nested wrapped cause", func(t *testing.T) {
			wrappedErr := NewPrivilegedErr("failed to save yourself", err)

			so.So(t, wrappedErr.Error(), so.ShouldEqual, "failed to save yourself: unable to reach precisely 88 MPH")
			so.So(t, wrappedErr.Unwrap(), so.ShouldEqual, cause)
			so.So(t, wrappedErr.String(), so.ShouldEqual, "failed to save yourself: failed to time travel: unable to reach precisely 88 MPH")
		})
	})
}

func rootCause() error {
	return errors.New("unable to reach precisely 88 MPH")
}
