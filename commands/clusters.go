package commands

import (
	"fmt"

	"github.com/10gen/stitch-cli/atlas"
	"github.com/10gen/stitch-cli/config"
	"github.com/10gen/stitch-cli/ui"
	flag "github.com/ogier/pflag"
)

var clusters = &Command{
	Run:  clustersRun,
	Name: "clusters",
	ShortUsage: `
USAGE:
    stitch clusters [--help] [<GROUP> [<CLUSTER>]]
`,
	LongUsage: `Show what Atlas clusters you can access.

ARGS:
    <GROUP>
            Show only clusters from the given MongoDB Cloud group.
    <CLUSTER>
            Show the URI of the specified Atlas cluster.
`,
}

var (
	clustersFlagSet *flag.FlagSet
)

func init() {
	clustersFlagSet = clusters.initFlags()
}

func clustersRun() error {
	args := clustersFlagSet.Args()
	var group, cluster string
	if len(args) > 0 {
		group = args[0]
		if !validGroup(group) {
			return errorf("invalid group %q.", group)
		}
		args = args[1:]
	}
	if len(args) > 0 {
		cluster = args[0]
		args = args[1:]
	}
	if len(args) > 0 {
		return errUnknownArg(args[0])
	}

	if !config.LoggedIn() {
		return config.ErrNotLoggedIn
	}
	if group == "" {
		css, err := clustersAll()
		if err != nil {
			return err
		}

		for i := range css {
			css[i].key = ui.Color(ui.Group, css[i].key)
		}
		printKV(css)
		return nil
	}
	if cluster == "" {
		cs, err := clustersForGroup(group)
		if err != nil {
			return err
		}
		printSingleKV(cs)
		return nil
	}
	uri, err := atlas.GetClusterURI(group, cluster)
	if err != nil {
		return err
	}
	fmt.Println(uri)
	return nil
}

func clustersAll() (items []kv, err error) {
	clusters, err := atlas.GetAllClusters()
	if err != nil {
		return
	}
	for group, cs := range clusters {
		items = append(items, kv{key: group, values: cs})
	}
	return
}

func clustersForGroup(group string) (item kv, err error) {
	clusters, err := atlas.GetClusters(group)
	if err != nil {
		return
	}
	clusterNames := make([]string, len(clusters))
	for i, cluster := range clusters {
		clusterNames[i] = cluster[0]
	}
	item = kv{key: group, values: clusterNames}
	return
}
