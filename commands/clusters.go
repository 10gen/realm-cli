package commands

import (
	"fmt"

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
	clustersFlagSet = clusters.InitFlags()
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
		return errorUnknownArg(args[0])
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
	uri, err := clusterURI(group, cluster)
	if err != nil {
		return err
	}
	fmt.Println(uri)
	return nil
}

func clustersAll() ([]kv, error) {
	// TODO
	// should all use {key, values}
	return []kv{
		{key: "group-1", values: []string{"cluster0", "cluster1"}},
		{key: "group-2", values: []string{"clustera", "clusterb"}},
	}, nil
}

func clustersForGroup(group string) (kv, error) {
	// TODO
	switch group {
	case "group-1":
		return kv{key: group, values: []string{"cluster0", "cluster1"}}, nil
	case "group-2":
		return kv{key: group, values: []string{"clustera", "clusterb"}}, nil
	}
	return kv{}, errorNotInGroup(group)
}
