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

		valuesList, ok := app["values"].([]interface{})
		u.So(t, valuesList, gc.ShouldHaveLength, 2)
		u.So(t, ok, gc.ShouldBeTrue)
		for _, value := range valuesList {
			valueMap, ok := value.(map[string]interface{})
			u.So(t, ok, gc.ShouldBeTrue)
			u.So(t, valueMap, gc.ShouldNotBeEmpty)
		}

		authProvidersList, ok := app["auth_providers"].([]interface{})
		u.So(t, ok, gc.ShouldBeTrue)
		u.So(t, authProvidersList, gc.ShouldHaveLength, 2)
		for _, provider := range authProvidersList {
			providerMap, ok := provider.(map[string]interface{})
			u.So(t, ok, gc.ShouldBeTrue)
			u.So(t, providerMap, gc.ShouldNotBeEmpty)
		}

		functionsList, ok := app["functions"].([]interface{})
		u.So(t, ok, gc.ShouldBeTrue)
		u.So(t, functionsList, gc.ShouldHaveLength, 2)
		for _, fn := range functionsList {
			fnMap, ok := fn.(map[string]interface{})
			u.So(t, ok, gc.ShouldBeTrue)
			u.So(t, fnMap["config"], gc.ShouldNotBeEmpty)
			u.So(t, fnMap["source"], gc.ShouldNotBeEmpty)
		}

		triggersList, ok := app["triggers"].([]interface{})
		u.So(t, ok, gc.ShouldBeTrue)
		u.So(t, triggersList, gc.ShouldHaveLength, 2)
		for _, triggr := range triggersList {
			triggerMap, ok := triggr.(map[string]interface{})
			u.So(t, ok, gc.ShouldBeTrue)
			u.So(t, triggerMap, gc.ShouldNotBeEmpty)
		}

		gqlServices, ok := app["graphql"].(map[string]interface{})
		u.So(t, ok, gc.ShouldBeTrue)

		c, ok := gqlServices["config"].(map[string]interface{})
		u.So(t, ok, gc.ShouldBeTrue)

		useNatPluralization, ok := c["use_natural_pluralization"]
		u.So(t, ok, gc.ShouldBeTrue)
		u.So(t, useNatPluralization, gc.ShouldBeTrue)

		customResolversList, ok := gqlServices["custom_resolvers"].([]interface{})
		u.So(t, ok, gc.ShouldBeTrue)
		u.So(t, customResolversList, gc.ShouldHaveLength, 1)
		for _, customResolver := range customResolversList {
			customResolverMap, ok := customResolver.(map[string]interface{})
			u.So(t, ok, gc.ShouldBeTrue)
			u.So(t, customResolverMap["field_name"], gc.ShouldNotBeEmpty)
		}

		servicesList, ok := app["services"].([]interface{})
		u.So(t, ok, gc.ShouldBeTrue)
		u.So(t, servicesList, gc.ShouldHaveLength, 3)
		for _, svc := range servicesList {
			svcMap, ok := svc.(map[string]interface{})
			u.So(t, ok, gc.ShouldBeTrue)

			u.So(t, svcMap["config"], gc.ShouldNotBeEmpty)
			u.So(t, svcMap["rules"], gc.ShouldNotBeEmpty)
			u.So(t, svcMap["incoming_webhooks"], gc.ShouldNotBeEmpty)
		}

		envsMap, ok := app["environments"].(map[string]interface{})
		u.So(t, envsMap, gc.ShouldHaveLength, 5)
		u.So(t, ok, gc.ShouldBeTrue)
		for _, env := range envsMap {
			envMap, ok := env.(map[string]interface{})
			u.So(t, ok, gc.ShouldBeTrue)
			u.So(t, envMap["values"], gc.ShouldNotBeEmpty)

			values, ok := envMap["values"].(map[string]interface{})
			u.So(t, ok, gc.ShouldBeTrue)

			greeting, ok := values["greeting"]
			u.So(t, ok, gc.ShouldBeTrue)

			u.So(t, greeting, gc.ShouldNotBeEmpty)
		}
	})
}
