package realm_test

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	u "github.com/10gen/realm-cli/internal/utils/test"

	"github.com/google/go-cmp/cmp"
)

func TestServerError(t *testing.T) {
	t.Run("Should unmarshal a non-json response successfully", func(t *testing.T) {
		err := realm.UnmarshalServerError(&http.Response{
			Body: ioutil.NopCloser(strings.NewReader("something bad happened")),
		})
		u.MustNotBeNil(t, err)
		u.MustMatch(t, cmp.Diff("something bad happened", err.Error()))
	})

	t.Run("Should unmarshal an empty response with its status", func(t *testing.T) {
		err := realm.UnmarshalServerError(&http.Response{
			Status: "something bad happened",
			Body:   ioutil.NopCloser(strings.NewReader("")),
		})
		u.MustNotBeNil(t, err)
		u.MustMatch(t, cmp.Diff("something bad happened", err.Error()))
	})

	t.Run("Should unmarshal a server error payload without an error code successfully", func(t *testing.T) {
		err := realm.UnmarshalServerError(&http.Response{
			Body: ioutil.NopCloser(strings.NewReader(`{"error": "something bad happened"}`)),
		})
		u.MustNotBeNil(t, err)
		u.MustMatch(t, cmp.Diff("something bad happened", err.Error()))
	})

	t.Run("Should unmarshal a server error payload with an error code successfully", func(t *testing.T) {
		err := realm.UnmarshalServerError(&http.Response{
			Body: ioutil.NopCloser(strings.NewReader(`{"error": "something bad happened","error_code": "AnErrorCode"}`)),
		})
		u.MustNotBeNil(t, err)
		u.MustMatch(t, cmp.Diff("something bad happened", err.Error()))
	})
}
