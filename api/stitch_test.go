package api_test

import (
	"strings"
	"testing"

	"github.com/10gen/stitch-cli/api"

	u "github.com/10gen/stitch-cli/utils/test"
	gc "github.com/smartystreets/goconvey/convey"
)

func TestErrStitchResponse(t *testing.T) {
	t.Run("with a non-JSON response should return the original content", func(t *testing.T) {
		err := api.UnmarshalReader(strings.NewReader("not-json"))
		u.So(t, err, gc.ShouldBeError, "error: not-json")
	})

	t.Run("with a JSON response should decode the error content", func(t *testing.T) {
		err := api.UnmarshalReader(strings.NewReader(`{ "error": "something went horribly, horribly wrong" }`))
		u.So(t, err, gc.ShouldBeError, "error: something went horribly, horribly wrong")
	})
}
