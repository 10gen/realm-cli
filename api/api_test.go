package api_test

import (
	"net/http"
	"testing"

	"github.com/10gen/realm-cli/api"
	"github.com/10gen/realm-cli/auth"
	"github.com/10gen/realm-cli/user"

	u "github.com/10gen/realm-cli/utils/test"
	gc "github.com/smartystreets/goconvey/convey"
)

func TestAuthClientRefreshAuth(t *testing.T) {
	t.Run("on success should return a new access token", func(t *testing.T) {
		client := u.NewMockClient([]*http.Response{
			{
				StatusCode: http.StatusCreated,
				Body: u.NewAuthResponseBody(auth.Response{
					AccessToken: "new.access.token",
				}),
			},
		})

		authClient := api.NewAuthClient(client, &user.User{AccessToken: "old.access.token", RefreshToken: "my.refresh.token"})

		authResponse, err := authClient.RefreshAuth()
		u.So(t, err, gc.ShouldBeNil)

		u.So(t, authResponse.AccessToken, gc.ShouldEqual, "new.access.token")
	})
}

func TestAuthClientExecuteRequest(t *testing.T) {
	t.Run("on unauthorized should refresh the token and make the request again", func(t *testing.T) {
		client := u.NewMockClient([]*http.Response{
			{
				StatusCode: http.StatusUnauthorized,
				Body:       u.NewAuthResponseBody(auth.Response{}),
			},
			{
				StatusCode: http.StatusCreated,
				Body: u.NewAuthResponseBody(auth.Response{
					AccessToken: "new.access.token",
				}),
			},
			{
				StatusCode: http.StatusOK,
				Body:       u.NewAuthResponseBody(auth.Response{}),
			},
		})

		authClient := api.NewAuthClient(client, &user.User{AccessToken: "old.access.token", RefreshToken: "my.refresh.token"})

		_, err := authClient.ExecuteRequest(http.MethodGet, "/somewhere", api.RequestOptions{})
		u.So(t, err, gc.ShouldBeNil)

		u.So(t, len(client.RequestData), gc.ShouldEqual, 3)
		u.So(t, client.RequestData[2].Method, gc.ShouldEqual, http.MethodGet)
		u.So(t, client.RequestData[2].Path, gc.ShouldEqual, "/somewhere")
		u.So(t, client.RequestData[2].Options.Header.Get("Authorization"), gc.ShouldEqual, "Bearer new.access.token")
	})
}
