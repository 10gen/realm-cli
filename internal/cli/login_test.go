package cli

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/mock"

	expect "github.com/Netflix/go-expect"
	"github.com/google/go-cmp/cmp"
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
		u.MustMatch(t, cmp.Diff(nil, err))

		u.MustNotBeNil(t, cmd.realmClient)
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
				u.MustMatch(t, cmp.Diff("username", cmd.publicAPIKey))
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
				u.MustMatch(t, cmp.Diff("password", cmd.privateAPIKey))
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
				u.MustMatch(t, cmp.Diff("username", cmd.publicAPIKey))
				u.MustMatch(t, cmp.Diff("password", cmd.privateAPIKey))
			},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			out := new(bytes.Buffer)
			console, _, ui, consoleErr := mock.NewVT10XConsole(mock.UIOptions{}, out)
			u.MustMatch(t, cmp.Diff(nil, consoleErr))
			defer console.Close()

			profile, profileErr := NewProfile(primitive.NewObjectID().Hex())
			u.MustMatch(t, cmp.Diff(nil, profileErr))

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
			u.MustMatch(t, cmp.Diff(nil, err))

			console.Tty().Close() // flush the writers
			<-doneCh              // wait for procedure to complete

			tc.test(t, cmd)
		})
	}
}

func TestLoginHandler(t *testing.T) {
	t.Run("With no existing credentials handler should save new credentials", func(t *testing.T) {
		tmpDir, teardownTmpDir, tmpDirErr := u.NewTempDir("home")
		u.MustMatch(t, cmp.Diff(nil, tmpDirErr))
		defer teardownTmpDir()

		_, teardownHomeDir := u.SetupHomeDir(tmpDir)
		defer teardownHomeDir()

		profile, profileErr := NewProfile(primitive.NewObjectID().Hex())
		u.MustMatch(t, cmp.Diff(nil, profileErr))

		_, statErr := os.Stat(profile.path())
		u.MustNotBeNil(t, statErr)
		u.MustMatch(t, cmp.Diff(true, os.IsNotExist(statErr)))

		realmClient := mock.RealmClient{}
		realmClient.AuthenticateFn = func(publicAPIKey, privateAPIKey string) (realm.AuthResponse, error) {
			return realm.AuthResponse{
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

		u.MustMatch(t, cmp.Diff(nil, cmd.Handler(profile, ui, nil)))

		expectedUser := User{"publicAPIKey", "privateAPIKey"}
		expectedSession := Session{"accessToken", "refreshToken"}

		u.MustMatch(t, cmp.Diff(expectedUser, profile.GetUser()))
		u.MustMatch(t, cmp.Diff(expectedSession, profile.GetSession()))
		ensureProfileContents(t, profile, expectedUser, expectedSession)
	})

	t.Run("With existing credentials", func(t *testing.T) {
		setup := func(t *testing.T) (*Profile, realm.Client, func()) {
			tmpDir, teardownTmpDir, tmpDirErr := u.NewTempDir("home")
			u.MustMatch(t, cmp.Diff(nil, tmpDirErr))

			_, teardownHomeDir := u.SetupHomeDir(tmpDir)

			profile, profileErr := NewProfile(primitive.NewObjectID().Hex())
			u.MustMatch(t, cmp.Diff(nil, profileErr))

			profile.SetUser("existingUser", "existing-password")
			profile.SetSession("existingAccessToken", "existingRefreshToken")
			u.MustMatch(t, cmp.Diff(nil, profile.Save()))

			realmClient := mock.RealmClient{}
			realmClient.AuthenticateFn = func(publicAPIKey, privateAPIKey string) (realm.AuthResponse, error) {
				return realm.AuthResponse{
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

			u.MustMatch(t, cmp.Diff(nil, cmd.Handler(profile, ui, nil)))

			expectedUser := User{"existingUser", "existing-password"}
			expectedSession := Session{"newAccessToken", "newRefreshToken"}

			u.MustMatch(t, cmp.Diff(expectedUser, profile.GetUser()))
			u.MustMatch(t, cmp.Diff(expectedSession, profile.GetSession()))
			ensureProfileContents(t, profile, expectedUser, expectedSession)
		})

		t.Run("That do not match the attempted login credentials should prompt the user to continue", func(t *testing.T) {
			for _, tc := range []struct {
				description     string
				confirmAnswer   string
				expectedUser    User
				expectedSession Session
			}{
				{
					description:     "And do nothing if the user does not want to proceed",
					confirmAnswer:   "n",
					expectedUser:    User{"existingUser", "existing-password"},
					expectedSession: Session{"existingAccessToken", "existingRefreshToken"},
				},
				{
					description:     "And save a new session if the user does want to proceed",
					confirmAnswer:   "y",
					expectedUser:    User{"newUser", "new-password"},
					expectedSession: Session{"newAccessToken", "newRefreshToken"},
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
					u.MustMatch(t, cmp.Diff(nil, consoleErr))

					doneCh := make(chan (struct{}))
					go func() {
						defer close(doneCh)
						console.ExpectString("This action will terminate the existing session for user: existingUser (********-password), would you like to proceed?")
						console.SendLine(tc.confirmAnswer)
						console.ExpectEOF()
					}()

					err := cmd.Handler(profile, ui, nil)
					u.MustMatch(t, cmp.Diff(nil, err))

					u.MustMatch(t, cmp.Diff(nil, console.Tty().Close()))
					<-doneCh

					u.MustMatch(t, cmp.Diff(tc.expectedUser, profile.GetUser()))
					u.MustMatch(t, cmp.Diff(tc.expectedSession, profile.GetSession()))
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
		u.MustMatch(t, cmp.Diff(nil, err))

		u.MustMatch(t, cmp.Diff("Successfully logged in.\n", out.String()))
	})
}

func ensureProfileContents(t *testing.T, profile *Profile, user User, session Session) {
	t.Log("ensure profile has the expected contents")
	contents, err := ioutil.ReadFile(profile.path())
	u.MustMatch(t, cmp.Diff(nil, err))
	u.MustContainSubstring(t, string(contents), fmt.Sprintf(`%s:
  access_token: %s
  private_api_key: %s
  public_api_key: %s
  refresh_token: %s
`, profile.Name, session.AccessToken, user.PrivateAPIKey, user.PublicAPIKey, session.RefreshToken))
}
