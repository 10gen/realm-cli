package local

import (
	"testing"

	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestAddAuthProvider(t *testing.T) {
	for _, tc := range []struct {
		description     string
		appData         AppData
		name            string
		config          map[string]interface{}
		appDataExpected AppData
	}{
		{
			description: "should add an auth provider to app stitch json",
			appData:     &AppStitchJSON{},
			config: map[string]interface{}{
				"name":     "api-key",
				"type":     "api-key",
				"disabled": true,
			},
			appDataExpected: &AppStitchJSON{AppDataV1{AppStructureV1{
				AuthProviders: []map[string]interface{}{
					{
						"name":     "api-key",
						"type":     "api-key",
						"disabled": true,
					},
				},
			}}},
		},
		{
			description: "should add an auth provider to app config json",
			appData:     &AppConfigJSON{},
			config: map[string]interface{}{
				"name":     "api-key",
				"type":     "api-key",
				"disabled": true,
			},
			appDataExpected: &AppConfigJSON{AppDataV1{AppStructureV1{
				AuthProviders: []map[string]interface{}{
					{
						"name":     "api-key",
						"type":     "api-key",
						"disabled": true,
					},
				},
			}}},
		},
		{
			description: "should create and add an auth provider to app realm config json",
			appData:     &AppRealmConfigJSON{},
			name:        "api-key",
			config: map[string]interface{}{
				"name":     "api-key",
				"type":     "api-key",
				"disabled": true,
			},
			appDataExpected: &AppRealmConfigJSON{AppDataV2{AppStructureV2{
				Auth: AuthStructure{
					Providers: map[string]interface{}{
						"api-key": map[string]interface{}{
							"name":     "api-key",
							"type":     "api-key",
							"disabled": true,
						},
					},
				},
			}}},
		},
		{
			description: "should add an auth provider to app realm config json",
			appData: &AppRealmConfigJSON{AppDataV2{AppStructureV2{
				Auth: AuthStructure{
					Providers: map[string]interface{}{},
				},
			}}},
			name: "api-key",
			config: map[string]interface{}{
				"name":     "api-key",
				"type":     "api-key",
				"disabled": true,
			},
			appDataExpected: &AppRealmConfigJSON{AppDataV2{AppStructureV2{
				Auth: AuthStructure{
					Providers: map[string]interface{}{
						"api-key": map[string]interface{}{
							"name":     "api-key",
							"type":     "api-key",
							"disabled": true,
						},
					},
				},
			}}},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			AddAuthProvider(tc.appData, tc.name, tc.config)

			assert.Equal(t, tc.appDataExpected, tc.appData)
		})
	}
}

func TestAddDataSource(t *testing.T) {
	for _, tc := range []struct {
		description     string
		appData         AppData
		config          map[string]interface{}
		appDataExpected AppData
	}{
		{
			description: "should add a service to app stitch json",
			appData:     &AppStitchJSON{},
			config: map[string]interface{}{
				"name": "mongodb-atlas",
				"type": "mongodb-atlas",
				"config": map[string]interface{}{
					"clusterName":         "Cluster0",
					"readPreference":      "primary",
					"wireProtocolEnabled": false,
				},
			},
			appDataExpected: &AppStitchJSON{AppDataV1{AppStructureV1{
				Services: []ServiceStructure{
					{
						Config: map[string]interface{}{
							"name": "mongodb-atlas",
							"type": "mongodb-atlas",
							"config": map[string]interface{}{
								"clusterName":         "Cluster0",
								"readPreference":      "primary",
								"wireProtocolEnabled": false,
							},
						},
					},
				},
			}}},
		},
		{
			description: "should add a service to app config json",
			appData:     &AppConfigJSON{},
			config: map[string]interface{}{
				"name": "mongodb-atlas",
				"type": "mongodb-atlas",
				"config": map[string]interface{}{
					"clusterName":         "Cluster0",
					"readPreference":      "primary",
					"wireProtocolEnabled": false,
				},
			},
			appDataExpected: &AppConfigJSON{AppDataV1{AppStructureV1{
				Services: []ServiceStructure{
					{
						Config: map[string]interface{}{
							"name": "mongodb-atlas",
							"type": "mongodb-atlas",
							"config": map[string]interface{}{
								"clusterName":         "Cluster0",
								"readPreference":      "primary",
								"wireProtocolEnabled": false,
							},
						},
					},
				},
			}}},
		},
		{
			description: "should add a data source to app realm config json",
			appData:     &AppRealmConfigJSON{},
			config: map[string]interface{}{
				"name": "mongodb-atlas",
				"type": "mongodb-atlas",
				"config": map[string]interface{}{
					"clusterName":         "Cluster0",
					"readPreference":      "primary",
					"wireProtocolEnabled": false,
				},
			},
			appDataExpected: &AppRealmConfigJSON{AppDataV2{AppStructureV2{
				DataSources: []DataSourceStructure{
					{
						Config: map[string]interface{}{
							"name": "mongodb-atlas",
							"type": "mongodb-atlas",
							"config": map[string]interface{}{
								"clusterName":         "Cluster0",
								"readPreference":      "primary",
								"wireProtocolEnabled": false,
							},
						},
					},
				},
			}}},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			AddDataSource(tc.appData, tc.config)

			assert.Equal(t, tc.appDataExpected, tc.appData)
		})
	}
}
