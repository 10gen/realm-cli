package commands

import (
	"fmt"
	"os"

	"github.com/10gen/stitch-cli/models"
	"github.com/10gen/stitch-cli/secrets"
	u "github.com/10gen/stitch-cli/user"
	"github.com/10gen/stitch-cli/utils"
	"github.com/mitchellh/cli"
)

// NewSecretsCommandFactory returns a new cli.CommandFactory given a cli.Ui
func NewSecretsCommandFactory(ui cli.Ui) cli.CommandFactory {
	return func() (cli.Command, error) {
		return &SecretsCommand{
			BaseCommand: &BaseCommand{
				Name: "secrets",
				UI:   ui,
			},
		}, nil
	}
}

// SecretsCommand is used to run CRUD operations on a Stitch App's secrets
type SecretsCommand struct {
	*BaseCommand
}

// Synopsis returns a one-liner description for this command
func (sc *SecretsCommand) Synopsis() string {
	return "Add or remove secrets for your Stitch App."
}

// Help returns long-form help information for this command
func (sc *SecretsCommand) Help() string {
	return sc.Synopsis()
}

// Run executes the command
func (sc *SecretsCommand) Run(args []string) int {
	return cli.RunResultHelp
}

const (
	flagSecretName                     = "name"
	flagSecretValue                    = "value"
	flagSecretID                       = "id"
	flagSecretNameIdentifier           = "name"
	flagSecretIDDeprecated             = "secret-id"
	flagSecretNameIdentifierDeprecated = "secret-name"
)

var (
	errSecretNameRequired     = fmt.Errorf("a name (--%s=[string]) is required", flagSecretName)
	errSecretValueRequired    = fmt.Errorf("a value (--%s=[string]) is required", flagSecretValue)
	errSecretIDOrNameRequired = fmt.Errorf("a Secret name or ID (--%s=[string] or --%s=[string]) is required", flagSecretNameIdentifier, flagSecretID)
)

// NewSecretsBaseCommand returns a new *SecretsBaseCommand
func NewSecretsBaseCommand(name, workingDirectory string, ui cli.Ui) *SecretsBaseCommand {
	return &SecretsBaseCommand{
		ProjectCommand:   NewProjectCommand(name, ui),
		workingDirectory: workingDirectory,
	}
}

// SecretsBaseCommand represents a common Atlas project-based secrets command
type SecretsBaseCommand struct {
	*ProjectCommand

	workingDirectory string

	flagAppID string
}

// Help returns long-form help information for the SecretsBaseCommand command
func (sbc *SecretsBaseCommand) Help() string {
	return `
OPTIONAL:
  --app-id [string]
	The App ID for your app (i.e. the name of your app followed by a unique suffix, like "my-app-nysja").
	Required if not being run from within a stitch project directory.` +
		sbc.ProjectCommand.Help()
}

func (sbc *SecretsBaseCommand) run(args []string) error {
	if sbc.FlagSet == nil {
		sbc.NewFlagSet()
	}

	sbc.FlagSet.StringVar(&sbc.flagAppID, flagAppIDName, "", "")

	if err := sbc.ProjectCommand.run(args); err != nil {
		return err
	}

	user, err := sbc.User()
	if err != nil {
		return err
	}

	if !user.LoggedIn() {
		return u.ErrNotLoggedIn
	}

	return nil
}

func (sbc *SecretsBaseCommand) resolveApp() (*models.App, error) {
	appID := sbc.flagAppID
	if sbc.flagAppID == "" {
		appPath, err := utils.ResolveAppDirectory("", sbc.workingDirectory)
		if err != nil {
			return nil, err
		}

		appInstanceData, err := utils.ResolveAppInstanceData(sbc.flagAppID, appPath)
		if err != nil {
			return nil, err
		}
		appID = appInstanceData.AppID()
	}

	stitchClient, err := sbc.StitchClient()
	if err != nil {
		return nil, err
	}

	var app *models.App
	if sbc.flagProjectID == "" {
		app, err = stitchClient.FetchAppByClientAppID(appID)
		if err != nil {
			return nil, err
		}
	} else {
		app, err = stitchClient.FetchAppByGroupIDAndClientAppID(sbc.flagProjectID, appID)
		if err != nil {
			return nil, err
		}
	}

	return app, nil
}

// NewSecretsListCommandFactory returns a new cli.CommandFactory given a cli.Ui
func NewSecretsListCommandFactory(ui cli.Ui) cli.CommandFactory {
	return func() (cli.Command, error) {
		workingDirectory, err := os.Getwd()
		if err != nil {
			return nil, err
		}

		return &SecretsListCommand{
			SecretsBaseCommand: NewSecretsBaseCommand("list", workingDirectory, ui),
		}, nil
	}
}

// SecretsListCommand is used to list secrets from a Stitch app
type SecretsListCommand struct {
	*SecretsBaseCommand
}

// Synopsis returns a one-liner description for this command
func (slc *SecretsListCommand) Synopsis() string {
	return "List secrets from your Stitch App."
}

// Help returns long-form help information for this command
func (slc *SecretsListCommand) Help() string {
	return `List secrets from your Stitch Application.

Usage: stitch-cli secrets list [options]
` +
		slc.SecretsBaseCommand.Help()
}

// Run executes the command
func (slc *SecretsListCommand) Run(args []string) int {
	if err := slc.SecretsBaseCommand.run(args); err != nil {
		slc.UI.Error(err.Error())
		return 1
	}

	secrets, err := slc.listSecrets()
	if err != nil {
		slc.UI.Error(err.Error())
		return 1
	}

	if len(secrets) == 0 {
		slc.UI.Info("No secrets found for this app")
		return 0
	}

	slc.UI.Info("ID                       Name")
	for _, secret := range secrets {
		slc.UI.Info(fmt.Sprintf("%s %s", secret.ID, secret.Name))
	}

	return 0
}

func (slc *SecretsListCommand) listSecrets() ([]secrets.Secret, error) {
	app, err := slc.resolveApp()
	if err != nil {
		return nil, err
	}

	stitchClient, err := slc.StitchClient()
	if err != nil {
		return nil, err
	}

	return stitchClient.ListSecrets(app.GroupID, app.ID)
}

// NewSecretsAddCommandFactory returns a new cli.CommandFactory given a cli.Ui
func NewSecretsAddCommandFactory(ui cli.Ui) cli.CommandFactory {
	return func() (cli.Command, error) {
		workingDirectory, err := os.Getwd()
		if err != nil {
			return nil, err
		}

		return &SecretsAddCommand{
			SecretsBaseCommand: NewSecretsBaseCommand("add", workingDirectory, ui),
		}, nil
	}
}

// SecretsAddCommand is used to add secrets to a Stitch app
type SecretsAddCommand struct {
	*SecretsBaseCommand

	flagSecretName  string
	flagSecretValue string
}

// Synopsis returns a one-liner description for this command
func (sac *SecretsAddCommand) Synopsis() string {
	return "Add a secret to your Stitch App."
}

// Help returns long-form help information for this command
func (sac *SecretsAddCommand) Help() string {
	return `Add a secret to your Stitch Application.

Usage: stitch-cli secrets add --name [string] --value [string] [options]

REQUIRED:
  --name [string]
	The name of your secret.

  --value [string]
	The value of your secret.
` +
		sac.SecretsBaseCommand.Help()
}

// Run executes the command
func (sac *SecretsAddCommand) Run(args []string) int {
	sac.NewFlagSet()

	sac.FlagSet.StringVar(&sac.flagSecretName, flagSecretName, "", "")
	sac.FlagSet.StringVar(&sac.flagSecretValue, flagSecretValue, "", "")

	if err := sac.SecretsBaseCommand.run(args); err != nil {
		sac.UI.Error(err.Error())
		return 1
	}

	if err := sac.addSecret(); err != nil {
		sac.UI.Error(err.Error())
		return 1
	}

	return 0
}

func (sac *SecretsAddCommand) addSecret() error {
	if sac.flagSecretName == "" {
		return errSecretNameRequired
	}

	if sac.flagSecretValue == "" {
		return errSecretValueRequired
	}

	app, err := sac.resolveApp()
	if err != nil {
		return err
	}

	stitchClient, err := sac.StitchClient()
	if err != nil {
		return err
	}

	if addErr := stitchClient.AddSecret(app.GroupID, app.ID, secrets.Secret{
		Name:  sac.flagSecretName,
		Value: sac.flagSecretValue,
	}); addErr != nil {
		return addErr
	}

	sac.UI.Info(fmt.Sprintf("New secret created: %s", sac.flagSecretName))
	return nil
}

// NewSecretsUpdateCommandFactory returns a new cli.CommandFactory given a cli.Ui
func NewSecretsUpdateCommandFactory(ui cli.Ui) cli.CommandFactory {
	return func() (cli.Command, error) {
		workingDirectory, err := os.Getwd()
		if err != nil {
			return nil, err
		}

		return &SecretsUpdateCommand{
			SecretsBaseCommand: NewSecretsBaseCommand("update", workingDirectory, ui),
		}, nil
	}
}

// SecretsUpdateCommand is used to update a secret from a Stitch app
type SecretsUpdateCommand struct {
	*SecretsBaseCommand

	flagSecretID    string
	flagSecretName  string
	flagSecretValue string
}

// Synopsis returns a one-liner description for this command
func (suc *SecretsUpdateCommand) Synopsis() string {
	return "Update a secret for your Stitch App."
}

// Help returns long-form help information for this command
func (suc *SecretsUpdateCommand) Help() string {
	return `Update a secret for your Stitch Application.

Usage:
  stitch-cli secrets update --name [string] --value [string] [options]
  stitch-cli secrets update --id [string] --value [string] [options]

REQUIRED:
  --name [string] OR --id [string]
	The name or ID of your secret.

  --value [string]
	The value that your secret is being updated to.
` +
		suc.SecretsBaseCommand.Help()
}

// Run executes the command
func (suc *SecretsUpdateCommand) Run(args []string) int {
	suc.NewFlagSet()

	suc.FlagSet.StringVar(&suc.flagSecretID, flagSecretID, "", "")
	suc.FlagSet.StringVar(&suc.flagSecretID, flagSecretIDDeprecated, "", "")
	suc.FlagSet.StringVar(&suc.flagSecretName, flagSecretNameIdentifier, "", "")
	suc.FlagSet.StringVar(&suc.flagSecretName, flagSecretNameIdentifierDeprecated, "", "")
	suc.FlagSet.StringVar(&suc.flagSecretValue, flagSecretValue, "", "")

	if err := suc.SecretsBaseCommand.run(args); err != nil {
		suc.UI.Error(err.Error())
		return 1
	}

	if err := suc.updateSecret(); err != nil {
		suc.UI.Error(err.Error())
		return 1
	}

	return 0
}

func (suc *SecretsUpdateCommand) updateSecret() error {
	if suc.flagSecretID == "" && suc.flagSecretName == "" {
		return errSecretIDOrNameRequired
	}

	app, err := suc.resolveApp()
	if err != nil {
		return err
	}

	stitchClient, err := suc.StitchClient()
	if err != nil {
		return err
	}

	if suc.flagSecretID != "" {
		if updateErr := stitchClient.UpdateSecretByID(app.GroupID, app.ID, suc.flagSecretID, suc.flagSecretValue); updateErr != nil {
			return updateErr
		}
		suc.UI.Info(fmt.Sprintf("Secret updated: %s", suc.flagSecretID))
	} else {
		if updateErr := stitchClient.UpdateSecretByName(app.GroupID, app.ID, suc.flagSecretName, suc.flagSecretValue); updateErr != nil {
			return updateErr
		}
		suc.UI.Info(fmt.Sprintf("Secret updated: %s", suc.flagSecretName))
	}

	return nil
}

// NewSecretsRemoveCommandFactory returns a new cli.CommandFactory given a cli.Ui
func NewSecretsRemoveCommandFactory(ui cli.Ui) cli.CommandFactory {
	return func() (cli.Command, error) {
		workingDirectory, err := os.Getwd()
		if err != nil {
			return nil, err
		}

		return &SecretsRemoveCommand{
			SecretsBaseCommand: NewSecretsBaseCommand("remove", workingDirectory, ui),
		}, nil
	}
}

// SecretsRemoveCommand is used to remove secrets from a Stitch app
type SecretsRemoveCommand struct {
	*SecretsBaseCommand

	flagSecretID   string
	flagSecretName string
}

// Synopsis returns a one-liner description for this command
func (src *SecretsRemoveCommand) Synopsis() string {
	return "Remove a secret from your Stitch App."
}

// Help returns long-form help information for this command
func (src *SecretsRemoveCommand) Help() string {
	return `Remove a secret from your Stitch Application.

Usage:
  stitch-cli secrets remove --name [string] [options]
  stitch-cli secrets remove --id [string] [options]

REQUIRED:
  --name [string] OR --id [string]
	The name or ID of your secret.
` +
		src.SecretsBaseCommand.Help()
}

// Run executes the command
func (src *SecretsRemoveCommand) Run(args []string) int {
	src.NewFlagSet()

	src.FlagSet.StringVar(&src.flagSecretID, flagSecretID, "", "")
	src.FlagSet.StringVar(&src.flagSecretID, flagSecretIDDeprecated, "", "")
	src.FlagSet.StringVar(&src.flagSecretName, flagSecretNameIdentifier, "", "")
	src.FlagSet.StringVar(&src.flagSecretName, flagSecretNameIdentifierDeprecated, "", "")

	if err := src.SecretsBaseCommand.run(args); err != nil {
		src.UI.Error(err.Error())
		return 1
	}

	if err := src.removeSecret(); err != nil {
		src.UI.Error(err.Error())
		return 1
	}

	return 0
}

func (src *SecretsRemoveCommand) removeSecret() error {
	if src.flagSecretID == "" && src.flagSecretName == "" {
		return errSecretIDOrNameRequired
	}

	app, err := src.resolveApp()
	if err != nil {
		return err
	}

	stitchClient, err := src.StitchClient()
	if err != nil {
		return err
	}

	if src.flagSecretID != "" {
		if removeErr := stitchClient.RemoveSecretByID(app.GroupID, app.ID, src.flagSecretID); removeErr != nil {
			return removeErr
		}
		src.UI.Info(fmt.Sprintf("Secret removed: %s", src.flagSecretID))
	} else {
		if removeErr := stitchClient.RemoveSecretByName(app.GroupID, app.ID, src.flagSecretName); removeErr != nil {
			return removeErr
		}
		src.UI.Info(fmt.Sprintf("Secret removed: %s", src.flagSecretName))
	}

	return nil
}
