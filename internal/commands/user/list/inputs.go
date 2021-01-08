package list

import (
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"
)

type inputs struct {
	cli.ProjectAppInputs
	UserState     realm.UserState
	Pending       bool
	ProviderTypes []string
	Users         []string
}

func (i *inputs) Resolve(profile *cli.Profile, ui terminal.UI) error {
	if !areValidProviderTypes(i.ProviderTypes) {
		return errInvalidProviderType
	}

	if err := i.ProjectAppInputs.Resolve(ui, profile.WorkingDirectory); err != nil {
		return err
	}

	return nil
}
