package api_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/10gen/stitch-cli/api"

	u "github.com/10gen/stitch-cli/utils/test"
	gc "github.com/smartystreets/goconvey/convey"
)

func TestErrStitchResponse(t *testing.T) {
	t.Run("with a non-JSON response should return the original content", func(t *testing.T) {
		err := api.UnmarshalStitchError(&http.Response{
			Body: u.NewResponseBody(strings.NewReader("not-json")),
		})
		u.So(t, err, gc.ShouldBeError, "error: not-json")
	})

	t.Run("with an empty non-JSON response should respond with the status", func(t *testing.T) {
		err := api.UnmarshalStitchError(&http.Response{
			Status: "418 Toot toot",
			Body:   u.NewResponseBody(strings.NewReader("")),
		})
		u.So(t, err, gc.ShouldBeError, "error: 418 Toot toot")
	})

	t.Run("with a JSON response should decode the error content", func(t *testing.T) {
		err := api.UnmarshalStitchError(&http.Response{
			Body: u.NewResponseBody(strings.NewReader(`{ "error": "something went horribly, horribly wrong" }`)),
		})
		u.So(t, err, gc.ShouldBeError, "error: something went horribly, horribly wrong")
	})
}
