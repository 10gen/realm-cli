package commands

import (
	"github.com/10gen/realm-cli/internal/commands/app"
	"github.com/10gen/realm-cli/internal/commands/login"
	"github.com/10gen/realm-cli/internal/commands/logout"
	"github.com/10gen/realm-cli/internal/commands/user"
	"github.com/10gen/realm-cli/internal/commands/whoami"
)

// set of supported CLI commands
var (
	AppCommand    = app.Command
	LoginCommand  = login.Command
	LogoutCommand = logout.Command
	UserCommand   = user.Command
	WhoamiCommand = whoami.Command
)
