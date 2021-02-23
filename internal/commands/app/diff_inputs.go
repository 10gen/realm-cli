package app

import (
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/terminal"
)

const (
	flagIncludeDependencies      = "include-dependencies"
	flagIncludeDependenciesShort = "d"
	flagIncludeDependenciesUsage = "include to diff Realm app dependencies changes as well"
	flagIncludeHosting           = "include-hosting"
	flagIncludeHostingShort      = "s"
	flagIncludeHostingUsage      = "include to diff Realm app hosting changes as well"
)

func (i *diffInputs) Resolve(profile *cli.Profile, ui terminal.UI) error {
	if i.AppDirectory == "" {
		i.AppDirectory = profile.WorkingDirectory
	}
	return nil
}
