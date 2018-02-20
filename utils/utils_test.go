package utils_test

import (
	"testing"

	"github.com/10gen/stitch-cli/utils"
	u "github.com/10gen/stitch-cli/utils/test"
	gc "github.com/smartystreets/goconvey/convey"
)

func TestAppLoadFromDirectory(t *testing.T) {
	t.Run("should successfully load an app from the provided directory", func(t *testing.T) {
		app, err := utils.UnmarshalFromDir("../testdata/full_app")
		u.So(t, err, gc.ShouldBeNil)

		u.So(t, app["name"], gc.ShouldEqual, "full-app")
		u.So(t, app["secrets"], gc.ShouldNotBeEmpty)
		u.So(t, app["values"], gc.ShouldHaveLength, 2)
		for _, svc := range app["values"].([]interface{}) {
			u.So(t, svc.(map[string]interface{}), gc.ShouldNotBeEmpty)
		}

		u.So(t, app["auth_providers"], gc.ShouldHaveLength, 2)
		for _, provider := range app["auth_providers"].([]interface{}) {
			u.So(t, provider.(map[string]interface{}), gc.ShouldNotBeEmpty)
		}

		u.So(t, app["functions"], gc.ShouldHaveLength, 2)
		for _, fn := range app["functions"].([]interface{}) {
			fnMap := fn.(map[string]interface{})
			u.So(t, fnMap["config"], gc.ShouldNotBeEmpty)
			u.So(t, fnMap["source"], gc.ShouldNotBeEmpty)
		}

		u.So(t, app["services"], gc.ShouldHaveLength, 3)
		for _, svc := range app["services"].([]interface{}) {
			svcMap := svc.(map[string]interface{})
			u.So(t, svcMap["config"], gc.ShouldNotBeEmpty)
			u.So(t, svcMap["rules"], gc.ShouldNotBeEmpty)
			u.So(t, svcMap["incoming_webhooks"], gc.ShouldNotBeEmpty)
		}
	})
}
