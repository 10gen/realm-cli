package commands

import (
	"errors"
	"testing"

	"github.com/10gen/stitch-cli/models"
	"github.com/10gen/stitch-cli/secrets"
	"github.com/10gen/stitch-cli/user"
	u "github.com/10gen/stitch-cli/utils/test"

	"github.com/mitchellh/cli"
	gc "github.com/smartystreets/goconvey/convey"
)

func setUpBasicSecretsCommand(
	listSecretsFn func(groupID, appID string) ([]secrets.Secret, error),
	addSecretFn func(groupID, appID string, secret secrets.Secret) error,
	updateSecretByIDFn func(groupID, appID, secretID, secretValue string) error,
	updateSecretByNameFn func(groupID, appID, secretName, secretValue string) error,
	removeSecretByIDFn func(groupID, appID, secretID string) error,
	removeSecretByNameFn func(groupID, appID, secretName string) error,
) (*SecretsCommand, *cli.MockUi) {
	mockUI := cli.NewMockUi()
	cmd, err := NewSecretsCommandFactory(mockUI)()
	if err != nil {
		panic(err)
	}

	secretsCommand := cmd.(*SecretsCommand)
	secretsCommand.storage = u.NewEmptyStorage()

	mockStitchClient := &u.MockStitchClient{
		ListSecretsFn:        listSecretsFn,
		AddSecretFn:          addSecretFn,
		UpdateSecretByIDFn:   updateSecretByIDFn,
		UpdateSecretByNameFn: updateSecretByNameFn,
		RemoveSecretByIDFn:   removeSecretByIDFn,
		RemoveSecretByNameFn: removeSecretByNameFn,
		FetchAppByClientAppIDFn: func(clientAppID string) (*models.App, error) {
			return &models.App{
				GroupID: "group-id",
				ID:      "app-id",
			}, nil
		},
	}
	secretsCommand.stitchClient = mockStitchClient
	return secretsCommand, mockUI
}

func TestSecretsCommand(t *testing.T) {
	validListArgs := []string{"--app-id=my-app-abcdef"}
	validAddArgs := []string{"--app-id=my-app-abcdef", "--name=foo", "--value=bar"}
	validUpdateByIDArgs := []string{"--app-id=my-app-abcdef", "--secret-id=thisisanid", "--value=newvalue"}
	validUpdateByNameArgs := []string{"--app-id=my-app-abcdef", "--secret-name=thisisaname", "--value=newvalue"}
	validRemoveByIDArgs := []string{"--app-id=my-app-abcdef", "--secret-id=thisisanid"}
	validRemoveByNameArgs := []string{"--app-id=my-app-abcdef", "--secret-name=thisisaname"}

	t.Run("listing a secret should require the user to be logged in", func(t *testing.T) {
		secretsCommand, mockUI := setUpBasicSecretsCommand(nil, nil, nil, nil, nil, nil)
		exitCode := secretsCommand.Run(append([]string{"list"}, validListArgs...))
		u.So(t, exitCode, gc.ShouldEqual, 1)

		u.So(t, mockUI.ErrorWriter.String(), gc.ShouldContainSubstring, user.ErrNotLoggedIn.Error())
	})

	t.Run("adding a secret should require the user to be logged in", func(t *testing.T) {
		secretsCommand, mockUI := setUpBasicSecretsCommand(nil, nil, nil, nil, nil, nil)
		exitCode := secretsCommand.Run(append([]string{"add"}, validAddArgs...))
		u.So(t, exitCode, gc.ShouldEqual, 1)

		u.So(t, mockUI.ErrorWriter.String(), gc.ShouldContainSubstring, user.ErrNotLoggedIn.Error())
	})

	t.Run("updating a secret should require the user to be logged in", func(t *testing.T) {
		secretsCommand, mockUI := setUpBasicSecretsCommand(nil, nil, nil, nil, nil, nil)
		exitCode := secretsCommand.Run(append([]string{"update"}, validUpdateByIDArgs...))
		u.So(t, exitCode, gc.ShouldEqual, 1)

		u.So(t, mockUI.ErrorWriter.String(), gc.ShouldContainSubstring, user.ErrNotLoggedIn.Error())
	})

	t.Run("removing a secret should require the user to be logged in", func(t *testing.T) {
		secretsCommand, mockUI := setUpBasicSecretsCommand(nil, nil, nil, nil, nil, nil)
		exitCode := secretsCommand.Run(append([]string{"remove"}, validRemoveByIDArgs...))
		u.So(t, exitCode, gc.ShouldEqual, 1)

		u.So(t, mockUI.ErrorWriter.String(), gc.ShouldContainSubstring, user.ErrNotLoggedIn.Error())
	})

	t.Run("when the user is logged in", func(t *testing.T) {
		setup := func(
			listSecretsFn func(appID, groupID string) ([]secrets.Secret, error),
			addSecretsFn func(appID, groupID string, secret secrets.Secret) error,
			updateSecretByIDFn func(appID, groupID, secretID, secretValue string) error,
			updateSecretsByNameFn func(appID, groupID, secretName, secretValue string) error,
			removeSecretByIDFn func(appID, groupID, secretID string) error,
			removeSecretsByNameFn func(appID, groupID, secretName string) error,
		) (*SecretsCommand, *cli.MockUi) {
			secretsCommand, mockUI := setUpBasicSecretsCommand(
				listSecretsFn,
				addSecretsFn,
				updateSecretByIDFn,
				updateSecretsByNameFn,
				removeSecretByIDFn,
				removeSecretsByNameFn,
			)

			secretsCommand.user = &user.User{
				APIKey:      "my-api-key",
				AccessToken: u.GenerateValidAccessToken(),
			}

			return secretsCommand, mockUI
		}

		t.Run("it fails if there is no sub command", func(t *testing.T) {
			secretsCommand, _ := setup(nil, nil, nil, nil, nil, nil)
			exitCode := secretsCommand.Run(append([]string{}, validAddArgs...))
			u.So(t, exitCode, gc.ShouldEqual, 127)
		})

		t.Run("it fails if there is an invalid sub command", func(t *testing.T) {
			secretsCommand, _ := setup(nil, nil, nil, nil, nil, nil)
			exitCode := secretsCommand.Run(append([]string{"invalid"}, validAddArgs...))
			u.So(t, exitCode, gc.ShouldEqual, 127)
		})

		t.Run("adding a secret fails if the secret name is missing", func(t *testing.T) {
			secretsCommand, mockUI := setup(nil, nil, nil, nil, nil, nil)
			exitCode := secretsCommand.Run(append([]string{"add", "--app-id=my-app-abcdef", "--value=bar"}))
			u.So(t, exitCode, gc.ShouldEqual, 1)
			u.So(t, mockUI.ErrorWriter.String(), gc.ShouldContainSubstring, "is required")
		})

		t.Run("adding a secret fails if the secret value is missing", func(t *testing.T) {
			secretsCommand, mockUI := setup(nil, nil, nil, nil, nil, nil)
			exitCode := secretsCommand.Run(append([]string{"add", "--app-id=my-app-abcdef", "--name=foo"}))
			u.So(t, exitCode, gc.ShouldEqual, 1)
			u.So(t, mockUI.ErrorWriter.String(), gc.ShouldContainSubstring, "is required")
		})

		t.Run("removing a secret fails if the secret id and name are missing", func(t *testing.T) {
			secretsCommand, mockUI := setup(nil, nil, nil, nil, nil, nil)
			exitCode := secretsCommand.Run(append([]string{"remove", "--app-id=my-app-abcdef"}))
			u.So(t, exitCode, gc.ShouldEqual, 1)
			u.So(t, mockUI.ErrorWriter.String(), gc.ShouldContainSubstring, "is required")
		})

		t.Run("listing a secret fails if the listing method fails", func(t *testing.T) {
			secretsCommand, mockUI := setup(func(appID, groupID string) ([]secrets.Secret, error) {
				return nil, errors.New("oopsies")
			}, nil, nil, nil, nil, nil)
			exitCode := secretsCommand.Run(append([]string{"list"}, validListArgs...))
			u.So(t, exitCode, gc.ShouldEqual, 1)
			u.So(t, mockUI.ErrorWriter.String(), gc.ShouldContainSubstring, "oopsies")
		})

		t.Run("adding a secret fails if adding the secret fails", func(t *testing.T) {
			secretsCommand, mockUI := setup(nil, func(appID, groupID string, secret secrets.Secret) error {
				return errors.New("oopsies")
			}, nil, nil, nil, nil)
			exitCode := secretsCommand.Run(append([]string{"add"}, validAddArgs...))
			u.So(t, exitCode, gc.ShouldEqual, 1)
			u.So(t, mockUI.ErrorWriter.String(), gc.ShouldContainSubstring, "oopsies")
		})

		t.Run("updating a secret by id fails if removing the secret fails", func(t *testing.T) {
			secretsCommand, mockUI := setup(nil, nil, func(appID, groupID, secretID, secretValue string) error {
				return errors.New("oopsies")
			}, nil, nil, nil)
			exitCode := secretsCommand.Run(append([]string{"update"}, validUpdateByIDArgs...))
			u.So(t, exitCode, gc.ShouldEqual, 1)
			u.So(t, mockUI.ErrorWriter.String(), gc.ShouldContainSubstring, "oopsies")
		})

		t.Run("updating a secret by name fails if removing the secret fails", func(t *testing.T) {
			secretsCommand, mockUI := setup(nil, nil, nil, func(appID, groupID, secretName, secretValue string) error {
				return errors.New("oopsies")
			}, nil, nil)
			exitCode := secretsCommand.Run(append([]string{"update"}, validUpdateByNameArgs...))
			u.So(t, exitCode, gc.ShouldEqual, 1)
			u.So(t, mockUI.ErrorWriter.String(), gc.ShouldContainSubstring, "oopsies")
		})

		t.Run("removing a secret by id fails if removing the secret fails", func(t *testing.T) {
			secretsCommand, mockUI := setup(nil, nil, nil, nil, func(appID, groupID, secretID string) error {
				return errors.New("oopsies")
			}, nil)
			exitCode := secretsCommand.Run(append([]string{"remove"}, validRemoveByIDArgs...))
			u.So(t, exitCode, gc.ShouldEqual, 1)
			u.So(t, mockUI.ErrorWriter.String(), gc.ShouldContainSubstring, "oopsies")
		})

		t.Run("removing a secret by name fails if removing the secret fails", func(t *testing.T) {
			secretsCommand, mockUI := setup(nil, nil, nil, nil, nil, func(appID, groupID, secretID string) error {
				return errors.New("oopsies")
			})
			exitCode := secretsCommand.Run(append([]string{"remove"}, validRemoveByNameArgs...))
			u.So(t, exitCode, gc.ShouldEqual, 1)
			u.So(t, mockUI.ErrorWriter.String(), gc.ShouldContainSubstring, "oopsies")
		})

		t.Run("it passes the correct flags to AddSecret", func(t *testing.T) {
			var secretName string
			var secretValue string
			secretsCommand, _ := setup(nil, func(appID, groupID string, secret secrets.Secret) error {
				secretName = secret.Name
				secretValue = secret.Value
				return nil
			}, nil, nil, nil, nil)
			exitCode := secretsCommand.Run(append([]string{"add"}, validAddArgs...))
			u.So(t, exitCode, gc.ShouldEqual, 0)
			u.So(t, secretName, gc.ShouldEqual, "foo")
			u.So(t, secretValue, gc.ShouldEqual, "bar")
		})

		t.Run("it passes the correct flags to UpdateSecretByID", func(t *testing.T) {
			var secretID string
			secretsCommand, _ := setup(nil, nil, func(appID, groupID, id, value string) error {
				secretID = id
				return nil
			}, nil, nil, nil)
			exitCode := secretsCommand.Run(append([]string{"update"}, validUpdateByIDArgs...))
			u.So(t, exitCode, gc.ShouldEqual, 0)
			u.So(t, secretID, gc.ShouldEqual, "thisisanid")
		})

		t.Run("it passes the correct flags to UpdateSecretByName", func(t *testing.T) {
			var secretName string
			secretsCommand, _ := setup(nil, nil, nil, func(appID, groupID, name, value string) error {
				secretName = name
				return nil
			}, nil, nil)
			exitCode := secretsCommand.Run(append([]string{"update"}, validUpdateByNameArgs...))
			u.So(t, exitCode, gc.ShouldEqual, 0)
			u.So(t, secretName, gc.ShouldEqual, "thisisaname")
		})

		t.Run("it passes the correct flags to RemoveSecretByID", func(t *testing.T) {
			var secretID string
			secretsCommand, _ := setup(nil, nil, nil, nil, func(appID, groupID, id string) error {
				secretID = id
				return nil
			}, nil)
			exitCode := secretsCommand.Run(append([]string{"remove"}, validRemoveByIDArgs...))
			u.So(t, exitCode, gc.ShouldEqual, 0)
			u.So(t, secretID, gc.ShouldEqual, "thisisanid")
		})

		t.Run("it passes the correct flags to RemoveSecretByName", func(t *testing.T) {
			var secretName string
			secretsCommand, _ := setup(nil, nil, nil, nil, nil, func(appID, groupID, name string) error {
				secretName = name
				return nil
			})
			exitCode := secretsCommand.Run(append([]string{"remove"}, validRemoveByNameArgs...))
			u.So(t, exitCode, gc.ShouldEqual, 0)
			u.So(t, secretName, gc.ShouldEqual, "thisisaname")
		})

		t.Run("listing secrets works when there are none", func(t *testing.T) {
			secretsCommand, mockUI := setup(nil, nil, nil, nil, nil, nil)
			exitCode := secretsCommand.Run(append([]string{"list"}, validListArgs...))
			u.So(t, exitCode, gc.ShouldEqual, 0)
			u.So(t, mockUI.OutputWriter.String(), gc.ShouldContainSubstring, "No secrets found for this app")
		})

		t.Run("listing secrets works", func(t *testing.T) {
			secretsCommand, mockUI := setup(func(appID, groupID string) ([]secrets.Secret, error) {
				return []secrets.Secret{{ID: "123", Name: "hello", Value: "there"}}, nil
			}, nil, nil, nil, nil, nil)
			exitCode := secretsCommand.Run(append([]string{"list"}, validListArgs...))
			u.So(t, exitCode, gc.ShouldEqual, 0)
			u.So(t, mockUI.OutputWriter.String(), gc.ShouldContainSubstring, "123")
			u.So(t, mockUI.OutputWriter.String(), gc.ShouldContainSubstring, "hello")
			u.So(t, mockUI.OutputWriter.String(), gc.ShouldNotContainSubstring, "there")
		})

		t.Run("adding a secret works", func(t *testing.T) {
			secretsCommand, mockUI := setup(nil, nil, nil, nil, nil, nil)
			exitCode := secretsCommand.Run(append([]string{"add"}, validAddArgs...))
			u.So(t, exitCode, gc.ShouldEqual, 0)
			u.So(t, mockUI.OutputWriter.String(), gc.ShouldContainSubstring, "New secret created")
		})

		t.Run("updating a secret by id works", func(t *testing.T) {
			secretsCommand, mockUI := setup(nil, nil, nil, nil, nil, nil)
			exitCode := secretsCommand.Run(append([]string{"update"}, validUpdateByIDArgs...))
			u.So(t, exitCode, gc.ShouldEqual, 0)
			u.So(t, mockUI.OutputWriter.String(), gc.ShouldContainSubstring, "Secret updated: thisisanid")
		})

		t.Run("updating a secret by name works", func(t *testing.T) {
			secretsCommand, mockUI := setup(nil, nil, nil, nil, nil, nil)
			exitCode := secretsCommand.Run(append([]string{"update"}, validUpdateByNameArgs...))
			u.So(t, exitCode, gc.ShouldEqual, 0)
			u.So(t, mockUI.OutputWriter.String(), gc.ShouldContainSubstring, "Secret updated: thisisaname")
		})

		t.Run("removing a secret by id works", func(t *testing.T) {
			secretsCommand, mockUI := setup(nil, nil, nil, nil, nil, nil)
			exitCode := secretsCommand.Run(append([]string{"remove"}, validRemoveByIDArgs...))
			u.So(t, exitCode, gc.ShouldEqual, 0)
			u.So(t, mockUI.OutputWriter.String(), gc.ShouldContainSubstring, "Secret removed: thisisanid")
		})

		t.Run("removing a secret by name works", func(t *testing.T) {
			secretsCommand, mockUI := setup(nil, nil, nil, nil, nil, nil)
			exitCode := secretsCommand.Run(append([]string{"remove"}, validRemoveByNameArgs...))
			u.So(t, exitCode, gc.ShouldEqual, 0)
			u.So(t, mockUI.OutputWriter.String(), gc.ShouldContainSubstring, "Secret removed: thisisaname")
		})
	})
}
