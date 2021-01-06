package list

import (
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/terminal"
)

// input field names, per survey
const (
	inputFieldEmail      = "email"
	inputFieldPassword   = "password"
	inputFieldAPIKeyName = "apiKeyName"
)

type inputs struct {
	cli.ProjectAppInputs
	StateValue    stateValue
	Pending       bool
	ProviderTypes []string
}

func (i *inputs) Resolve(profile *cli.Profile, ui terminal.UI) error {

	if err := i.ProjectAppInputs.Resolve(ui, profile.WorkingDirectory); err != nil {
		return err
	}

	//todo user multiselect

	return nil
}
