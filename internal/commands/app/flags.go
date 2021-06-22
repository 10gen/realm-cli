package app

import (
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/flags"
)

const (
	flagRemoteApp       = "remote"
	flagName            = "name"
	flagDeploymentModel = "deployment-model"
	flagLocation        = "location"
	flagEnvironment     = "environment"
	flagProject         = "project"

	flagConfigVersionDescription = "Specify the config version for the new Realm app"
)

func remoteAppFlag(value *string) flags.StringFlag {
	return flags.StringFlag{
		Value: value,
		Meta: flags.Meta{
			Name: flagRemoteApp,
			Usage: flags.Usage{
				Description: "Specify the name or ID of a remote Realm app to clone",
			},
		},
	}
}

func nameFlag(value *string) flags.StringFlag {
	return flags.StringFlag{
		Value: value,
		Meta: flags.Meta{
			Name:      "name",
			Shorthand: "n",
			Usage: flags.Usage{
				Description: "Name your new Realm app",
			},
		},
	}
}

func locationFlag(value *realm.Location) flags.CustomFlag {
	return flags.CustomFlag{
		Value: value,
		Meta: flags.Meta{
			Name:      "location",
			Shorthand: "l",
			Usage: flags.Usage{
				Description:  "Select the Realm app's location",
				DefaultValue: "<none>",
				AllowedValues: []string{
					string(realm.LocationVirginia),
					string(realm.LocationOregon),
					string(realm.LocationFrankfurt),
					string(realm.LocationIreland),
					string(realm.LocationSydney),
					string(realm.LocationMumbai),
					string(realm.LocationSingapore),
				},
			},
		},
	}
}

func deploymentModelFlag(value *realm.DeploymentModel) flags.CustomFlag { //nolint: interfacer
	return flags.CustomFlag{
		Value: value,
		Meta: flags.Meta{
			Name:      "deployment-model",
			Shorthand: "d",
			Usage: flags.Usage{
				Description:  "Select the Realm app's deployment model",
				DefaultValue: "<none>",
				AllowedValues: []string{
					string(realm.DeploymentModelGlobal),
					string(realm.DeploymentModelLocal),
				},
			},
		},
	}
}

func environmentFlag(value *realm.Environment) flags.CustomFlag {
	return flags.CustomFlag{
		Value: value,
		Meta: flags.Meta{
			Name:      "environment",
			Shorthand: "e",
			Usage: flags.Usage{
				Description:  "Select the Realm app's environment",
				DefaultValue: "<none>",
				AllowedValues: []string{
					string(realm.EnvironmentDevelopment),
					string(realm.EnvironmentTesting),
					string(realm.EnvironmentQA),
					string(realm.EnvironmentProduction),
				},
			},
		},
	}
}
