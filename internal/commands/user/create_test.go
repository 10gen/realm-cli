package user

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestUserCreateSetup(t *testing.T) {
	t.Run("Should construct a Realm client with the configured base url", func(t *testing.T) {
		profile := mock.NewProfile(t)
		profile.SetRealmBaseURL("http://localhost:8080")

		cmd := &CommandCreate{inputs: createInputs{
			UserType: userTypeEmailPassword,
			Email:    "user@domain.com",
			Password: "password",
		}}
		assert.Nil(t, cmd.realmClient)

		assert.Nil(t, cmd.Setup(profile, nil))
		assert.NotNil(t, cmd.realmClient)
	})
}

func TestUserCreateHandler(t *testing.T) {
	testApp := realm.App{
		ID:          primitive.NewObjectID().Hex(),
		GroupID:     primitive.NewObjectID().Hex(),
		ClientAppID: "eggcorn-abcde",
		Name:        "eggcorn",
	}

	newMockClient := func() mock.RealmClient {
		realmClient := mock.RealmClient{}
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{testApp}, nil
		}
		return realmClient
	}

	t.Run("Should create a email password user when email type is set", func(t *testing.T) {
		testUser := realm.User{}

		realmClient := newMockClient()
		realmClient.CreateUserFn = func(groupID, appID, email, password string) (realm.User, error) {
			return testUser, nil
		}

		cmd := &CommandCreate{
			inputs:      createInputs{UserType: userTypeEmailPassword},
			realmClient: realmClient,
		}

		assert.Nil(t, cmd.Handler(nil, nil))
		assert.Equal(t, realm.APIKey{}, cmd.outputs.apiKey)
		assert.Equal(t, testUser, cmd.outputs.user)
	})

	t.Run("Should create an api key when apiKey type is set", func(t *testing.T) {
		testAPIKey := realm.APIKey{}

		realmClient := newMockClient()
		realmClient.CreateAPIKeyFn = func(groupID, appID, apiKeyName string) (realm.APIKey, error) {
			return testAPIKey, nil
		}

		cmd := &CommandCreate{
			inputs:      createInputs{UserType: userTypeAPIKey},
			realmClient: realmClient,
		}

		assert.Nil(t, cmd.Handler(nil, nil))
		assert.Equal(t, testAPIKey, cmd.outputs.apiKey)
		assert.Equal(t, realm.User{}, cmd.outputs.user)
	})

	t.Run("Should return an error", func(t *testing.T) {
		for _, tc := range []struct {
			description string
			userType    userType
			setupClient func() realm.Client
			expectedErr error
		}{
			{
				description: "When resolving the app fails",
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
				description: "When creating an email password user fails",
				userType:    userTypeEmailPassword,
				setupClient: func() realm.Client {
					realmClient := mock.RealmClient{}
					realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
						return []realm.App{testApp}, nil
					}
					realmClient.CreateUserFn = func(groupID, appID, email, password string) (realm.User, error) {
						return realm.User{}, errors.New("something bad happened")
					}
					return realmClient
				},
				expectedErr: errors.New("failed to create user: something bad happened"),
			},
			{
				description: "When creating an api key fails",
				userType:    userTypeAPIKey,
				setupClient: func() realm.Client {
					realmClient := mock.RealmClient{}
					realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
						return []realm.App{testApp}, nil
					}
					realmClient.CreateAPIKeyFn = func(groupID, appID, apiKeyName string) (realm.APIKey, error) {
						return realm.APIKey{}, errors.New("something bad happened")
					}
					return realmClient
				},
				expectedErr: errors.New("failed to create api key: something bad happened"),
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				realmClient := tc.setupClient()

				cmd := &CommandCreate{
					inputs:      createInputs{UserType: tc.userType},
					realmClient: realmClient,
				}

				err := cmd.Handler(nil, nil)
				assert.Equal(t, tc.expectedErr, err)
			})
		}
	})

	t.Run("Should create an api key when apiKey type is set", func(t *testing.T) {
		testAPIKey := realm.APIKey{}

		realmClient := newMockClient()
		realmClient.CreateAPIKeyFn = func(groupID, appID, apiKeyName string) (realm.APIKey, error) {
			return testAPIKey, nil
		}

		cmd := &CommandCreate{
			inputs:      createInputs{UserType: userTypeAPIKey},
			realmClient: realmClient,
		}

		assert.Nil(t, cmd.Handler(nil, nil))
		assert.Equal(t, testAPIKey, cmd.outputs.apiKey)
		assert.Equal(t, realm.User{}, cmd.outputs.user)
	})
}

func TestUserCreateFeedback(t *testing.T) {
	id := primitive.NewObjectID().Hex()

	for _, tc := range []struct {
		description    string
		userType       userType
		outputs        outputs
		expectedOutput string
	}{
		{
			description: "Should print the email password user details when email type is set",
			userType:    userTypeEmailPassword,
			outputs: outputs{user: realm.User{
				ID:   id,
				Type: "normal",
				Data: map[string]interface{}{"email": "user@domain.com"},
			}},
			expectedOutput: strings.Join([]string{
				"01:23:45 UTC INFO  Successfully created user",
				"  ID                        Enabled  Email            Type  ",
				"  ------------------------  -------  ---------------  ------",
				fmt.Sprintf("  %s  true     user@domain.com  normal\n", id),
			}, "\n"),
		},
		{
			description: "Should print the api key details when apiKey type is set",
			userType:    userTypeAPIKey,
			outputs: outputs{apiKey: realm.APIKey{
				ID:   id,
				Name: "name",
				Key:  "key",
			}},
			expectedOutput: strings.Join([]string{
				"01:23:45 UTC INFO  Successfully created api key",
				"  ID                        Enabled  Name  API Key",
				"  ------------------------  -------  ----  -------",
				fmt.Sprintf("  %s  true     name  key    \n", id),
			}, "\n"),
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			out, ui := mock.NewUI()

			cmd := &CommandCreate{
				inputs:  createInputs{UserType: tc.userType},
				outputs: tc.outputs,
			}

			assert.Nil(t, cmd.Feedback(nil, ui))

			assert.Equal(t, tc.expectedOutput, out.String())
		})
	}
}
