package login

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/10gen/realm-cli/internal/auth"
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
)

func TestLoginSetup(t *testing.T) {
	t.Run("Should construct a Realm client with the configured base url", func(t *testing.T) {
		profile := mock.NewProfile(t)
		profile.SetRealmBaseURL("http://localhost:8080")

		cmd := &Command{inputs: inputs{
			PublicAPIKey:  "publicAPIKey",
			PrivateAPIKey: "privateAPIKey",
		}}

		err := cmd.Setup(profile, nil)
		assert.Nil(t, err)
		assert.NotNil(t, cmd.realmClient)
	})
}

func TestLoginHandler(t *testing.T) {
	t.Run("With no existing credentials handler should save new credentials", func(t *testing.T) {
		tmpDir, teardownTmpDir, tmpDirErr := u.NewTempDir("home")
		assert.Nil(t, tmpDirErr)
		defer teardownTmpDir()

		_, teardownHomeDir := u.SetupHomeDir(tmpDir)
		defer teardownHomeDir()

		profile := mock.NewProfile(t)

		_, statErr := os.Stat(profile.Path())
		assert.NotNil(t, statErr)
		assert.True(t, os.IsNotExist(statErr), "profile must not exist")

		realmClient := mock.RealmClient{}
		realmClient.AuthenticateFn = func(publicAPIKey, privateAPIKey string) (realm.Session, error) {
			return realm.Session{
				AccessToken:  "accessToken",
				RefreshToken: "refreshToken",
			}, nil
		}

		cmd := &Command{
			realmClient: realmClient,
			inputs: inputs{
				PublicAPIKey:  "publicAPIKey",
				PrivateAPIKey: "privateAPIKey",
			},
		}

		_, ui := mock.NewUI()

		assert.Nil(t, cmd.Handler(profile, ui))

		expectedUser := auth.User{"publicAPIKey", "privateAPIKey"}
		expectedSession := auth.Session{"accessToken", "refreshToken"}

		assert.Equal(t, expectedUser, profile.User())
		assert.Equal(t, expectedSession, profile.Session())

		ensureProfileContents(t, profile, expectedUser, expectedSession)
	})

	t.Run("With existing credentials", func(t *testing.T) {
		setup := func(t *testing.T) (*cli.Profile, realm.Client, func()) {
			tmpDir, teardownTmpDir, tmpDirErr := u.NewTempDir("home")
			assert.Nil(t, tmpDirErr)

			_, teardownHomeDir := u.SetupHomeDir(tmpDir)

			profile := mock.NewProfile(t)

			profile.SetUser(auth.User{"existingUser", "existing-password"})
			profile.SetSession(auth.Session{"existingAccessToken", "existingRefreshToken"})
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

			cmd := &Command{
				realmClient: realmClient,
				inputs: inputs{
					PublicAPIKey:  "existingUser",
					PrivateAPIKey: "existing-password",
				},
			}

			_, ui := mock.NewUI()

			assert.Nil(t, cmd.Handler(profile, ui))

			expectedUser := auth.User{"existingUser", "existing-password"}
			expectedSession := auth.Session{"newAccessToken", "newRefreshToken"}

			assert.Equal(t, expectedUser, profile.User())
			assert.Equal(t, expectedSession, profile.Session())

			ensureProfileContents(t, profile, expectedUser, expectedSession)
		})

		t.Run("That do not match the attempted login credentials should prompt the user to continue", func(t *testing.T) {
			for _, tc := range []struct {
				description     string
				confirmAnswer   string
				expectedUser    auth.User
				expectedSession auth.Session
			}{
				{
					description:     "And do nothing if the user does not want to proceed",
					confirmAnswer:   "n",
					expectedUser:    auth.User{"existingUser", "existing-password"},
					expectedSession: auth.Session{"existingAccessToken", "existingRefreshToken"},
				},
				{
					description:     "And save a new session if the user does want to proceed",
					confirmAnswer:   "y",
					expectedUser:    auth.User{"newUser", "new-password"},
					expectedSession: auth.Session{"newAccessToken", "newRefreshToken"},
				},
			} {
				t.Run(tc.description, func(t *testing.T) {
					profile, realmClient, teardown := setup(t)
					defer teardown()

					cmd := &Command{
						realmClient: realmClient,
						inputs: inputs{
							PublicAPIKey:  "newUser",
							PrivateAPIKey: "new-password",
						},
					}

					_, console, _, ui, consoleErr := mock.NewVT10XConsole()
					assert.Nil(t, consoleErr)

					doneCh := make(chan (struct{}))
					go func() {
						defer close(doneCh)
						console.ExpectString("This action will terminate the existing session for user: existingUser (********-password), would you like to proceed?")
						console.SendLine(tc.confirmAnswer)
						console.ExpectEOF()
					}()

					err := cmd.Handler(profile, ui)
					assert.Nil(t, err)

					assert.Nil(t, console.Tty().Close())
					<-doneCh

					assert.Equal(t, tc.expectedUser, profile.User())
					assert.Equal(t, tc.expectedSession, profile.Session())
					ensureProfileContents(t, profile, tc.expectedUser, tc.expectedSession)
				})
			}
		})
	})
}

func TestLoginFeedback(t *testing.T) {
	t.Run("Feedback should print a message that login was successful", func(t *testing.T) {
		out, ui := mock.NewUI()

		cmd := &Command{}

		err := cmd.Feedback(nil, ui)
		assert.Nil(t, err)

		assert.Equal(t, "01:23:45 UTC INFO  Successfully logged in\n", out.String())
	})
}

func ensureProfileContents(t *testing.T, profile *cli.Profile, user auth.User, session auth.Session) {
	contents, err := ioutil.ReadFile(profile.Path())
	assert.Nil(t, err)
	assert.True(t, strings.Contains(string(contents), fmt.Sprintf(`%s:
  access_token: %s
  private_api_key: %s
  public_api_key: %s
  refresh_token: %s
`, profile.Name, session.AccessToken, user.PrivateAPIKey, user.PublicAPIKey, session.RefreshToken)), "profile must contain the expected contents")
}
