package list

import (
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/terminal"
)

type inputs struct {
	cli.ProjectAppInputs
	StateValue    stateValue
	Pending       bool
	ProviderTypes []string
	Users         []string
}

func (i *inputs) Resolve(profile *cli.Profile, ui terminal.UI) error {
	if err := i.ProjectAppInputs.Resolve(ui, profile.WorkingDirectory); err != nil {
		return err
	}

	return nil
}
