package commands

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/10gen/stitch-cli/api"
	u "github.com/10gen/stitch-cli/user"
	"github.com/10gen/stitch-cli/utils"

	"github.com/mitchellh/cli"
)

const (
	flagExportAppIDName   = "app-id"
	flagExportGroupIDName = "group-id"
)

var (
	errExportMissingFilename = errors.New("the app export response did not specify a filename")
	errAppIDRequired         = fmt.Errorf("an App ID (--%s=<APP_ID>) must be supplied to export an app", flagExportAppIDName)
	errGroupIDRequired       = fmt.Errorf("a Group ID (--%s=<GROUP_ID>) must be supplied to export an app", flagExportGroupIDName)
)

// NewExportCommandFactory returns a new cli.CommandFactory given a cli.Ui
func NewExportCommandFactory(ui cli.Ui) cli.CommandFactory {
	return func() (cli.Command, error) {
		return &ExportCommand{
			exportToDirectory: utils.WriteZipToDir,
			BaseCommand: &BaseCommand{
				Name: "export",
				UI:   ui,
			},
		}, nil
	}
}

// ExportCommand is used to export a Stitch App
type ExportCommand struct {
	*BaseCommand

	exportToDirectory func(dest string, zipData io.Reader) error

	flagAppID   string
	flagGroupID string
	flagOutput  string
}

// Help returns long-form help information for this command
func (ec *ExportCommand) Help() string {
	return `Export a stitch application to a local directory.

REQUIRED:
  --app-id <STRING>

OPTIONS:
  -o, --output <DIRECTORY>
	Directory to write the exported configuration. Defaults to "<app_name>_<timestamp>"`
}

// Synopsis returns a one-liner description for this command
func (ec *ExportCommand) Synopsis() string {
	return `Export a stitch application to a local directory.`
}

// Run executes the command
func (ec *ExportCommand) Run(args []string) int {
	set := ec.NewFlagSet()

	set.StringVar(&ec.flagAppID, flagExportAppIDName, "", "")
	set.StringVar(&ec.flagGroupID, flagExportGroupIDName, "", "")
	set.StringVarP(&ec.flagOutput, "output", "o", "", "")

	if err := ec.BaseCommand.run(args); err != nil {
		ec.UI.Error(err.Error())
		return 1
	}

	if err := ec.export(); err != nil {
		ec.UI.Error(err.Error())
		return 1
	}

	return 0
}

func (ec *ExportCommand) export() error {
	if ec.flagAppID == "" {
		return errAppIDRequired
	}

	if ec.flagGroupID == "" {
		return errGroupIDRequired
	}

	user, err := ec.User()
	if err != nil {
		return err
	}

	if !user.LoggedIn() {
		return u.ErrNotLoggedIn
	}

	authClient, err := ec.AuthClient()
	if err != nil {
		return err
	}

	res, err := api.NewStitchClient(ec.flagBaseURL, authClient).Export(ec.flagGroupID, ec.flagAppID)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("expected API status to be %d, received %d instead: %s", http.StatusOK, res.StatusCode, res.Status)
	}

	_, params, err := mime.ParseMediaType(res.Header.Get("Content-Disposition"))
	if err != nil {
		return err
	}

	filename := params["filename"]
	if len(filename) == 0 {
		return errExportMissingFilename
	}

	if ec.flagOutput != "" {
		filename = ec.flagOutput
	}

	return ec.exportToDirectory(strings.Replace(filename, ".zip", "", 1), res.Body)
}
