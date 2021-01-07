package realm

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestServerError(t *testing.T) {
	var jsonContentTypeHeader http.Header = map[string][]string{"Content-Type": []string{"application/json"}}

	t.Run("Should unmarshal a non-json response successfully", func(t *testing.T) {
		err := parseResponseError(&http.Response{
			Body: ioutil.NopCloser(strings.NewReader("something bad happened")),
		})
		assert.Equal(t, "", err.Error())
	})

	t.Run("Should create error from an empty response with its status", func(t *testing.T) {
		err := parseResponseError(&http.Response{
			Status: "something bad happened",
			Body:   ioutil.NopCloser(strings.NewReader("")),
		})
		assert.Equal(t, "something bad happened", err.Error())
	})

	t.Run("Should unmarshal a server error payload without an error code successfully", func(t *testing.T) {
		err := parseResponseError(&http.Response{
			Body:   ioutil.NopCloser(strings.NewReader(`{"error": "something bad happened"}`)),
			Header: jsonContentTypeHeader,
		})
		assert.Equal(t, ServerError{Message: "something bad happened"}, err)
	})

	t.Run("Should unmarshal a server error payload with an error code successfully", func(t *testing.T) {
		err := parseResponseError(&http.Response{
			Body:   ioutil.NopCloser(strings.NewReader(`{"error": "something bad happened","error_code": "AnErrorCode"}`)),
			Header: jsonContentTypeHeader,
		})
		assert.Equal(t, ServerError{Code: "AnErrorCode", Message: "something bad happened"}, err)
	})
}
