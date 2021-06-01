package ip_access

import (
	"fmt"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"
)

type updateInputs struct {
	cli.ProjectInputs
	IP      string
	NewIP   string
	Comment string
}

func (i *updateInputs) Resolve(profile *user.Profile, ui terminal.UI) error {
	if err := i.ProjectInputs.Resolve(ui, profile.WorkingDirectory, false); err != nil {
		return err
	}
	return nil
}

func (i *updateInputs) resolveAllowedIP(ui terminal.UI, allowedIPs []realm.AllowedIP) (realm.AllowedIP, error) {
	if len(i.IP) > 0 {
		for _, allowedIP := range allowedIPs {
			if allowedIP.IP == i.IP {
				return allowedIP, nil
			}
		}
		return realm.AllowedIP{}, fmt.Errorf("unable to find allowed IP: %s", i.IP)
	}

	return realm.AllowedIP{}, nil
}
