# Realm CLI Guidelines

The purpose of this README is to provide information around code patterns and concepts while writing and maintaining CLI commands.  It is meant to be a working agreement capable of enabling new contributors to quickly get up-to-speed on how to contribute effectively and allowing familiar contributors to continue to document the evolution of the codebase's best practices.

## Writing a New Command

When writing a new command, did you:

  - [ ] adhere to the patterns and best practices outlined below
  - [ ] define the command (or command's parent) in `internal/commands/commands.go`
  - [ ] add the command (or command's parent) to the CLI in `cmd/root.go` (and/or to its corresponding parent's `Subcommands` field)
  - [ ] add a test case for the command in `e2e/cli_test.go`

## Anatomy of a "Command"

Before diving into the various stages of a command's execution, there are few characteristics of commands (and all things command-related) worth noting.

When naming things, such as commands or flags, it is important to keep in mind a few things:

* make it easy to type
* make it consistent with other, similar names across the codebase
* consider auto-completion
* avoid uppercase letters (and/or making capitalization matter)

The first two points relate to each other and puts the user experience in focus.  Keeping things concise and consistent allow the user to build up a muscle-memory allowing them to use new features without thinking too much.  When considering auto-completion, remember that using a "top-down approach" to name construction will allow users to seamlessly "tab through" related concepts (e.g. name a flag `--ip-new` instead of `--new-ip`).  Lastly, there will be conflicts here if/when we need to follow patterns found in `mCLI`, but keep a preference towards naming things with all lowercase and only use the `-` delimiter when necessary (e.g. prefer `accesslist` over `access-list` because in an auto-complete scenario, there's no need to pause the completion halfway through the single noun "access list").

When building out commands, make sure you allow a user to achieve every possible, supported code path without requiring the use of any user interaction (read: input prompts).  The CLI supports a `--yes, -y` flag which will circumvent and accept any UI confirmation, so just make sure any UI prompts (e.g. inputs, selects, etc.) are also pieces of information that can be provided through flags.

> Note: This is to allow the CLI to be fully scriptable (think those using this within a CI/CD environment).

A command should be viewed as the combination of two separate stages: "input resolution" and "command handler".  It is through these two stages that a command is implemented and where command failures can occur.

### Command Inputs

Each command may have inputs.  The inputs will hold the fields required for a command's flags and can also implement logic to prompt the user for any missing, required information.

#### Flags

Flags are ultimately just fields on the input struct.  When defining them, use the `internal/utils/flags` utility package and the various `Flag` implementations to ensure consistent usage string formatting.

#### User Prompts

As mentioned before, take care to ensure each and every command is executable without the need for any manual user input.

##### Input Resolution

Command inputs may implement the `Resolve` function, which will always run prior to the command handler.  Failures here will trigger the CLI to print the command's usage/help text alongside the error message, so the errors returned here should be due to mis-use of CLI inputs (e.g. invalid values, conflicting flag values, etc.)

> Note: The `Resolve` interface should make this obvious, but avoid making any network or client calls in this stage.  Sometimes those calls may fail due to network issues, and as such wouldn't need to trigger printing the usage/help text along with the error message.

### The Handler

Every command implements the `Handler` function.  Failures here will assume the user provided valid inputs and so do not print the command's usage/help text alongside the error message.

It is acceptable to still use user prompts at this stage.  These prompts often require the use of an API call to get a list of available resources before presenting selectable options back to the user.

#### Feedback

It is imperative to provide concise and consistent messaging back to the user as the command executes.  When printing these messages, use the `internal/terminal` package and the various `Log` implementations to ensure consistent formatting as well as support the CLI's structured output format.

##### Long-Running Processes

Occasionally, the CLI will make a long-running request or use polling to wait for a certain state.  In these instances, make sure to present the user with a "loading icon" to indicate the CLI is not yet done working.

When choosing the status text to appear alongside the spinner, consider the following things:

* while the message can be updated to reflect progress, prefer having a consistent prefix to the message as it will appear less jarring to the user when switching between messages
* the message (or at least the consistent prefix to the message) should be in the "present continuous tense" (e.g. "Downloading assets...", "Deploying app...")
* avoid printing messages (with the "UI") _during_ the display of the spinner as it will lead to messages not appearing on their own line, instead defer the printing of any informational/warning messages until after the spinner has been stopped and removed

When the spinner is finally closed, make sure to follow with a `ui.Print(terminal.NewTextLog(...))` where the message becomes the "past tense" version of the earlier status message (or at least the consistent prefix to the status message).
