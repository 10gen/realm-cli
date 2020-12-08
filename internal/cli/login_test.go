package cli

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"

	expect "github.com/Netflix/go-expect"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestLoginSetup(t *testing.T) {
	t.Run("Should construct a Realm client with the configured base url", func(t *testing.T) {
		config := CommandConfig{RealmBaseURL: "http://localhost:8080"}

		cmd := &loginCommand{
			publicAPIKey:  "publicAPIKey",
			privateAPIKey: "privateAPIKey",
		}

		err := cmd.Setup(nil, nil, config)
		assert.Nil(t, err)
		assert.NotNil(t, cmd.realmClient)
	})

	for _, tc := range []struct {
		description   string
		publicAPIKey  string
		privateAPIKey string
		procedure     func(c *expect.Console)
		test          func(t *testing.T, cmd *loginCommand)
	}{
		{
			description:   "Should prompt for public api key when not provided",
			privateAPIKey: "password",
			procedure: func(c *expect.Console) {
				c.ExpectString("API Key")
				c.SendLine("username")
			},
			test: func(t *testing.T, cmd *loginCommand) {
				assert.Equal(t, "username", cmd.publicAPIKey)
			},
		},
		{
			description:  "Should prompt for private api key when not provided",
			publicAPIKey: "username",
			procedure: func(c *expect.Console) {
				c.ExpectString("Private API Key")
				c.SendLine("password")
				c.ExpectEOF()
			},
			test: func(t *testing.T, cmd *loginCommand) {
				assert.Equal(t, "password", cmd.privateAPIKey)
			},
		},
		{
			description: "Should prompt for both api keys when not provided",
			procedure: func(c *expect.Console) {
				c.ExpectString("API Key")
				c.SendLine("username")
				c.ExpectString("Private API Key")
				c.SendLine("password")
				c.ExpectEOF()
			},
			test: func(t *testing.T, cmd *loginCommand) {
				assert.Equal(t, "username", cmd.publicAPIKey)
				assert.Equal(t, "password", cmd.privateAPIKey)
			},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			out := new(bytes.Buffer)
			console, _, ui, consoleErr := mock.NewVT10XConsole(mock.UIOptions{}, out)
			assert.Nil(t, consoleErr)
			defer console.Close()

			profile, profileErr := NewProfile(primitive.NewObjectID().Hex())
			assert.Nil(t, profileErr)

			cmd := &loginCommand{
				publicAPIKey:  tc.publicAPIKey,
				privateAPIKey: tc.privateAPIKey,
			}

			doneCh := make(chan (struct{}))
			go func() {
				defer close(doneCh)
				tc.procedure(console)
			}()

			err := cmd.Setup(profile, ui, CommandConfig{})
			assert.Nil(t, err)

			console.Tty().Close() // flush the writers
			<-doneCh              // wait for procedure to complete

			tc.test(t, cmd)
		})
	}
}

func TestLoginHandler(t *testing.T) {
	t.Run("With no existing credentials handler should save new credentials", func(t *testing.T) {
		tmpDir, teardownTmpDir, tmpDirErr := u.NewTempDir("home")
		assert.Nil(t, tmpDirErr)
		defer teardownTmpDir()

		_, teardownHomeDir := u.SetupHomeDir(tmpDir)
		defer teardownHomeDir()

		profile, profileErr := NewProfile(primitive.NewObjectID().Hex())
		assert.Nil(t, profileErr)

		_, statErr := os.Stat(profile.path())
		assert.NotNil(t, statErr)
		assert.True(t, os.IsNotExist(statErr), "profile must not exist")

		realmClient := mock.RealmClient{}
		realmClient.AuthenticateFn = func(publicAPIKey, privateAPIKey string) (realm.Session, error) {
			return realm.Session{
				AccessToken:  "accessToken",
				RefreshToken: "refreshToken",
			}, nil
		}

		cmd := &loginCommand{
			realmClient:   realmClient,
			publicAPIKey:  "publicAPIKey",
			privateAPIKey: "privateAPIKey",
		}

		out := new(bytes.Buffer)
		ui := mock.NewUI(mock.UIOptions{}, out)

		assert.Nil(t, cmd.Handler(profile, ui, nil))

		expectedUser := User{"publicAPIKey", "privateAPIKey"}
		expectedSession := realm.Session{"accessToken", "refreshToken"}

		assert.Match(t, expectedUser, profile.GetUser())
		assert.Match(t, expectedSession, profile.GetSession())
		ensureProfileContents(t, profile, expectedUser, expectedSession)
	})

	t.Run("With existing credentials", func(t *testing.T) {
		setup := func(t *testing.T) (*Profile, realm.Client, func()) {
			tmpDir, teardownTmpDir, tmpDirErr := u.NewTempDir("home")
			assert.Nil(t, tmpDirErr)

			_, teardownHomeDir := u.SetupHomeDir(tmpDir)

			profile, profileErr := NewProfile(primitive.NewObjectID().Hex())
			assert.Nil(t, profileErr)

			profile.SetUser("existingUser", "existing-password")
			profile.SetSession("existingAccessToken", "existingRefreshToken")
			assert.Nil(t, profile.Save())

			realmClient := mock.RealmClient{}
			realmClient.AuthenticateFn = func(publicAPIKey, privateAPIKey string) (realm.Session, error) {
				return realm.Session{
					AccessToken:  "newAccessToken",
					RefreshToken: "newRefreshToken",
				}, nil
			}

			return profile, realmClient, func() {
				teardownHomeDir()
				teardownTmpDir()
			}
		}

		t.Run("That match the attempted login credentials handler should refresh the existing session", func(t *testing.T) {
			profile, realmClient, teardown := setup(t)
			defer teardown()

			cmd := &loginCommand{
				realmClient:   realmClient,
				publicAPIKey:  "existingUser",
				privateAPIKey: "existing-password",
			}

			out := new(bytes.Buffer)
			ui := mock.NewUI(mock.UIOptions{}, out)

			assert.Nil(t, cmd.Handler(profile, ui, nil))

			expectedUser := User{"existingUser", "existing-password"}
			expectedSession := realm.Session{"newAccessToken", "newRefreshToken"}

			assert.Match(t, expectedUser, profile.GetUser())
			assert.Match(t, expectedSession, profile.GetSession())
			ensureProfileContents(t, profile, expectedUser, expectedSession)
		})

		t.Run("That do not match the attempted login credentials should prompt the user to continue", func(t *testing.T) {
			for _, tc := range []struct {
				description     string
				confirmAnswer   string
				expectedUser    User
				expectedSession realm.Session
			}{
				{
					description:     "And do nothing if the user does not want to proceed",
					confirmAnswer:   "n",
					expectedUser:    User{"existingUser", "existing-password"},
					expectedSession: realm.Session{"existingAccessToken", "existingRefreshToken"},
				},
				{
					description:     "And save a new session if the user does want to proceed",
					confirmAnswer:   "y",
					expectedUser:    User{"newUser", "new-password"},
					expectedSession: realm.Session{"newAccessToken", "newRefreshToken"},
				},
			} {
				t.Run(tc.description, func(t *testing.T) {
					profile, realmClient, teardown := setup(t)
					defer teardown()

					cmd := &loginCommand{
						realmClient:   realmClient,
						publicAPIKey:  "newUser",
						privateAPIKey: "new-password",
					}

					out := new(bytes.Buffer)
					console, _, ui, consoleErr := mock.NewVT10XConsole(mock.UIOptions{}, out)
					assert.Nil(t, consoleErr)

					doneCh := make(chan (struct{}))
					go func() {
						defer close(doneCh)
						console.ExpectString("This action will terminate the existing session for user: existingUser (********-password), would you like to proceed?")
						console.SendLine(tc.confirmAnswer)
						console.ExpectEOF()
					}()

					err := cmd.Handler(profile, ui, nil)
					assert.Nil(t, err)

					assert.Nil(t, console.Tty().Close())
					<-doneCh

					assert.Match(t, tc.expectedUser, profile.GetUser())
					assert.Match(t, tc.expectedSession, profile.GetSession())
					ensureProfileContents(t, profile, tc.expectedUser, tc.expectedSession)
				})
			}
		})
	})
}

func TestLoginFeedback(t *testing.T) {
	t.Run("Feedback should print a message that login was successful", func(t *testing.T) {
		out := new(bytes.Buffer)
		ui := mock.NewUI(mock.UIOptions{}, out)

		cmd := &loginCommand{}

		err := cmd.Feedback(nil, ui)
		assert.Nil(t, err)

		assert.Equal(t, "01:23:45 UTC INFO  Successfully logged in.\n", out.String())
	})
}

func ensureProfileContents(t *testing.T, profile *Profile, user User, session realm.Session) {
	contents, err := ioutil.ReadFile(profile.path())
	assert.Nil(t, err)
	assert.True(t, strings.Contains(string(contents), fmt.Sprintf(`%s:
  access_token: %s
  private_api_key: %s
  public_api_key: %s
  refresh_token: %s
`, profile.Name, session.AccessToken, user.PrivateAPIKey, user.PublicAPIKey, session.RefreshToken)), "profile must contain the expected contents")
}
