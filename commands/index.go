package commands

import "fmt"

// Index is the command for the root stitch command.
var Index Command = &SuperCommand{
	Command: &SimpleCommand{
		N: "stitch",
		F: func(args []string) error {
			if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
				return ErrShowHelp
			}
			return fmt.Errorf("stitch: %q is not a stitch command.", args[0])
		},
	},
	SubCommandHelp: []struct{ Name, Help string }{
		{"login", "Authenticate as an administrator"},
		{"logout", "Deauthenticate"},
		{"me", "Show your user admin info"},
		{"groups", "Show what groups you are a part of"},
		{"apps", "Show what apps you can administrate"},
		{"info", "Show info about a particular app"},
		{"clusters", "Show what Atlas clusters you can access"},
		{"clone", "Export a stitch app"},
		{"create", "Create a new stitch app"},
		{"sync", "Push changes made locally to a stitch app configuration"},
		{"diff", "See the difference between the local app configuration and its remote version"},
		{"validate", "Validate the local app configuration"},
		{"migrate", "Migrate to a new version of the configuration spec"},
	},
	SubCommands: map[string]Command{
		"login": login,
	},
}
