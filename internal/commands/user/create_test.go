package user

import (
	"errors"
	"fmt"
	"testing"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestUserCreateHandler(t *testing.T) {
	app := realm.App{
		ID:          primitive.NewObjectID().Hex(),
		GroupID:     primitive.NewObjectID().Hex(),
		ClientAppID: "eggcorn-abcde",
		Name:        "eggcorn",
	}

	newMockClient := func() mock.RealmClient {
		realmClient := mock.RealmClient{}
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{app}, nil
		}
		return realmClient
	}

	t.Run("should create a email password user when email type is set", func(t *testing.T) {
		id := primitive.NewObjectID().Hex()
		testUser := realm.User{ID: id, Type: "normal", Data: map[string]interface{}{"email": "user@domain.com"}}

		out, ui := mock.NewUI()

		realmClient := newMockClient()
		realmClient.CreateUserFn = func(groupID, appID, email, password string) (realm.User, error) {
			return testUser, nil
		}

		cmd := &CommandCreate{createInputs{UserType: userTypeEmailPassword}}

		assert.Nil(t, cmd.Handler(nil, ui, cli.Clients{Realm: realmClient}))
		assert.Equal(t, fmt.Sprintf(`Successfully created user
{
  "id": %q,
  "enabled": true,
  "email": "user@domain.com",
  "type": "normal"
}
`, id), out.String())
	})

	t.Run("should create an api key when apiKey type is set", func(t *testing.T) {
		id := primitive.NewObjectID().Hex()
		testAPIKey := realm.APIKey{ID: id, Name: "name", Key: "key"}

		out, ui := mock.NewUI()

		realmClient := newMockClient()
		realmClient.CreateAPIKeyFn = func(groupID, appID, apiKeyName string) (realm.APIKey, error) {
			return testAPIKey, nil
		}

		cmd := &CommandCreate{createInputs{UserType: userTypeAPIKey}}

		assert.Nil(t, cmd.Handler(nil, ui, cli.Clients{Realm: realmClient}))
		assert.Equal(t, fmt.Sprintf(`Successfully created API Key
{
  "id": %q,
  "enabled": true,
  "name": "name",
  "key": "key"
}
`, id), out.String())
	})

	t.Run("should return an error", func(t *testing.T) {
		for _, tc := range []struct {
			description string
			userType    userType
			setupClient func() realm.Client
			expectedErr error
		}{
			{
				description: "when resolving the app fails",
				setupClient: func() realm.Client {
					realmClient := mock.RealmClient{}
					realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
						return nil, errors.New("something bad happened")
					}
					return realmClient
				},
				expectedErr: errors.New("something bad happened"),
			},
			{
				description: "when creating an email password user fails",
				userType:    userTypeEmailPassword,
				setupClient: func() realm.Client {
					realmClient := mock.RealmClient{}
					realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
						return []realm.App{app}, nil
					}
					realmClient.CreateUserFn = func(groupID, appID, email, password string) (realm.User, error) {
						return realm.User{}, errors.New("something bad happened")
					}
					return realmClient
				},
				expectedErr: errors.New("failed to create user: something bad happened"),
			},
			{
				description: "when creating an api key fails",
				userType:    userTypeAPIKey,
				setupClient: func() realm.Client {
					realmClient := mock.RealmClient{}
					realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
						return []realm.App{app}, nil
					}
					realmClient.CreateAPIKeyFn = func(groupID, appID, apiKeyName string) (realm.APIKey, error) {
						return realm.APIKey{}, errors.New("something bad happened")
					}
					return realmClient
				},
				expectedErr: errors.New("failed to create API Key: something bad happened"),
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				realmClient := tc.setupClient()

				cmd := &CommandCreate{createInputs{UserType: tc.userType}}

				err := cmd.Handler(nil, nil, cli.Clients{Realm: realmClient})
				assert.Equal(t, tc.expectedErr, err)
			})
		}
	})
}
