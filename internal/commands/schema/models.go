package schema

import (
	"fmt"
	"sort"
	"strings"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/flags"
)

// CommandMetaDatamodels is the command meta for the `schema datamodels` command
var CommandMetaDatamodels = cli.CommandMeta{
	Use:         "datamodels",
	Aliases:     []string{"datamodel"},
	Display:     "schema datamodels",
	Description: "Generate data models based on your Schema",
	HelpText: `Translates your Schema’s objects into Realm data models. The data models define
your data as native objects, which can be easily integrated into your own repo
to use with Realm Sync.

NOTE: You must have a valid JSON Schema before using this command.

With this command, you can:
  - Specify the language with a "--language" flag
  - Filter which Schema objects you’d like to include in your output with "--name" flags
  - Combine your Schema objects into a single output with a "--flat" flag
  - Omit import groups from your model with a "--no-imports" flag`,
}

// CommandDatamodels is the `schema datamodels` command
type CommandDatamodels struct {
	inputs datamodelsInputs
}

// Flags is the command flags
func (cmd *CommandDatamodels) Flags() []flags.Flag {
	return []flags.Flag{
		cli.AppFlagWithContext(&cmd.inputs.App, "to generate its data models"),
		cli.ProjectFlag(&cmd.inputs.Project),
		cli.ProductFlag(&cmd.inputs.Products),
		flags.CustomFlag{
			Value: &cmd.inputs.Language,
			Meta: flags.Meta{
				Name:      "language",
				Shorthand: "l",
				Usage: flags.Usage{
					Description:   "Specify the language to generate schema data models in",
					DefaultValue:  "<none>",
					AllowedValues: []string{}, // TODO
				},
			},
		},
		flags.BoolFlag{
			Value: &cmd.inputs.Flat,
			Meta: flags.Meta{
				Name:  "flat",
				Usage: flags.Usage{Description: "View generated data models (and associated imports) as a single code block"},
			},
		},
		flags.BoolFlag{
			Value: &cmd.inputs.NoImports,
			Meta: flags.Meta{
				Name:  "no-imports",
				Usage: flags.Usage{Description: "View generated data models without imports"},
			},
		},
		flags.StringSliceFlag{
			Value: &cmd.inputs.Names,
			Meta: flags.Meta{
				Name:  "name",
				Usage: flags.Usage{Description: "Filter generated data models by name(s)"},
			},
		},
	}
}

// Inputs is the command inputs
func (cmd *CommandDatamodels) Inputs() cli.InputResolver {
	return &cmd.inputs
}

// Handler is the command handler
func (cmd *CommandDatamodels) Handler(profile *user.Profile, ui terminal.UI, clients cli.Clients) error {
	app, err := cli.ResolveApp(ui, clients.Realm, cli.AppOptions{
		AppMeta: cmd.inputs.AppMeta,
		Filter:  cmd.inputs.Filter(),
	})
	if err != nil {
		return err
	}

	models, err := clients.Realm.SchemaModels(app.GroupID, app.ID, languageType(cmd.inputs.Language))
	if err != nil {
		return err
	}

	if len(models) == 0 {
		ui.Print(terminal.NewTextLog("No %s models were generated, check that you have defined a schema", languageDisplay(cmd.inputs.Language)))
		return nil
	}

	if len(cmd.inputs.nameSet) > 0 {
		filtered := make([]realm.SchemaModel, 0, len(cmd.inputs.nameSet))
		for _, model := range models {
			if _, ok := cmd.inputs.nameSet[model.Name]; ok {
				filtered = append(filtered, model)
			}
		}
		models = filtered
	}

	var modelsWithError, modelsWithWarnings []realm.SchemaModel
	var modelsWithWarningsSize int

	// track all errors and warnings across models
	inspectModelAlerts := func(model realm.SchemaModel) {
		if model.Error.Code != "" && model.Error.Message != "" {
			modelsWithError = append(modelsWithError, model)
		}

		if len(model.Warnings) > 0 {
			modelsWithWarnings = append(modelsWithWarnings, model)
			modelsWithWarningsSize += len(model.Warnings)
		}
	}

	// produce deterministic,  generated code output (based on inputs)
	codeSnippet := func(imports []string, code string) string {
		if cmd.inputs.NoImports {
			return code
		}

		tmp := make([]string, len(imports))
		copy(tmp, imports)
		sort.SliceStable(importSorter(tmp))

		return strings.Join(tmp, "") + "\n" + code
	}

	logs := make([]terminal.Log, 0, len(models))

	if cmd.inputs.Flat {
		var allImports []string
		importsSet := map[string]struct{}{}

		allCodes := make([]string, 0, len(models))

		for _, model := range models {
			inspectModelAlerts(model)

			for _, imprt := range model.Imports {
				if _, ok := importsSet[imprt]; !ok {
					allImports = append(allImports, imprt)
					importsSet[imprt] = struct{}{}
				}
			}

			allCodes = append(allCodes, model.Code)
		}

		logs = append(logs, terminal.NewTextLog(
			"The following %s data models were generated from your schema\n\n%s",
			languageDisplay(cmd.inputs.Language),
			codeSnippet(allImports, strings.Join(allCodes, "\n")),
		))
	} else {
		for _, model := range models {
			inspectModelAlerts(model)

			logs = append(logs, terminal.NewTextLog(
				"The following %s data model was generated from your schema: %s\n\n%s",
				languageDisplay(cmd.inputs.Language),
				strings.ToUpper(model.Name),
				codeSnippet(model.Imports, model.Code),
			))
		}
	}

	// print data models
	ui.Print(logs...)

	// report any errors
	if len(modelsWithError) > 0 {
		rows := make([]map[string]interface{}, 0, len(modelsWithError))
		for _, model := range modelsWithError {
			rows = append(rows, modelAlertTableRow(profile, app, model, model.Error))
		}

		ui.Print(terminal.NewTableLog(
			"The following collections have schemas with errors",
			modelAlertsTableHeaders,
			rows...,
		))
	}

	// report any warnings
	if len(modelsWithWarnings) > 0 {
		rows := make([]map[string]interface{}, 0, modelsWithWarningsSize)
		for _, model := range modelsWithWarnings {
			for _, warning := range model.Warnings {
				rows = append(rows, modelAlertTableRow(profile, app, model, warning))
			}
		}

		ui.Print(terminal.NewTableLog(
			"The following collections have schemas with warnings",
			modelAlertsTableHeaders,
			rows...,
		))
	}

	return nil
}

const (
	modelWarningsTableHeaderCollection = "Collection"
	modelWarningsTableHeaderModelName  = "Model Name"
	modelWarningsTableHeaderErrorCode  = "Error Code"
	modelWarningsTableHeaderMessage    = "Message"
	modelWarningsTableHeaderLink       = "Link"
)

var (
	modelAlertsTableHeaders = []string{
		modelWarningsTableHeaderCollection,
		modelWarningsTableHeaderModelName,
		modelWarningsTableHeaderErrorCode,
		modelWarningsTableHeaderMessage,
		modelWarningsTableHeaderLink,
	}
)

func modelAlertTableRow(profile *user.Profile, app realm.App, model realm.SchemaModel, alert realm.SchemaModelAlert) map[string]interface{} {
	return map[string]interface{}{
		modelWarningsTableHeaderCollection: model.Namespace,
		modelWarningsTableHeaderModelName:  model.Name,
		modelWarningsTableHeaderErrorCode:  alert.Code,
		modelWarningsTableHeaderMessage:    alert.Message,
		modelWarningsTableHeaderLink: fmt.Sprintf(
			"%s/groups/%s/apps/%s/clusters/%s/rules/%s/schema",
			profile.RealmBaseURL(),
			app.GroupID,
			app.ID,
			model.ServiceID,
			model.RuleID,
		),
	}
}

func importSorter(imports []string) (interface{}, func(int, int) bool) {
	return imports, func(i, j int) bool {
		return imports[i] < imports[j]
	}
}

func languageType(l language) string {
	switch l {
	case languageCSharp:
		return realm.DataModelLanguageCSharp
	case languageJava:
		return realm.DataModelLanguageJava
	case languageJavascript:
		return realm.DataModelLanguageJavascript
	case languageKotlin:
		return realm.DataModelLanguageKotlin
	case languageObjectiveC:
		return realm.DataModelLanguageObjectiveC
	case languageSwift:
		return realm.DataModelLanguageSwift
	case languageTypescript:
		return realm.DataModelLanguageTypescript
	}
	return ""
}

func languageDisplay(l language) string {
	switch l {
	case languageCSharp:
		return "C#"
	case languageJava:
		return "Java"
	case languageJavascript:
		return "Javascript"
	case languageKotlin:
		return "Kotlin"
	case languageObjectiveC:
		return "Objective-C"
	case languageSwift:
		return "Swift"
	case languageTypescript:
		return "Typescript"
	}
	return ""
}
