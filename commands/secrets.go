package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/10gen/stitch-cli/secrets"
	u "github.com/10gen/stitch-cli/user"
	"github.com/10gen/stitch-cli/utils"
	"github.com/mitchellh/cli"
)

// NewSecretsCommandFactory returns a new cli.CommandFactory given a cli.Ui
func NewSecretsCommandFactory(ui cli.Ui) cli.CommandFactory {
	return func() (cli.Command, error) {
		c := cli.NewCLI(filepath.Base(os.Args[0]), utils.CLIVersion)
		c.Args = os.Args[1:]

		workingDirectory, err := os.Getwd()
		if err != nil {
			return nil, err
		}

		sc := &SecretsCommand{
			BasePath: filepath.Base(os.Args[0]),
			BaseCommand: &BaseCommand{
				Name: "secrets",
				CLI:  c,
				UI:   ui,
			},
			workingDirectory: workingDirectory,
		}

		c.Commands = map[string]cli.CommandFactory{
			"list":   NewSecretsListCommandFactory(sc),
			"add":    NewSecretsAddCommandFactory(sc),
			"update": NewSecretsUpdateCommandFactory(sc),
			"remove": NewSecretsRemoveCommandFactory(sc),
		}

		return sc, nil
	}
}

// SecretsCommand is used to run CRUD operations on a Stitch App's secrets
type SecretsCommand struct {
	*BaseCommand

	Name     string
	BasePath string

	workingDirectory string
}

// Synopsis returns a one-liner description for this command
func (sc *SecretsCommand) Synopsis() string {
	return "Add or remove secrets for your Stitch App."
}

// Help returns long-form help information for this command
func (sc *SecretsCommand) Help() string {
	return sc.BaseCommand.CLI.HelpFunc(sc.CLI.Commands)
}

// Run executes the command
func (sc *SecretsCommand) Run(args []string) int {
	sc.BaseCommand.CLI.Args = args

	exitStatus, err := sc.BaseCommand.CLI.Run()
	if err != nil {
		sc.BaseCommand.UI.Error(err.Error())
	}

	return exitStatus
}

const (
	flagSecretName           = "name"
	flagSecretValue          = "value"
	flagSecretID             = "secret-id"
	flagSecretNameIdentifier = "secret-name"
)

var (
	errSecretNameRequired     = fmt.Errorf("a name (--%s=[string]) is required", flagSecretName)
	errSecretValueRequired    = fmt.Errorf("a value (--%s=[string]) is required", flagSecretValue)
	errSecretIDOrNameRequired = fmt.Errorf("a Secret name or ID (--%s=[string] or --%s=[string]) is required", flagSecretNameIdentifier, flagSecretID)
)

// NewSecretsListCommandFactory returns a new cli.CommandFactory given a cli.Ui
func NewSecretsListCommandFactory(sc *SecretsCommand) cli.CommandFactory {
	return func() (cli.Command, error) {
		sc.Name = "list"

		return &SecretsListCommand{
			SecretsCommand: sc,
		}, nil
	}
}

// SecretsListCommand is used to list secrets from a Stitch app
type SecretsListCommand struct {
	*SecretsCommand

	flagAppID string
}

// Synopsis returns a one-liner description for this command
func (slc *SecretsListCommand) Synopsis() string {
	return "List secrets from your Stitch App."
}

// Help returns long-form help information for this command
func (slc *SecretsListCommand) Help() string {
	return `List secrets from your Stitch Application.

OPTIONAL:
  --app-id [string]
	The App ID for your app (i.e. the name of your app followed by a unique suffix, like "my-app-nysja").
	Required if not being run from within a stitch project directory.
	` +
		slc.BaseCommand.Help()
}

// Run executes the command
func (slc *SecretsListCommand) Run(args []string) int {
	flags := slc.NewFlagSet()

	flags.StringVar(&slc.flagAppID, flagAppIDName, "", "")

	if err := slc.BaseCommand.run(args); err != nil {
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
	appID := slc.flagAppID
	if slc.flagAppID == "" {
		appPath, err := utils.ResolveAppDirectory("", slc.workingDirectory)
		if err != nil {
			return nil, err
		}

		appInstanceData, err := utils.ResolveAppInstanceData(slc.flagAppID, appPath)
		if err != nil {
			return nil, err
		}
		appID = appInstanceData.AppID()
	}

	user, err := slc.User()
	if err != nil {
		return nil, err
	}

	if !user.LoggedIn() {
		return nil, u.ErrNotLoggedIn
	}

	stitchClient, err := slc.StitchClient()
	if err != nil {
		return nil, err
	}

	app, err := stitchClient.FetchAppByClientAppID(appID)
	if err != nil {
		return nil, err
	}

	return stitchClient.ListSecrets(app.GroupID, app.ID)
}

// NewSecretsAddCommandFactory returns a new cli.CommandFactory given a cli.Ui
func NewSecretsAddCommandFactory(sc *SecretsCommand) cli.CommandFactory {
	return func() (cli.Command, error) {
		sc.Name = "add"

		return &SecretsAddCommand{
			SecretsCommand: sc,
		}, nil
	}
}

// SecretsAddCommand is used to add secrets to a Stitch app
type SecretsAddCommand struct {
	*SecretsCommand

	flagAppID       string
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

REQUIRED:
  --name [string]
	The name of your secret.

  --value [string]
	The value of your secret.

OPTIONAL:
  --app-id [string]
	The App ID for your app (i.e. the name of your app followed by a unique suffix, like "my-app-nysja").
	Required if not being run from within a stitch project directory.
	` +
		sac.BaseCommand.Help()
}

// Run executes the command
func (sac *SecretsAddCommand) Run(args []string) int {
	flags := sac.NewFlagSet()

	flags.StringVar(&sac.flagAppID, flagAppIDName, "", "")
	flags.StringVar(&sac.flagSecretName, flagSecretName, "", "")
	flags.StringVar(&sac.flagSecretValue, flagSecretValue, "", "")

	if err := sac.BaseCommand.run(args); err != nil {
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
	appID := sac.flagAppID
	if sac.flagAppID == "" {
		appPath, err := utils.ResolveAppDirectory("", sac.workingDirectory)
		if err != nil {
			return err
		}

		appInstanceData, err := utils.ResolveAppInstanceData(sac.flagAppID, appPath)
		if err != nil {
			return err
		}
		appID = appInstanceData.AppID()
	}

	if sac.flagSecretName == "" {
		return errSecretNameRequired
	}

	if sac.flagSecretValue == "" {
		return errSecretValueRequired
	}

	user, err := sac.User()
	if err != nil {
		return err
	}

	if !user.LoggedIn() {
		return u.ErrNotLoggedIn
	}

	stitchClient, err := sac.StitchClient()
	if err != nil {
		return err
	}

	app, err := stitchClient.FetchAppByClientAppID(appID)
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
func NewSecretsUpdateCommandFactory(sc *SecretsCommand) cli.CommandFactory {
	return func() (cli.Command, error) {
		sc.Name = "update"

		return &SecretsUpdateCommand{
			SecretsCommand: sc,
		}, nil
	}
}

// SecretsUpdateCommand is used to update a secret from a Stitch app
type SecretsUpdateCommand struct {
	*SecretsCommand

	flagAppID       string
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

REQUIRED:
  --secret-name [string] OR --secret-id [string]
	The name or ID of your secret.

  --value [string]
	The value that your secret is being updated to.

OPTIONAL:
  --app-id [string]
	The App ID for your app (i.e. the name of your app followed by a unique suffix, like "my-app-nysja").
	Required if not being run from within a stitch project directory.
	` +
		suc.BaseCommand.Help()
}

// Run executes the command
func (suc *SecretsUpdateCommand) Run(args []string) int {
	flags := suc.NewFlagSet()

	flags.StringVar(&suc.flagAppID, flagAppIDName, "", "")
	flags.StringVar(&suc.flagSecretID, flagSecretID, "", "")
	flags.StringVar(&suc.flagSecretName, flagSecretNameIdentifier, "", "")
	flags.StringVar(&suc.flagSecretValue, flagSecretValue, "", "")

	if err := suc.BaseCommand.run(args); err != nil {
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
	appID := suc.flagAppID
	if suc.flagAppID == "" {
		appPath, err := utils.ResolveAppDirectory("", suc.workingDirectory)
		if err != nil {
			return err
		}

		appInstanceData, err := utils.ResolveAppInstanceData(suc.flagAppID, appPath)
		if err != nil {
			return err
		}
		appID = appInstanceData.AppID()
	}

	if suc.flagSecretID == "" && suc.flagSecretName == "" {
		return errSecretIDOrNameRequired
	}

	if suc.flagSecretValue == "" {
		return errSecretValueRequired
	}

	user, err := suc.User()
	if err != nil {
		return err
	}

	if !user.LoggedIn() {
		return u.ErrNotLoggedIn
	}

	stitchClient, err := suc.StitchClient()
	if err != nil {
		return err
	}

	app, err := stitchClient.FetchAppByClientAppID(appID)
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
func NewSecretsRemoveCommandFactory(sc *SecretsCommand) cli.CommandFactory {
	return func() (cli.Command, error) {
		sc.Name = "remove"

		return &SecretsRemoveCommand{
			SecretsCommand: sc,
		}, nil
	}
}

// SecretsRemoveCommand is used to remove secrets from a Stitch app
type SecretsRemoveCommand struct {
	*SecretsCommand

	flagAppID      string
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

REQUIRED:
  --secret-name [string] OR --secret-id [string]
	The name or ID of your secret.

OPTIONAL:
  --app-id [string]
	The App ID for your app (i.e. the name of your app followed by a unique suffix, like "my-app-nysja").
	Required if not being run from within a stitch project directory.
	` +
		src.BaseCommand.Help()
}

// Run executes the command
func (src *SecretsRemoveCommand) Run(args []string) int {
	flags := src.NewFlagSet()

	flags.StringVar(&src.flagAppID, flagAppIDName, "", "")
	flags.StringVar(&src.flagSecretID, flagSecretID, "", "")
	flags.StringVar(&src.flagSecretName, flagSecretNameIdentifier, "", "")

	if err := src.BaseCommand.run(args); err != nil {
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
	appID := src.flagAppID
	if src.flagAppID == "" {
		appPath, err := utils.ResolveAppDirectory("", src.workingDirectory)
		if err != nil {
			return err
		}

		appInstanceData, err := utils.ResolveAppInstanceData(src.flagAppID, appPath)
		if err != nil {
			return err
		}
		appID = appInstanceData.AppID()
	}

	if src.flagSecretID == "" && src.flagSecretName == "" {
		return errSecretIDOrNameRequired
	}

	user, err := src.User()
	if err != nil {
		return err
	}

	if !user.LoggedIn() {
		return u.ErrNotLoggedIn
	}

	stitchClient, err := src.StitchClient()
	if err != nil {
		return err
	}

	app, err := stitchClient.FetchAppByClientAppID(appID)
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
