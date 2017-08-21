package commands

import (
	"encoding/json"
	"fmt"

	"github.com/10gen/stitch-cli/app"
	"github.com/10gen/stitch-cli/local"
	"github.com/10gen/stitch-cli/ui"
	flag "github.com/ogier/pflag"
)

var info = &Command{
	Run:  infoRun,
	Name: "info",
	ShortUsage: `
USAGE:
    stitch info [--help] [--app <APP-ID>] [<TOP-LEVEL-SPEC> [<INNER-SPEC>...]]
`,
	LongUsage: `Show info about a particular app.

ARGS:
    <TOP-LEVEL-SPEC>
            The top level specifier, one of "group", "name", "id", "client-id",
            "clusters", "services", "pipelines", "values", or "authentication".
            Gives specific information on the given specifier.
   <INNER-SPEC>
            An inner specifier, according to whatever keys are made availables
            for the particular top level specifier:
            clusters: list of clusters
                - name of a cluster: MongoDB URI
            services: list of services
                - name of a service. Within a specified services:
                    - type: type of service
                    - name: name of service
                    - webhooks: list of webhooks
                        - name of a webhook. Within a specified webhook:
                            - id: identifer of the webhook
                            - name: name of the webhook
                            - output: output type of the webhook
                            - pipeline: JSON of the webhook's pipeline
                    - rules: list of rules
                        - name of a rule. Within a specified webhook:
                            - id: identifer of the rule
                            - name: name of the rule
                            - rule: JSON of the complete rule
            pipelines: list of named pipelines
                - name of pipeline. Within a specified pipeline:
                    - id: identifer of the pipeline
                    - name: name of pipeline
                    - output: output type of the pipeline
                    - private: whether the pipeline is private
                    - canEvaluate: JSON of the evaluation condition on the pipeline
                    - skipRules: whether the pipeline skips rules
                    - pipeline: JSON of the pipeline
                    - parameters: list of parameters to the pipeline
                    - values: list of value names associated with the pipeline
                        - name of value: assigned value
            - values: list of value names
                - name of value: assigned value
            - authentication: list of configure authentication providers
                - name of auth provider. Within a specified auth provider,
                  available inner specifiers depends on the particular provider.

OPTIONS:
    --app <APP-ID>
            Shows info for the specified app according to the stitch server.
            Leave unset to use local config.

    --json
            Show info in JSON form.
`,
}

var (
	infoFlagSet *flag.FlagSet

	flagInfoApp  string
	flagInfoJSON bool
)

func init() {
	infoFlagSet = info.InitFlags()
	infoFlagSet.StringVar(&flagInfoApp, "app", "", "")
	infoFlagSet.BoolVar(&flagInfoJSON, "json", false, "")
}

func infoRun() error {
	args := infoFlagSet.Args()
	if len(args) == 0 {
		return infoAll()
	}
	switch args[0] {
	case "clusters":
		return infoClusters(args[1:])
	case "services":
		return infoServices(args[1:])
	case "pipelines":
		return infoPipelines(args[1:])
	case "values":
		return infoValues(args[1:])
	case "authentication":
		return infoAuthentication(args[1:])
	default:
		if len(args) > 1 {
			return errorUnknownArg(args[2])
		}
		return infoItem(args[0])
	}
}

func infoGetApp() (a app.App, isLocal bool, err error) {
	if flagInfoApp == "" {
		a, isLocal = local.GetApp()
		if !isLocal {
			err = errorf("no local project found and --app was not specified.")
		}
		return
	}
	// TODO
	err = errorf("could not find app %q", flagInfoApp)
	return
}

func infoAll() error {
	app, isLocal, err := infoGetApp()
	if err != nil {
		return err
	}

	var clusters, pipelines, values, authentication []string
	var servicesInfo [][2]string

	for _, cluster := range app.Clusters {
		clusters = append(clusters, cluster.Name)
	}
	for _, pipeline := range app.Pipelines {
		pipelines = append(pipelines, pipeline.Name)
	}
	for _, value := range app.Values {
		values = append(values, value.Name)
	}
	for _, authProvider := range app.AuthProviders {
		authentication = append(authentication, authProvider.Name)
	}
	for _, service := range app.Services {
		servicesInfo = append(servicesInfo, [2]string{service.Type, service.Name})
	}

	if flagInfoJSON {
		servicesJSON := make([]interface{}, len(servicesInfo))
		for i, serviceInfo := range servicesInfo {
			servicesJSON[i] = map[string]interface{}{
				"type": serviceInfo[0],
				"name": serviceInfo[1],
			}
		}
		obj := map[string]interface{}{
			"group":          app.Group,
			"name":           app.Name,
			"id":             app.ID,
			"client-id":      app.ClientID,
			"clusters":       clusters,
			"services":       servicesJSON,
			"pipelines":      pipelines,
			"values":         values,
			"authentication": authentication,
		}
		if isLocal {
			obj["local"] = "yes"
		}
		raw, _ := json.Marshal(obj)
		fmt.Printf("%s\n", raw)
	} else {
		var items []kv
		if isLocal {
			items = append(items, kv{key: "local", value: "yes"})
		}
		for i := range servicesInfo {
			servicesInfo[i][0] = ui.Color(ui.ServiceType, servicesInfo[i][0])
		}
		items = append(items,
			kv{key: "group", value: ui.Color(ui.Group, app.Group)},
			kv{key: "name", value: app.Name},
			kv{key: "id", value: app.ID},
			kv{key: "client-id", value: ui.Color(ui.AppClientID, app.ClientID)},
			kv{key: "clusters", values: clusters},
			kv{key: "services", valuePairs: servicesInfo},
			kv{key: "pipelines", values: pipelines},
			kv{key: "values", values: values},
			kv{key: "authentication", values: authentication},
		)
		printKV(items)
	}
	return nil
}

func infoItem(item string) error {
	app, _, err := infoGetApp()
	if err != nil {
		return err
	}
	var output string
	var v ui.Variant = ui.None
	switch item {
	case "group":
		output = app.Group
		v = ui.Group
	case "name":
		output = app.Name
	case "id":
		output = app.ID
	case "client-id":
		output = app.ClientID
		v = ui.AppClientID
	default:
		return errorUnknownArg(item)
	}
	if flagInfoJSON {
		raw, _ := json.Marshal(output)
		fmt.Printf("%s\n", raw)
	} else {
		fmt.Println(ui.Color(v, output))
	}
	return nil
}

func infoClusters(args []string) error {
	var clusterName string
	if len(args) > 0 {
		clusterName = args[0]
		args = args[1:]
	}
	if len(args) > 0 {
		return errorUnknownArg(args[0])
	}

	app, _, err := infoGetApp()
	if err != nil {
		return err
	}

	if clusterName != "" {
		for _, cluster := range app.Clusters {
			if cluster.Name == clusterName {
				if flagInfoJSON {
					raw, _ := json.Marshal(cluster.URI)
					fmt.Printf("%s\n", raw)
				} else {
					fmt.Println(cluster.URI)
				}
				return nil
			}
		}
		return errorf("cluster %q not found.", clusterName)
	}
	var clusterNames []string
	for _, cluster := range app.Clusters {
		clusterNames = append(clusterNames, cluster.Name)
	}
	if flagInfoJSON {
		raw, _ := json.Marshal(clusterNames)
		fmt.Printf("%s\n", raw)
	} else {
		for _, clusterName := range clusterNames {
			fmt.Println(clusterName)
		}
	}
	return nil
}

func infoServices(args []string) error {
	if len(args) > 1 {
		return infoServicesParticular(args)
	}
	app, _, err := infoGetApp()
	if err != nil {
		return err
	}

	var services [][2]string
	for _, service := range app.Services {
		services = append(services, [2]string{service.Type, service.Name})
	}
	if flagInfoJSON {
		raw, _ := json.Marshal(services)
		fmt.Printf("%s\n", raw)
	} else {
		for i := range services {
			services[i][0] = ui.Color(ui.ServiceType, services[i][0])
		}
		printSingleKV(kv{valuePairs: services})
	}
	return nil
}

func infoServicesParticular(args []string) error {
	return nil
}

func infoPipelines(args []string) error {
	return nil
}

func infoValues(args []string) error {
	return nil
}

func infoAuthentication(args []string) error {
	return nil
}
