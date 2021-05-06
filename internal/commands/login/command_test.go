package login

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
)

func TestLoginHandler(t *testing.T) {
	t.Run("with no existing credentials handler should save new credentials", func(t *testing.T) {
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

		cmd := &Command{inputs{
			PublicAPIKey:  "publicAPIKey",
			PrivateAPIKey: "privateAPIKey",
		}}

		_, ui := mock.NewUI()

		assert.Nil(t, cmd.Handler(profile, ui, cli.Clients{Realm: realmClient}))

		expectedUser := user.Credentials{"publicAPIKey", "privateAPIKey"}
		expectedSession := user.Session{"accessToken", "refreshToken"}

		assert.Equal(t, expectedUser, profile.Credentials())
		assert.Equal(t, expectedSession, profile.Session())

		ensureProfileContents(t, profile, expectedUser, expectedSession)
	})

	t.Run("with existing credentials", func(t *testing.T) {
		setup := func(t *testing.T) (*user.Profile, realm.Client, func()) {
			tmpDir, teardownTmpDir, tmpDirErr := u.NewTempDir("home")
			assert.Nil(t, tmpDirErr)

			_, teardownHomeDir := u.SetupHomeDir(tmpDir)

			profile := mock.NewProfile(t)

			profile.SetCredentials(user.Credentials{"existingUser", "existing-password"})
			profile.SetSession(user.Session{"existingAccessToken", "existingRefreshToken"})
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

		t.Run("that match the attempted login credentials handler should refresh the existing session", func(t *testing.T) {
			profile, realmClient, teardown := setup(t)
			defer teardown()

			cmd := &Command{inputs{
				PublicAPIKey:  "existingUser",
				PrivateAPIKey: "existing-password",
			}}

			_, ui := mock.NewUI()

			assert.Nil(t, cmd.Handler(profile, ui, cli.Clients{Realm: realmClient}))

			expectedUser := user.Credentials{"existingUser", "existing-password"}
			expectedSession := user.Session{"newAccessToken", "newRefreshToken"}

			assert.Equal(t, expectedUser, profile.Credentials())
			assert.Equal(t, expectedSession, profile.Session())

			ensureProfileContents(t, profile, expectedUser, expectedSession)
		})

		t.Run("that do not match the attempted login credentials should prompt the user to continue", func(t *testing.T) {
			for _, tc := range []struct {
				description     string
				confirmAnswer   string
				expectedUser    user.Credentials
				expectedSession user.Session
			}{
				{
					description:     "And do nothing if the user does not want to proceed",
					confirmAnswer:   "n",
					expectedUser:    user.Credentials{"existingUser", "existing-password"},
					expectedSession: user.Session{"existingAccessToken", "existingRefreshToken"},
				},
				{
					description:     "And save a new session if the user does want to proceed",
					confirmAnswer:   "y",
					expectedUser:    user.Credentials{"newUser", "new-password"},
					expectedSession: user.Session{"newAccessToken", "newRefreshToken"},
				},
			} {
				t.Run(tc.description, func(t *testing.T) {
					profile, realmClient, teardown := setup(t)
					defer teardown()

					cmd := &Command{
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

					err := cmd.Handler(profile, ui, cli.Clients{Realm: realmClient})
					assert.Nil(t, err)

					assert.Nil(t, console.Tty().Close())
					<-doneCh

					assert.Equal(t, tc.expectedUser, profile.Credentials())
					assert.Equal(t, tc.expectedSession, profile.Session())
					ensureProfileContents(t, profile, tc.expectedUser, tc.expectedSession)
				})
			}
		})
	})
}

func TestLoginFeedback(t *testing.T) {
	t.Run("should print a message that login was successful", func(t *testing.T) {
		tc := setup(t)
		defer tc.teardown()

		out := new(bytes.Buffer)
		ui := mock.NewUIWithOptions(mock.UIOptions{AutoConfirm: true}, out)

		cmd := &Command{}

		err := cmd.Handler(tc.profile, ui, cli.Clients{Realm: tc.realmClient})
		assert.Nil(t, err)

		assert.Equal(t, "Successfully logged in\n", out.String())
	})
}

func ensureProfileContents(t *testing.T, profile *user.Profile, user user.Credentials, session user.Session) {
	contents, err := ioutil.ReadFile(profile.Path())
	assert.Nil(t, err)
	assert.True(t, strings.Contains(string(contents), fmt.Sprintf(`%s:
  access_token: %s
  private_api_key: %s
  public_api_key: %s
  refresh_token: %s
`, profile.Name, session.AccessToken, user.PrivateAPIKey, user.PublicAPIKey, session.RefreshToken)), "profile must contain the expected contents")
}

type testContext struct {
	profile     *user.Profile
	realmClient realm.Client
	teardown    func()
}

func setup(t *testing.T) testContext {
	t.Helper()

	tmpDir, teardownTmpDir, err := u.NewTempDir("home")
	assert.Nil(t, err)

	_, teardownHomeDir := u.SetupHomeDir(tmpDir)

	profile := mock.NewProfile(t)

	profile.SetCredentials(user.Credentials{"existingUser", "existing-password"})
	profile.SetSession(user.Session{"existingAccessToken", "existingRefreshToken"})
	assert.Nil(t, profile.Save())

	realmClient := mock.RealmClient{}
	realmClient.AuthenticateFn = func(publicAPIKey, privateAPIKey string) (realm.Session, error) {
		return realm.Session{
			AccessToken:  "newAccessToken",
			RefreshToken: "newRefreshToken",
		}, nil
	}

	teardown := func() {
		teardownHomeDir()
		teardownTmpDir()
	}

	return testContext{profile, realmClient, teardown}
}
