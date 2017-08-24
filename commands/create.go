package commands

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/10gen/stitch-cli/app"
	"github.com/10gen/stitch-cli/atlas"
	"github.com/10gen/stitch-cli/config"
	"github.com/10gen/stitch-cli/ui"

	flag "github.com/ogier/pflag"
)

var create = &Command{
	Run:  createRun,
	Name: "create",
	ShortUsage: `
USAGE:
    stitch create [--help] [-y] [-g <GROUP>] [-c <CLUSTER>] [-n <APP-NAME>] [-i <STITCH-CONFIG>]
`,
	LongUsage: `Create a new stitch application.

OPTIONS:
    -c, --cluster <CLUSTER>
            Connect the given cluster to the new stitch app. A group must be
            set (using -g/--group) to use this option.

    -g, --group <GROUP>
            Associate the app with the given MongoDB Cloud group.

    -i, --input <STITCH-CONFIG>
            Bootstrap the new app with the given config. Defaults to "stitch.json".
            If the file is not found, a default configuration is used.

    -n, --name <APP-NAME>
            Give the app a particular name. If left unspecified, one will be
            assigned.

    -y      Skip all confirmation prompts.
`,
}

var (
	createFlagSet *flag.FlagSet

	flagCreateConfig,
	flagCreateGroup,
	flagCreateCluster,
	flagCreateAppName string
)

func init() {
	createFlagSet = create.initFlags()
	createFlagSet.StringVarP(&flagCreateConfig, "input", "i", "stitch.json", "")
	createFlagSet.StringVarP(&flagCreateGroup, "group", "g", "", "")
	createFlagSet.StringVarP(&flagCreateCluster, "cluster", "c", "", "")
	createFlagSet.StringVarP(&flagCreateAppName, "name", "n", "", "")
}

func createRun() error {
	args := createFlagSet.Args()
	name, group, cluster, conf := flagCreateAppName, flagCreateGroup, flagCreateCluster, flagCreateConfig
	if len(args) > 0 {
		return errUnknownArg(args[0])
	}
	if group == "" && cluster != "" {
		return errorf("cannot supply cluster without also selecting a group (use --group)")
	}
	if group == "" {
		var err error
		group, err = config.DefaultGroup()
		if err != nil {
			return errorf("a group must be specified to create an app")
		}
	}

	if !config.LoggedIn() {
		return config.ErrNotLoggedIn
	}

	if _, err := os.Stat(conf); err != nil {
		fmt.Fprint(os.Stderr, "creating app from default config...\n")
		return nil
	}

	fmt.Fprintf(os.Stderr, "creating app from %q...\n", conf)
	raw, err := ioutil.ReadFile(conf)
	if err != nil {
		return err
	}
	userApp, err := app.Import(raw)
	if err != nil {
		return err
	}
	if userApp.Group != "" ||
		userApp.Name != "" ||
		userApp.ID != "" ||
		userApp.ClientID != "" {
		if !ui.Ask("the config already has app identification fields specified.\ncontinue anyway, ignoring those fields?") {
			return nil
		}
	}
	if createAppHasMongoDBServices(userApp) {
		if flagCreateCluster == "" {
			return errorf("cannot create using a config that specifies MongoDB clusters without providing a replacement cluster (use --group <GROUP> --cluster <CLUSTER>)")
		}
		if !ui.Ask("use the cluster you specified to substitute the cluster(s) used in the config? (if no, creation will fail)") {
			return nil
		}

		uri, err := atlas.GetClusterURI(group, cluster)
		if err != nil {
			return err
		}
		for i, service := range userApp.Services {
			if service.Type != "MongoDB" {
				continue
			}
			var config map[string]interface{}
			json.Unmarshal(service.Config, &config)
			config["uri"] = uri
			raw, _ := json.Marshal(config)
			userApp.Services[i].Config = raw
		}
	}
	userApp.Name = name
	userApp.Group = group

	// TODO: create an app using call to stitch admin SDK
	userApp.ID, userApp.ClientID = "fakeAppID", "fakeAppClientID-qwert"
	fmt.Fprintf(os.Stderr, "successfully created app %s.\n", name)

	if !ui.Ask(fmt.Sprintf("write new/updated config to %q?", flagCreateConfig)) {
		return nil
	}
	raw = userApp.Export()
	return ioutil.WriteFile(conf, raw, 0600)
}

func createAppHasMongoDBServices(userApp app.App) bool {
	for _, service := range userApp.Services {
		if service.Type == "MongoDB" {
			return true
		}
	}
	return false
}
