package utils_test

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/10gen/stitch-cli/utils"
	u "github.com/10gen/stitch-cli/utils/test"
	gc "github.com/smartystreets/goconvey/convey"
)

func TestCheckForNewCLIVersion(t *testing.T) {
	for _, tc := range []struct {
		description      string
		expectedResponse string
		responseBody     string
	}{
		{
			description: "should place nicely with non-semver versions",
			responseBody: `
				{
					"version":"20180702",
					"info":{
						"default": {
							"url":"http://whatever.com/test"
						}
					}
				}
				`,
		},
		{
			description: "should return nothing if the current version is not greater than the current version",
			responseBody: `
				{
					"version":"1.0.0",
					"info":{
						"default": {
							"url":"http://whatever.com/test"
						}
					}
				}
				`,
		},
		{
			description:      "should return a helpful message if there is a newer version",
			expectedResponse: "New version (v1.1.0) of CLI available at http://whatever.com/test",
			responseBody: `
				{
					"version":"1.1.0",
					"info":{
						"macos-amd64": {
							"url":"http://whatever.com/test"
						}
					}
				}
				`,
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			utils.CLIOSArch = "macos-amd64"
			client := &testClient{
				response: &http.Response{
					StatusCode: http.StatusOK,
					Body:       &testBody{strings.NewReader(tc.responseBody)},
				},
			}
			u.So(t, utils.CheckForNewCLIVersion(client), gc.ShouldEqual, tc.expectedResponse)
		})
	}
}

type testClient struct {
	response *http.Response
}

func (tc *testClient) Get(url string) (*http.Response, error) {
	return tc.response, nil
}

type testBody struct {
	io.Reader
}

func (tb *testBody) Close() error {
	return nil
}
