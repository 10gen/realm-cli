package utils_test

import (
	"testing"

	"github.com/10gen/realm-cli/utils"
	u "github.com/10gen/realm-cli/utils/test"
	gc "github.com/smartystreets/goconvey/convey"
)

func TestAppLoadFromDirectory(t *testing.T) {
	t.Run("should successfully load an app from the provided directory", func(t *testing.T) {
		app, err := utils.UnmarshalFromDir("../testdata/full_app")
		u.So(t, err, gc.ShouldBeNil)

		u.So(t, app["name"], gc.ShouldEqual, "full-app")
		u.So(t, app["secrets"], gc.ShouldNotBeEmpty)
		u.So(t, app["values"], gc.ShouldHaveLength, 2)
		for _, value := range app["values"].([]interface{}) {
			u.So(t, value.(map[string]interface{}), gc.ShouldNotBeEmpty)
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

		u.So(t, app["triggers"], gc.ShouldHaveLength, 2)
		for _, trgger := range app["triggers"].([]interface{}) {
			u.So(t, trgger.(map[string]interface{}), gc.ShouldNotBeEmpty)
		}

		gqlServices, ok := app["graphql"].(map[string]interface{})
		u.So(t, ok, gc.ShouldBeTrue)

		c, ok := gqlServices["config"].(map[string]interface{})
		u.So(t, ok, gc.ShouldBeTrue)

		useNatPluralization, ok := c["use_natural_pluralization"]
		u.So(t, ok, gc.ShouldBeTrue)
		u.So(t, useNatPluralization, gc.ShouldEqual, true)

		customResolvers, ok := gqlServices["custom_resolvers"].([]interface{})
		u.So(t, ok, gc.ShouldBeTrue)
		u.So(t, customResolvers, gc.ShouldHaveLength, 1)

		for _, customResolver := range customResolvers {
			customResolverMap := customResolver.(map[string]interface{})
			u.So(t, customResolverMap["field_name"], gc.ShouldNotBeEmpty)
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
