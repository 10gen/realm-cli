package app

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/local"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"

	"github.com/Netflix/go-expect"
)

func TestAppDescribeInputsResolve(t *testing.T) {
	t.Run("should set app if in app directory", func(t *testing.T) {
		profile, teardown := mock.NewProfileFromTmpDir(t, "app-describe-test")
		defer teardown()

		assert.Nil(t, ioutil.WriteFile(
			filepath.Join(profile.WorkingDirectory, local.FileRealmConfig.String()),
			[]byte(`{"config_version":20210101,"app_id":"test-app-abcde","name":"test-app"}`),
			0666,
		))

		inputs := describeInputs{}
		assert.Nil(t, inputs.Resolve(profile, nil))
		assert.Equal(t, "test-app-abcde", inputs.App)
	})

	t.Run("should not set app if not in app directory", func(t *testing.T) {
		profile := mock.NewProfile(t)

		inputs := describeInputs{}
		assert.Nil(t, inputs.Resolve(profile, nil))
		assert.Equal(t, "", inputs.App)
	})
}

func TestAppDescribeHandler(t *testing.T) {
	t.Run("should error if apps not found", func(t *testing.T) {
		realmClient := mock.RealmClient{}

		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{}, nil
		}

		cmd := &CommandDescribe{inputs: describeInputs{cli.ProjectInputs{App: "test-app"}}}
		assert.Equal(t, cli.ErrAppNotFound{App: "test-app"}, cmd.Handler(nil, nil, cli.Clients{Realm: realmClient}))
	})

	for _, tc := range []struct {
		description string
		inputs      describeInputs
		apps        []realm.App
		procedure   func(c *expect.Console)
	}{
		{
			description: "should prompt user to select app to describe if nothing set",
			apps:        []realm.App{{ID: "456", ClientAppID: "test-app-abcde", GroupID: "123"}, {ClientAppID: "another-one-efghi"}},
			procedure: func(c *expect.Console) {
				c.ExpectString("Select App")
				c.Send("test-app-abcde")
				c.SendLine(" ")
				c.ExpectEOF()
			},
		},
		{
			description: "should prompt user to select app to describe if app is set and not found",
			inputs:      describeInputs{cli.ProjectInputs{App: "another-test"}},
			apps:        []realm.App{{ID: "456", ClientAppID: "test-app-abcde", GroupID: "123"}, {ClientAppID: "another-one-efghi"}},
			procedure: func(c *expect.Console) {
				c.ExpectString("Select App")
				c.Send("test-app-abcde")
				c.SendLine(" ")
				c.ExpectEOF()
			},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			_, console, _, ui, err := mock.NewVT10XConsole()
			assert.Nil(t, err)
			defer console.Close()

			doneCh := make(chan (struct{}))
			go func() {
				defer close(doneCh)
				tc.procedure(console)
			}()

			realmClient := mock.RealmClient{}
			realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
				return tc.apps, nil
			}
			var groupIDActual, appIDActual string
			realmClient.AppDescriptionFn = func(groupID, appID string) (realm.AppDescription, error) {
				groupIDActual = groupID
				appIDActual = appID
				return realm.AppDescription{}, nil
			}

			cmd := &CommandDescribe{inputs: tc.inputs}
			assert.Nil(t, cmd.Handler(nil, ui, cli.Clients{Realm: realmClient}))
			assert.Equal(t, "123", groupIDActual)
			assert.Equal(t, "456", appIDActual)
		})
	}

	t.Run("should describe app", func(t *testing.T) {
		out, ui := mock.NewUI()

		realmClient := mock.RealmClient{}
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{{ID: "456", ClientAppID: "todo-abcde", GroupID: "123"}}, nil
		}
		realmClient.AppDescriptionFn = func(groupID, appID string) (realm.AppDescription, error) {
			return realm.AppDescription{
				ClientAppID: "todo-abcde",
				Name:        "todo",
				RealmURL:    fmt.Sprintf("https://admin-base.url/groups/%s/apps/%s/dashboard", groupID, appID),
				DataSources: []realm.DataSourceSummary{
					{
						Name:       "mongodb-atlas",
						Type:       "mongodb-atlas",
						DataSource: "Cluster0",
					},
					{
						Name:       "mongodb-datalake",
						Type:       "datalake",
						DataSource: "DataLake0",
					},
					{
						Name: "mdb1",
						Type: "mongodb",
					},
				},
				HTTPEndpoints: realm.HTTPEndpoints{[]interface{}{
					realm.HTTPServiceSummary{
						Name: "http",
						IncomingWebhooks: []realm.IncomingWebhookSummary{
							{
								Name: "webhook0",
								URL:  "https://webhook-base.url/api/client/v2.0/app/todo-abcde/service/http/incoming_webhook/webhook0",
							},
						},
					},
					realm.EndpointSummary{
						Route:      "/bob/the/route/builder",
						HTTPMethod: "GET",
						URL:        "https://endpoint-base.url/app/todo-abcde/endpoint/bob/the/route/builder",
					},
					realm.EndpointSummary{
						Route:      "/bob/the/route/builder",
						HTTPMethod: "*",
						URL:        "https://endpoint-base.url/app/todo-abcde/endpoint/bob/the/route/builder",
					},
				}},
				ServiceDescs: []realm.ServiceSummary{
					{
						Name:             "tw1",
						Type:             "twilio",
						IncomingWebhooks: []realm.IncomingWebhookSummary{},
					},
				},
				AuthProviders: []realm.AuthProviderSummary{
					{
						Name:    "oauth2/google",
						Type:    "oauth2/google",
						Enabled: true,
					},
					{
						Name:    "oauth2/facebook",
						Type:    "oauth2/facebook",
						Enabled: true,
					},
				},
				CustomUserData: realm.CustomUserDataSummary{
					Enabled:     true,
					DataSource:  "Cluster0",
					Database:    "db1",
					Collection:  "col1",
					UserIDField: "id",
				},
				Values: []string{
					"value1",
					"value2",
				},
				Hosting: realm.HostingSummary{
					Enabled: true,
					Status:  "setup_ok",
					URL:     "https://hosting.domain/",
				},
				Functions: []realm.FunctionSummary{
					{
						Name: "func1",
						Path: "",
					},
					{
						Name: "func2",
						Path: "nested",
					},
				},
				Sync: realm.SyncSummary{
					State:                  "enabled",
					DataSource:             "Cluster0",
					Database:               "db1",
					DevelopmentModeEnabled: true,
				},
				GraphQL: realm.GraphQLSummary{
					URL: "https://api-base.url/api/client/v2.0/app/todo-abcde/graphql",
					CustomResolvers: []string{
						"type.fieldName",
					},
				},
				Environment: "development",
				EventSubscription: []realm.EventSubscriptionSummary{
					{
						Name:    "trigger1",
						Type:    "DATABASE",
						Enabled: true,
					},
				},
				LogForwarders: []realm.LogForwarderSummary{
					{
						Name:    "logforwarder1",
						Enabled: true,
					},
				},
			}, nil
		}

		cmd := &CommandDescribe{inputs: describeInputs{cli.ProjectInputs{App: "todo-abcde"}}}
		assert.Nil(t, cmd.Handler(nil, ui, cli.Clients{Realm: realmClient}))

		assert.Equal(t, `App description
{
  "client_app_id": "todo-abcde",
  "name": "todo",
  "realm_url": "https://admin-base.url/groups/123/apps/456/dashboard",
  "data_sources": [
    {
      "name": "mongodb-atlas",
      "type": "mongodb-atlas",
      "data_source": "Cluster0"
    },
    {
      "name": "mongodb-datalake",
      "type": "datalake",
      "data_source": "DataLake0"
    },
    {
      "name": "mdb1",
      "type": "mongodb",
      "data_source": ""
    }
  ],
  "http_endpoints": [
    {
      "name": "http",
      "webhooks": [
        {
          "name": "webhook0",
          "url": "https://webhook-base.url/api/client/v2.0/app/todo-abcde/service/http/incoming_webhook/webhook0"
        }
      ]
    },
    {
      "route": "/bob/the/route/builder",
      "http_method": "GET",
      "url": "https://endpoint-base.url/app/todo-abcde/endpoint/bob/the/route/builder"
    },
    {
      "route": "/bob/the/route/builder",
      "http_method": "*",
      "url": "https://endpoint-base.url/app/todo-abcde/endpoint/bob/the/route/builder"
    }
  ],
  "services": [
    {
      "name": "tw1",
      "type": "twilio",
      "webhooks": []
    }
  ],
  "auth_providers": [
    {
      "name": "oauth2/google",
      "type": "oauth2/google",
      "enabled": true
    },
    {
      "name": "oauth2/facebook",
      "type": "oauth2/facebook",
      "enabled": true
    }
  ],
  "custom_user_data": {
    "enabled": true,
    "data_source": "Cluster0",
    "database": "db1",
    "collection": "col1",
    "user_id_field": "id"
  },
  "values": [
    "value1",
    "value2"
  ],
  "hosting": {
    "enabled": true,
    "status": "setup_ok",
    "url": "https://hosting.domain/"
  },
  "functions": [
    {
      "name": "func1",
      "path": ""
    },
    {
      "name": "func2",
      "path": "nested"
    }
  ],
  "sync": {
    "state": "enabled",
    "data_source": "Cluster0",
    "database": "db1",
    "development_mode_enabled": true
  },
  "graphql": {
    "url": "https://api-base.url/api/client/v2.0/app/todo-abcde/graphql",
    "custom_resolvers": [
      "type.fieldName"
    ]
  },
  "environment": "development",
  "event_subscription": [
    {
      "name": "trigger1",
      "type": "DATABASE",
      "enabled": true
    }
  ],
  "log_forwarders": [
    {
      "name": "logforwarder1",
      "enabled": true
    }
  ]
}
`, out.String())
	})

	t.Run("should return an error when finding apps fails", func(t *testing.T) {
		realmClient := mock.RealmClient{}
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return nil, errors.New("realm client error")
		}

		cmd := &CommandDescribe{inputs: describeInputs{cli.ProjectInputs{Project: "test", App: "test-app"}}}
		assert.Equal(t, errors.New("realm client error"), cmd.Handler(nil, nil, cli.Clients{Realm: realmClient}))
	})

	t.Run("should return an error when describing an app fails", func(t *testing.T) {
		realmClient := mock.RealmClient{}
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{{Name: "test-app"}}, nil
		}
		realmClient.AppDescriptionFn = func(groupID, appID string) (realm.AppDescription, error) {
			return realm.AppDescription{}, errors.New("realm client error")
		}

		cmd := &CommandDescribe{inputs: describeInputs{cli.ProjectInputs{Project: "test", App: "test-app"}}}
		assert.Equal(t, errors.New("realm client error"), cmd.Handler(nil, nil, cli.Clients{Realm: realmClient}))
	})
}
