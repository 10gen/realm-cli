package commands

import (
	"errors"
	"testing"

	"github.com/10gen/realm-cli/utils/telemetry"

	"github.com/10gen/realm-cli/models"
	"github.com/10gen/realm-cli/secrets"
	"github.com/10gen/realm-cli/user"
	u "github.com/10gen/realm-cli/utils/test"

	"github.com/mitchellh/cli"
	gc "github.com/smartystreets/goconvey/convey"
)

type mockClientFunctions struct {
	listSecretsFn        func(groupID, appID string) ([]secrets.Secret, error)
	addSecretFn          func(groupID, appID string, secret secrets.Secret) error
	updateSecretByIDFn   func(groupID, appID, secretID, secretValue string) error
	updateSecretByNameFn func(groupID, appID, secretName, secretValue string) error
	removeSecretByIDFn   func(groupID, appID, secretID string) error
	removeSecretByNameFn func(groupID, appID, secretName string) error
}

func setUpBasicSecretsCommand(
	baseCommand *SecretsBaseCommand,
	mcf *mockClientFunctions,
) {
	baseCommand.storage = u.NewEmptyStorage()

	if mcf == nil {
		mcf = &mockClientFunctions{}
	}

	mockRealmClient := &u.MockRealmClient{
		ListSecretsFn:        mcf.listSecretsFn,
		AddSecretFn:          mcf.addSecretFn,
		UpdateSecretByIDFn:   mcf.updateSecretByIDFn,
		UpdateSecretByNameFn: mcf.updateSecretByNameFn,
		RemoveSecretByIDFn:   mcf.removeSecretByIDFn,
		RemoveSecretByNameFn: mcf.removeSecretByNameFn,
		FetchAppByClientAppIDFn: func(clientAppID string) (*models.App, error) {
			return &models.App{
				GroupID: "group-id",
				ID:      "app-id",
			}, nil
		},
	}
	baseCommand.realmClient = mockRealmClient
}

func TestSecretsCommand(t *testing.T) {
	mockService := &telemetry.Service{}
	validListArgs := []string{"--app-id=my-app-abcdef"}
	validAddArgs := []string{"--app-id=my-app-abcdef", "--name=foo", "--value=bar"}
	validUpdateByIDArgs := []string{"--app-id=my-app-abcdef", "--id=thisisanid", "--value=newvalue"}
	validUpdateByIDDeprecatedArgs := []string{"--app-id=my-app-abcdef", "--secret-id=thisisanid", "--value=newvalue"}
	validUpdateByNameArgs := []string{"--app-id=my-app-abcdef", "--name=thisisaname", "--value=newvalue"}
	validUpdateByNameDeprecatedArgs := []string{"--app-id=my-app-abcdef", "--secret-name=thisisaname", "--value=newvalue"}
	validRemoveByIDArgs := []string{"--app-id=my-app-abcdef", "--id=thisisanid"}
	validRemoveByIDDeprecatedArgs := []string{"--app-id=my-app-abcdef", "--secret-id=thisisanid"}
	validRemoveByNameArgs := []string{"--app-id=my-app-abcdef", "--name=thisisaname"}
	validRemoveByNameDeprecatedArgs := []string{"--app-id=my-app-abcdef", "--secret-name=thisisaname"}

	t.Run("listing a secret should require the user to be logged in", func(t *testing.T) {
		mockUI := cli.NewMockUi()
		cmd, err := NewSecretsListCommandFactory(mockUI, mockService)()
		u.So(t, err, gc.ShouldBeNil)

		listCommand := cmd.(*SecretsListCommand)
		setUpBasicSecretsCommand(listCommand.SecretsBaseCommand, nil)

		exitCode := listCommand.Run(validListArgs)
		u.So(t, exitCode, gc.ShouldEqual, 1)

		u.So(t, mockUI.ErrorWriter.String(), gc.ShouldContainSubstring, user.ErrNotLoggedIn.Error())
	})

	t.Run("adding a secret should require the user to be logged in", func(t *testing.T) {
		mockUI := cli.NewMockUi()
		cmd, err := NewSecretsAddCommandFactory(mockUI, mockService)()
		u.So(t, err, gc.ShouldBeNil)

		addCommand := cmd.(*SecretsAddCommand)
		setUpBasicSecretsCommand(addCommand.SecretsBaseCommand, nil)

		exitCode := addCommand.Run(validAddArgs)
		u.So(t, exitCode, gc.ShouldEqual, 1)

		u.So(t, mockUI.ErrorWriter.String(), gc.ShouldContainSubstring, user.ErrNotLoggedIn.Error())
	})

	for _, tc := range []struct {
		description string
		args        []string
	}{
		{
			description: "updating a secret args should require the user to be logged in",
			args:        validUpdateByIDArgs,
		},
		{
			description: "updating a secret with deprecated args should require the user to be logged in",
			args:        validUpdateByIDDeprecatedArgs,
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			mockUI := cli.NewMockUi()
			cmd, err := NewSecretsUpdateCommandFactory(mockUI, mockService)()
			u.So(t, err, gc.ShouldBeNil)

			updateCommand := cmd.(*SecretsUpdateCommand)
			setUpBasicSecretsCommand(updateCommand.SecretsBaseCommand, nil)

			exitCode := updateCommand.Run(tc.args)
			u.So(t, exitCode, gc.ShouldEqual, 1)

			u.So(t, mockUI.ErrorWriter.String(), gc.ShouldContainSubstring, user.ErrNotLoggedIn.Error())
		})
	}

	for _, tc := range []struct {
		description string
		args        []string
	}{
		{
			description: "removing a secret should require the user to be logged in",
			args:        validRemoveByIDArgs,
		},
		{
			description: "removing a secret with deprecated args should require the user to be logged in",
			args:        validRemoveByIDDeprecatedArgs,
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			mockUI := cli.NewMockUi()
			cmd, err := NewSecretsRemoveCommandFactory(mockUI, mockService)()
			u.So(t, err, gc.ShouldBeNil)

			removeCommand := cmd.(*SecretsRemoveCommand)
			setUpBasicSecretsCommand(removeCommand.SecretsBaseCommand, nil)

			exitCode := removeCommand.Run(tc.args)
			u.So(t, exitCode, gc.ShouldEqual, 1)

			u.So(t, mockUI.ErrorWriter.String(), gc.ShouldContainSubstring, user.ErrNotLoggedIn.Error())
		})
	}

	t.Run("when the user is logged in", func(t *testing.T) {
		setup := func(
			baseCommand *SecretsBaseCommand,
			mcf *mockClientFunctions,
		) {
			setUpBasicSecretsCommand(baseCommand, mcf)
			baseCommand.user = &user.User{
				APIKey:      "my-api-key",
				AccessToken: u.GenerateValidAccessToken(),
			}
		}

		t.Run("adding a secret fails if the secret name is missing", func(t *testing.T) {
			mockUI := cli.NewMockUi()
			cmd, err := NewSecretsAddCommandFactory(mockUI, mockService)()
			u.So(t, err, gc.ShouldBeNil)

			addCommand := cmd.(*SecretsAddCommand)
			setup(addCommand.SecretsBaseCommand, nil)

			exitCode := addCommand.Run([]string{"--app-id=my-app-abcdef", "--value=bar"})
			u.So(t, exitCode, gc.ShouldEqual, 1)
			u.So(t, mockUI.ErrorWriter.String(), gc.ShouldContainSubstring, "is required")
		})

		t.Run("adding a secret fails if the secret value is missing", func(t *testing.T) {
			mockUI := cli.NewMockUi()
			cmd, err := NewSecretsAddCommandFactory(mockUI, mockService)()
			u.So(t, err, gc.ShouldBeNil)

			addCommand := cmd.(*SecretsAddCommand)
			setup(addCommand.SecretsBaseCommand, nil)

			exitCode := addCommand.Run([]string{"--app-id=my-app-abcdef", "--name=foo"})
			u.So(t, exitCode, gc.ShouldEqual, 1)
			u.So(t, mockUI.ErrorWriter.String(), gc.ShouldContainSubstring, "is required")
		})

		t.Run("removing a secret fails if the secret id and name are missing", func(t *testing.T) {
			mockUI := cli.NewMockUi()
			cmd, err := NewSecretsRemoveCommandFactory(mockUI, mockService)()
			u.So(t, err, gc.ShouldBeNil)

			removeCommand := cmd.(*SecretsRemoveCommand)
			setup(removeCommand.SecretsBaseCommand, nil)

			exitCode := removeCommand.Run([]string{"--app-id=my-app-abcdef"})
			u.So(t, exitCode, gc.ShouldEqual, 1)
			u.So(t, mockUI.ErrorWriter.String(), gc.ShouldContainSubstring, "is required")
		})

		t.Run("listing a secret fails if the listing method fails", func(t *testing.T) {
			mockUI := cli.NewMockUi()
			cmd, err := NewSecretsListCommandFactory(mockUI, mockService)()
			u.So(t, err, gc.ShouldBeNil)

			listCommand := cmd.(*SecretsListCommand)
			setup(
				listCommand.SecretsBaseCommand,
				&mockClientFunctions{
					listSecretsFn: func(appID, groupID string) ([]secrets.Secret, error) {
						return nil, errors.New("oopsies")
					},
				},
			)

			exitCode := listCommand.Run([]string{"--app-id=my-app-abcdef"})
			u.So(t, exitCode, gc.ShouldEqual, 1)
			u.So(t, mockUI.ErrorWriter.String(), gc.ShouldContainSubstring, "oopsies")
		})

		t.Run("adding a secret fails if adding the secret fails", func(t *testing.T) {
			mockUI := cli.NewMockUi()
			cmd, err := NewSecretsAddCommandFactory(mockUI, mockService)()
			u.So(t, err, gc.ShouldBeNil)

			addCommand := cmd.(*SecretsAddCommand)
			setup(addCommand.SecretsBaseCommand, &mockClientFunctions{
				addSecretFn: func(appID, groupID string, secret secrets.Secret) error {
					return errors.New("oopsies")
				},
			})

			exitCode := addCommand.Run(validAddArgs)
			u.So(t, exitCode, gc.ShouldEqual, 1)
			u.So(t, mockUI.ErrorWriter.String(), gc.ShouldContainSubstring, "oopsies")
		})

		t.Run("updating a secret by id fails if updating the secret fails", func(t *testing.T) {
			mockUI := cli.NewMockUi()
			cmd, err := NewSecretsUpdateCommandFactory(mockUI, mockService)()
			u.So(t, err, gc.ShouldBeNil)

			updateCommand := cmd.(*SecretsUpdateCommand)
			setup(updateCommand.SecretsBaseCommand, &mockClientFunctions{
				updateSecretByIDFn: func(appID, groupID, secretID, secretValue string) error {
					return errors.New("oopsies")
				},
			})
			exitCode := updateCommand.Run(validUpdateByIDArgs)
			u.So(t, exitCode, gc.ShouldEqual, 1)
			u.So(t, mockUI.ErrorWriter.String(), gc.ShouldContainSubstring, "oopsies")
		})

		t.Run("updating a secret by name fails if updating the secret fails", func(t *testing.T) {
			mockUI := cli.NewMockUi()
			cmd, err := NewSecretsUpdateCommandFactory(mockUI, mockService)()
			u.So(t, err, gc.ShouldBeNil)

			updateCommand := cmd.(*SecretsUpdateCommand)
			setup(updateCommand.SecretsBaseCommand, &mockClientFunctions{
				updateSecretByNameFn: func(appID, groupID, secretName, secretValue string) error {
					return errors.New("oopsies")
				},
			})
			exitCode := updateCommand.Run(validUpdateByNameArgs)
			u.So(t, exitCode, gc.ShouldEqual, 1)
			u.So(t, mockUI.ErrorWriter.String(), gc.ShouldContainSubstring, "oopsies")
		})

		t.Run("removing a secret by id fails if removing the secret fails", func(t *testing.T) {
			mockUI := cli.NewMockUi()
			cmd, err := NewSecretsRemoveCommandFactory(mockUI, mockService)()
			u.So(t, err, gc.ShouldBeNil)

			removeCommand := cmd.(*SecretsRemoveCommand)
			setup(removeCommand.SecretsBaseCommand, &mockClientFunctions{
				removeSecretByIDFn: func(appID, groupID, secretID string) error {
					return errors.New("oopsies")
				},
			})
			exitCode := removeCommand.Run(validRemoveByIDArgs)
			u.So(t, exitCode, gc.ShouldEqual, 1)
			u.So(t, mockUI.ErrorWriter.String(), gc.ShouldContainSubstring, "oopsies")
		})

		t.Run("removing a secret by name fails if removing the secret fails", func(t *testing.T) {
			mockUI := cli.NewMockUi()
			cmd, err := NewSecretsRemoveCommandFactory(mockUI, mockService)()
			u.So(t, err, gc.ShouldBeNil)

			removeCommand := cmd.(*SecretsRemoveCommand)
			setup(removeCommand.SecretsBaseCommand, &mockClientFunctions{
				removeSecretByNameFn: func(appID, groupID, secretName string) error {
					return errors.New("oopsies")
				},
			})
			exitCode := removeCommand.Run(validRemoveByNameArgs)
			u.So(t, exitCode, gc.ShouldEqual, 1)
			u.So(t, mockUI.ErrorWriter.String(), gc.ShouldContainSubstring, "oopsies")
		})

		t.Run("it passes the correct flags to AddSecret", func(t *testing.T) {
			mockUI := cli.NewMockUi()
			cmd, err := NewSecretsAddCommandFactory(mockUI, mockService)()
			u.So(t, err, gc.ShouldBeNil)

			var secretName string
			var secretValue string
			addCommand := cmd.(*SecretsAddCommand)
			setup(addCommand.SecretsBaseCommand, &mockClientFunctions{
				addSecretFn: func(appID, groupID string, secret secrets.Secret) error {
					secretName = secret.Name
					secretValue = secret.Value
					return nil
				},
			})
			exitCode := addCommand.Run(validAddArgs)
			u.So(t, exitCode, gc.ShouldEqual, 0)
			u.So(t, secretName, gc.ShouldEqual, "foo")
			u.So(t, secretValue, gc.ShouldEqual, "bar")
		})

		for _, tc := range []struct {
			description string
			args        []string
		}{
			{
				description: "it passes the correct flags to UpdateSecretByID",
				args:        validUpdateByIDArgs,
			},
			{
				description: "it passes the correct flags to UpdateSecretByID with deprecated args",
				args:        validUpdateByIDDeprecatedArgs,
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				mockUI := cli.NewMockUi()
				cmd, err := NewSecretsUpdateCommandFactory(mockUI, mockService)()
				u.So(t, err, gc.ShouldBeNil)

				var secretID string
				updateCommand := cmd.(*SecretsUpdateCommand)
				setup(updateCommand.SecretsBaseCommand, &mockClientFunctions{
					updateSecretByIDFn: func(appID, groupID, id, value string) error {
						secretID = id
						return nil
					},
				})
				exitCode := updateCommand.Run(tc.args)
				u.So(t, exitCode, gc.ShouldEqual, 0)
				u.So(t, secretID, gc.ShouldEqual, "thisisanid")
			})
		}

		for _, tc := range []struct {
			description string
			args        []string
		}{
			{
				description: "it passes the correct flags to UpdateSecretByName",
				args:        validUpdateByNameArgs,
			},
			{
				description: "it passes the correct flags to UpdateSecretByName with deprecated args",
				args:        validUpdateByNameDeprecatedArgs,
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				mockUI := cli.NewMockUi()
				cmd, err := NewSecretsUpdateCommandFactory(mockUI, mockService)()
				u.So(t, err, gc.ShouldBeNil)

				var secretName string
				updateCommand := cmd.(*SecretsUpdateCommand)
				setup(updateCommand.SecretsBaseCommand, &mockClientFunctions{
					updateSecretByNameFn: func(appID, groupID, name, value string) error {
						secretName = name
						return nil
					},
				})
				exitCode := updateCommand.Run(tc.args)
				u.So(t, exitCode, gc.ShouldEqual, 0)
				u.So(t, secretName, gc.ShouldEqual, "thisisaname")
			})
		}

		for _, tc := range []struct {
			description string
			args        []string
		}{
			{
				description: "it passes the correct flags to RemoveSecretByID",
				args:        validRemoveByIDArgs,
			},
			{
				description: "it passes the correct flags to RemoveSecretByID with deprecated args",
				args:        validRemoveByIDDeprecatedArgs,
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				mockUI := cli.NewMockUi()
				cmd, err := NewSecretsRemoveCommandFactory(mockUI, mockService)()
				u.So(t, err, gc.ShouldBeNil)

				removeCommand := cmd.(*SecretsRemoveCommand)
				var secretID string
				setup(removeCommand.SecretsBaseCommand, &mockClientFunctions{
					removeSecretByIDFn: func(appID, groupID, id string) error {
						secretID = id
						return nil
					},
				})
				exitCode := removeCommand.Run(tc.args)
				u.So(t, exitCode, gc.ShouldEqual, 0)
				u.So(t, secretID, gc.ShouldEqual, "thisisanid")
			})
		}

		for _, tc := range []struct {
			description string
			args        []string
		}{
			{
				description: "it passes the correct flags to RemoveSecretByName",
				args:        validRemoveByNameArgs,
			},
			{
				description: "it passes the correct flags to RemoveSecretByName with deprecated args",
				args:        validRemoveByNameDeprecatedArgs,
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				mockUI := cli.NewMockUi()
				cmd, err := NewSecretsRemoveCommandFactory(mockUI, mockService)()
				u.So(t, err, gc.ShouldBeNil)

				removeCommand := cmd.(*SecretsRemoveCommand)
				var secretName string
				setup(removeCommand.SecretsBaseCommand, &mockClientFunctions{
					removeSecretByNameFn: func(appID, groupID, name string) error {
						secretName = name
						return nil
					},
				})
				exitCode := removeCommand.Run(tc.args)
				u.So(t, exitCode, gc.ShouldEqual, 0)
				u.So(t, secretName, gc.ShouldEqual, "thisisaname")
			})
		}

		t.Run("listing secrets works when there are none", func(t *testing.T) {
			mockUI := cli.NewMockUi()
			cmd, err := NewSecretsListCommandFactory(mockUI, mockService)()
			u.So(t, err, gc.ShouldBeNil)

			listCommand := cmd.(*SecretsListCommand)
			setup(listCommand.SecretsBaseCommand, nil)
			exitCode := listCommand.Run(validListArgs)
			u.So(t, exitCode, gc.ShouldEqual, 0)
			u.So(t, mockUI.OutputWriter.String(), gc.ShouldContainSubstring, "No secrets found for this app")
		})

		t.Run("listing secrets works", func(t *testing.T) {
			mockUI := cli.NewMockUi()
			cmd, err := NewSecretsListCommandFactory(mockUI, mockService)()
			u.So(t, err, gc.ShouldBeNil)

			listCommand := cmd.(*SecretsListCommand)
			setup(listCommand.SecretsBaseCommand, &mockClientFunctions{
				listSecretsFn: func(appID, groupID string) ([]secrets.Secret, error) {
					return []secrets.Secret{{ID: "123", Name: "hello", Value: "there"}}, nil
				},
			})
			exitCode := listCommand.Run(validListArgs)
			u.So(t, exitCode, gc.ShouldEqual, 0)
			u.So(t, mockUI.OutputWriter.String(), gc.ShouldContainSubstring, "123")
			u.So(t, mockUI.OutputWriter.String(), gc.ShouldContainSubstring, "hello")
			u.So(t, mockUI.OutputWriter.String(), gc.ShouldNotContainSubstring, "there")
		})

		t.Run("adding a secret works", func(t *testing.T) {
			mockUI := cli.NewMockUi()
			cmd, err := NewSecretsAddCommandFactory(mockUI, mockService)()
			u.So(t, err, gc.ShouldBeNil)

			addCommand := cmd.(*SecretsAddCommand)
			setup(addCommand.SecretsBaseCommand, nil)

			exitCode := addCommand.Run(validAddArgs)
			u.So(t, exitCode, gc.ShouldEqual, 0)
			u.So(t, mockUI.OutputWriter.String(), gc.ShouldContainSubstring, "New secret created")
		})

		t.Run("updating a secret by id works", func(t *testing.T) {
			mockUI := cli.NewMockUi()
			cmd, err := NewSecretsUpdateCommandFactory(mockUI, mockService)()
			u.So(t, err, gc.ShouldBeNil)

			updateCommand := cmd.(*SecretsUpdateCommand)
			setup(updateCommand.SecretsBaseCommand, nil)

			exitCode := updateCommand.Run(validUpdateByIDArgs)
			u.So(t, exitCode, gc.ShouldEqual, 0)
			u.So(t, mockUI.OutputWriter.String(), gc.ShouldContainSubstring, "Secret updated: thisisanid")
		})

		t.Run("updating a secret by name works", func(t *testing.T) {
			mockUI := cli.NewMockUi()
			cmd, err := NewSecretsUpdateCommandFactory(mockUI, mockService)()
			u.So(t, err, gc.ShouldBeNil)

			updateCommand := cmd.(*SecretsUpdateCommand)
			setup(updateCommand.SecretsBaseCommand, nil)

			exitCode := updateCommand.Run(validUpdateByNameArgs)
			u.So(t, exitCode, gc.ShouldEqual, 0)
			u.So(t, mockUI.OutputWriter.String(), gc.ShouldContainSubstring, "Secret updated: thisisaname")
		})

		t.Run("removing a secret by id works", func(t *testing.T) {
			mockUI := cli.NewMockUi()
			cmd, err := NewSecretsRemoveCommandFactory(mockUI, mockService)()
			u.So(t, err, gc.ShouldBeNil)

			removeCommand := cmd.(*SecretsRemoveCommand)
			setup(removeCommand.SecretsBaseCommand, nil)

			exitCode := removeCommand.Run(validRemoveByIDArgs)
			u.So(t, exitCode, gc.ShouldEqual, 0)
			u.So(t, mockUI.OutputWriter.String(), gc.ShouldContainSubstring, "Secret removed: thisisanid")
		})

		t.Run("removing a secret by name works", func(t *testing.T) {
			mockUI := cli.NewMockUi()
			cmd, err := NewSecretsRemoveCommandFactory(mockUI, mockService)()
			u.So(t, err, gc.ShouldBeNil)

			removeCommand := cmd.(*SecretsRemoveCommand)
			setup(removeCommand.SecretsBaseCommand, nil)

			exitCode := removeCommand.Run(validRemoveByNameArgs)
			u.So(t, exitCode, gc.ShouldEqual, 0)
			u.So(t, mockUI.OutputWriter.String(), gc.ShouldContainSubstring, "Secret removed: thisisaname")
		})
	})
}
