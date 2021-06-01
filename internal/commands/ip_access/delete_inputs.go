package ip_access

import (
	"fmt"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"
)

type deleteInputs struct {
	cli.ProjectInputs
	IPAddress string
}

func (i *deleteInputs) Resolve(profile *user.Profile, ui terminal.UI) error {
	if err := i.ProjectInputs.Resolve(ui, profile.WorkingDirectory, false); err != nil {
		return err
	}
	return nil
}

func (i *deleteInputs) resolveAllowedIP(ui terminal.UI, allowedIPs []realm.AllowedIP) (realm.AllowedIP, error) {
	if len(i.IPAddress) > 0 {
		for _, allowedIP := range allowedIPs {
			if allowedIP.IPAddress == i.IPAddress {
				return allowedIP, nil
			}
		}
		return realm.AllowedIP{}, fmt.Errorf("unable to find allowed IP: %s", i.IPAddress)
	}

	return realm.AllowedIP{}, nil
}
