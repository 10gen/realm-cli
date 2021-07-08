package cli

import (
	"fmt"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/flags"
)

const (
	appFlagUsage = "Specify the name or ID of a Realm app"

	// ProjectFlagName is the '--project' flag name
	ProjectFlagName = "project"
)

// AppFlag is the '--app' flag
func AppFlag(value *string) flags.Flag {
	return AppFlagWithDescription(value, appFlagUsage)
}

// AppFlagWithContext is the '--app' flag with the provided context set into a standard description
func AppFlagWithContext(value *string, context string) flags.Flag {
	return AppFlagWithDescription(value, fmt.Sprintf("%s %s", appFlagUsage, context))
}

// AppFlagWithDescription is the '--app' flag with the provided description
func AppFlagWithDescription(value *string, description string) flags.Flag {
	return flags.StringFlag{
		Value: value,
		Meta: flags.Meta{
			Name:      "app",
			Shorthand: "a",
			Usage: flags.Usage{
				Description: description,
			},
		},
	}
}

// ConfigVersionFlag is the '--config-version' flag with the provided description
func ConfigVersionFlag(value *realm.AppConfigVersion, description string) flags.Flag { //nolint: interfacer
	return flags.CustomFlag{
		Value: value,
		Meta: flags.Meta{
			Name: "config-version",
			Usage: flags.Usage{
				Description: description,
			},
			Hidden: true,
		},
	}
}

// ProjectFlag is the '--project' flag
func ProjectFlag(value *string) flags.Flag {
	return flags.StringFlag{
		Value: value,
		Meta: flags.Meta{
			Name: ProjectFlagName,
			Usage: flags.Usage{
				Description: "Specify the ID of a MongoDB Atlas project",
			},
			Hidden: true,
		},
	}
}

// ProductFlag is the '--product' flag
func ProductFlag(value *[]string) flags.Flag {
	return flags.StringSliceFlag{
		Value: value,
		Meta: flags.Meta{
			Name: "product",
			Usage: flags.Usage{
				Description:   `Specify the Realm app product(s)`,
				AllowedValues: []string{`"standard"`, `"atlas"`},
			},
			Hidden: true,
		},
	}
}
