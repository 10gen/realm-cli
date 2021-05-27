package schema

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
)

func TestSchemaModels(t *testing.T) {
	t.Run("should return an error when client fails to find app", func(t *testing.T) {
		realmClient := mock.RealmClient{}
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return nil, errors.New("something bad happened")
		}

		cmd := &CommandDatamodels{datamodelsInputs{}}

		assert.Equal(t,
			errors.New("something bad happened"),
			cmd.Handler(nil, nil, cli.Clients{Realm: realmClient}),
		)
	})

	t.Run("should return an error when client fails to get schema models", func(t *testing.T) {
		realmClient := mock.RealmClient{}
		realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
			return []realm.App{{}}, nil
		}
		realmClient.SchemaModelsFn = func(groupID, appID, language string) ([]realm.SchemaModel, error) {
			return nil, errors.New("something bad happened")
		}

		cmd := &CommandDatamodels{datamodelsInputs{}}

		assert.Equal(t,
			errors.New("something bad happened"),
			cmd.Handler(nil, nil, cli.Clients{Realm: realmClient}),
		)
	})

	for _, tc := range []struct {
		description string
		flat        bool
		noImports   bool
		imports     [][]string
		output      string
	}{
		{
			description: "should print message when no models are found",
			output:      "No Typescript models were generated, check that you have defined a schema\n",
		},
		{
			description: "should print a named generated schema model",
			imports:     [][]string{{"import1\n", "import2\n"}},
			output: `The following Typescript data model was generated from your schema: RESOURCE_0

import1
import2

export const schema0 = {}

`,
		},
		{
			description: "should print a flat generated schema model",
			flat:        true,
			imports:     [][]string{{"import1\n", "import2\n"}},
			output: `The following Typescript data models were generated from your schema

import1
import2

export const schema0 = {}

`,
		},
		{
			description: "should print a flat list of generated schema models",
			flat:        true,
			imports: [][]string{
				{"import1\n", "import2\n"},
				{"import2\n", "import3\n"},
				{"import4\n", "import0\n"},
			},
			output: `The following Typescript data models were generated from your schema

import0
import1
import2
import3
import4

export const schema0 = {}

export const schema1 = {}

export const schema2 = {}

`,
		},
		{
			description: "should print a named list of generated schema models",
			imports: [][]string{
				{"import1\n", "import2\n"},
				{"import2\n", "import3\n"},
				{"import4\n", "import0\n"},
			},
			output: `The following Typescript data model was generated from your schema: RESOURCE_0

import1
import2

export const schema0 = {}

The following Typescript data model was generated from your schema: RESOURCE_1

import2
import3

export const schema1 = {}

The following Typescript data model was generated from your schema: RESOURCE_2

import0
import4

export const schema2 = {}

`,
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			realmClient := mock.RealmClient{}

			realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
				return []realm.App{{GroupID: "groupId", ID: "appId"}}, nil
			}

			var clientArgs struct {
				groupID, appID, language string
			}
			realmClient.SchemaModelsFn = func(groupID, appID, language string) ([]realm.SchemaModel, error) {
				clientArgs = struct {
					groupID, appID, language string
				}{groupID, appID, language}

				models := make([]realm.SchemaModel, len(tc.imports))
				for i, imports := range tc.imports {
					models[i] = realm.SchemaModel{
						ServiceID: fmt.Sprintf("sid%d", i), RuleID: fmt.Sprintf("rid%d", i),
						Name: fmt.Sprintf("resource_%d", i), Namespace: fmt.Sprintf("data.db.coll%d", i),
						Imports: imports,
						Code:    fmt.Sprintf("export const schema%d = {}\n", i),
					}
				}

				return models, nil
			}

			profile := mock.NewProfile(t)

			out, ui := mock.NewUI()

			cmd := &CommandDatamodels{datamodelsInputs{
				Flat:          tc.flat,
				NoImports:     tc.noImports,
				Language:      languageTypescript,
				ProjectInputs: cli.ProjectInputs{Project: "project", App: "test-app"},
			}}

			err := cmd.Handler(profile, ui, cli.Clients{Realm: realmClient})
			assert.Nil(t, err)

			assert.Equal(t, "groupId", clientArgs.groupID)
			assert.Equal(t, "appId", clientArgs.appID)
			assert.Equal(t, realm.DataModelLanguageTypescript, clientArgs.language)

			assert.Equal(t, tc.output, out.String())
		})
	}

	type codegen struct {
		error    realm.SchemaModelAlert
		warnings []realm.SchemaModelAlert
	}

	for _, tc := range []struct {
		description string
		models      []codegen
		output      string
	}{
		{
			description: "should print a list of all generated errors",
			models: []codegen{
				{
					error: realm.SchemaModelAlert{Code: "ErrCode", Message: "error"},
				},
				{},
				{
					error: realm.SchemaModelAlert{Code: "SomeCode", Message: "another error"},
				},
			},
			output: `The following Javascript data models were generated from your schema

export const schema0 = {}

export const schema1 = {}

export const schema2 = {}

The following collections have schemas with errors` + strings.Join([]string{"",
				"  Collection     Model Name  Error Code  Message        Link                                                      ",
				"  -------------  ----------  ----------  -------------  ----------------------------------------------------------",
				"  data.db.coll0  resource_0  ErrCode     error          /groups/groupId/apps/appId/clusters/sid0/rules/rid0/schema",
				"  data.db.coll2  resource_2  SomeCode    another error  /groups/groupId/apps/appId/clusters/sid2/rules/rid2/schema",
				""}, "\n"),
		},
		{
			description: "should print a list of all generated warnings",
			models: []codegen{
				{},
				{
					warnings: []realm.SchemaModelAlert{
						{Code: "WarningCode", Message: "warning"},
						{Code: "AnotherCode", Message: "another warning"},
					},
				},
				{},
			},
			output: `The following Javascript data models were generated from your schema

export const schema0 = {}

export const schema1 = {}

export const schema2 = {}

The following collections have schemas with warnings` + strings.Join([]string{"",
				"  Collection     Model Name  Error Code   Message          Link                                                      ",
				"  -------------  ----------  -----------  ---------------  ----------------------------------------------------------",
				"  data.db.coll1  resource_1  WarningCode  warning          /groups/groupId/apps/appId/clusters/sid1/rules/rid1/schema",
				"  data.db.coll1  resource_1  AnotherCode  another warning  /groups/groupId/apps/appId/clusters/sid1/rules/rid1/schema",
				""}, "\n"),
		},
		{
			description: "should print a combined list of all generated errors and warnings",
			models: []codegen{
				{
					error: realm.SchemaModelAlert{Code: "ErrCode", Message: "error"},
				},
				{
					warnings: []realm.SchemaModelAlert{
						{Code: "WarningCode", Message: "warning"},
						{Code: "AnotherCode", Message: "another warning"},
					},
				},
				{
					error: realm.SchemaModelAlert{Code: "SomeCode", Message: "another error"},
				},
			},
			output: `The following Javascript data models were generated from your schema

export const schema0 = {}

export const schema1 = {}

export const schema2 = {}

The following collections have schemas with errors
` + strings.Join([]string{
				"  Collection     Model Name  Error Code  Message        Link                                                      ",
				"  -------------  ----------  ----------  -------------  ----------------------------------------------------------",
				"  data.db.coll0  resource_0  ErrCode     error          /groups/groupId/apps/appId/clusters/sid0/rules/rid0/schema",
				"  data.db.coll2  resource_2  SomeCode    another error  /groups/groupId/apps/appId/clusters/sid2/rules/rid2/schema",
			}, "\n") + `
The following collections have schemas with warnings
` + strings.Join([]string{
				"  Collection     Model Name  Error Code   Message          Link                                                      ",
				"  -------------  ----------  -----------  ---------------  ----------------------------------------------------------",
				"  data.db.coll1  resource_1  WarningCode  warning          /groups/groupId/apps/appId/clusters/sid1/rules/rid1/schema",
				"  data.db.coll1  resource_1  AnotherCode  another warning  /groups/groupId/apps/appId/clusters/sid1/rules/rid1/schema",
				""}, "\n"),
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			realmClient := mock.RealmClient{}
			realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
				return []realm.App{{GroupID: "groupId", ID: "appId"}}, nil
			}
			realmClient.SchemaModelsFn = func(groupID, appID, language string) ([]realm.SchemaModel, error) {
				models := make([]realm.SchemaModel, len(tc.models))
				for i, m := range tc.models {
					models[i] = realm.SchemaModel{
						ServiceID: fmt.Sprintf("sid%d", i), RuleID: fmt.Sprintf("rid%d", i),
						Name: fmt.Sprintf("resource_%d", i), Namespace: fmt.Sprintf("data.db.coll%d", i),
						Imports:  []string{"import0\n"},
						Code:     fmt.Sprintf("export const schema%d = {}\n", i),
						Error:    m.error,
						Warnings: m.warnings,
					}
				}
				return models, nil
			}

			profile := mock.NewProfile(t)

			out, ui := mock.NewUI()

			cmd := &CommandDatamodels{datamodelsInputs{
				Flat:          true,
				NoImports:     true,
				Language:      languageJavascript,
				ProjectInputs: cli.ProjectInputs{Project: "project", App: "test-app"},
			}}

			err := cmd.Handler(profile, ui, cli.Clients{Realm: realmClient})
			assert.Nil(t, err)

			assert.Equal(t, tc.output, out.String())
		})
	}

	t.Run("should be able to filter responses", func(t *testing.T) {
		for _, tc := range []struct {
			description string
			names       []string
			flat        bool
			output      string
		}{
			{
				description: "by a single name",
				names:       []string{"two"},
				output: `The following Java data model was generated from your schema: TWO

import2
import3

export const schemaTwo = {}

`,
			},
			{
				description: "by many names",
				names:       []string{"one", "three"},
				flat:        true,
				output: `The following Java data models were generated from your schema

import0
import1
import2
import4

export const schemaOne = {}

export const schemaThree = {}

`,
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				realmClient := mock.RealmClient{}
				realmClient.FindAppsFn = func(filter realm.AppFilter) ([]realm.App, error) {
					return []realm.App{{}}, nil
				}
				realmClient.SchemaModelsFn = func(groupID, appID, language string) ([]realm.SchemaModel, error) {
					return []realm.SchemaModel{
						{
							ServiceID: "sid", RuleID: "rid",
							Name: "one", Namespace: "data.db.coll",
							Imports: []string{"import1\n", "import2\n"},
							Code:    "export const schemaOne = {}\n",
						},
						{
							ServiceID: "sid", RuleID: "rid",
							Name: "two", Namespace: "data.db.coll",
							Imports: []string{"import2\n", "import3\n"},
							Code:    "export const schemaTwo = {}\n",
						},
						{
							ServiceID: "sid", RuleID: "rid",
							Name: "three", Namespace: "data.db.coll",
							Imports: []string{"import4\n", "import0\n"},
							Code:    "export const schemaThree = {}\n",
						},
					}, nil
				}

				profile := mock.NewProfile(t)

				out, ui := mock.NewUI()

				i := datamodelsInputs{
					Flat:          tc.flat,
					Names:         tc.names,
					Language:      languageJava,
					ProjectInputs: cli.ProjectInputs{Project: "project", App: "test-app"},
				}
				assert.Nil(t, i.Resolve(profile, ui)) // sets empty nameSet

				cmd := &CommandDatamodels{i}

				err := cmd.Handler(profile, ui, cli.Clients{Realm: realmClient})
				assert.Nil(t, err)

				assert.Equal(t, tc.output, out.String())
			})
		}

	})
}

func TestLanguageType(t *testing.T) {
	for _, tc := range []struct {
		l        language
		expected string
	}{
		{l: language("eggcorn")},
		{languageCSharp, realm.DataModelLanguageCSharp},
		{languageJava, realm.DataModelLanguageJava},
		{languageJavascript, realm.DataModelLanguageJavascript},
		{languageKotlin, realm.DataModelLanguageKotlin},
		{languageObjectiveC, realm.DataModelLanguageObjectiveC},
		{languageSwift, realm.DataModelLanguageSwift},
		{languageTypescript, realm.DataModelLanguageTypescript},
	} {
		t.Run(fmt.Sprintf("should %s language to the correct realm type", tc.l), func(t *testing.T) {
			assert.Equal(t, tc.expected, languageType(tc.l))
		})
	}
}

func TestLanguageDisplay(t *testing.T) {
	for _, tc := range []struct {
		l        language
		expected string
	}{
		{l: language("eggcorn")},
		{languageCSharp, "C#"},
		{languageJava, "Java"},
		{languageJavascript, "Javascript"},
		{languageKotlin, "Kotlin"},
		{languageObjectiveC, "Objective-C"},
		{languageSwift, "Swift"},
		{languageTypescript, "Typescript"},
	} {
		t.Run(fmt.Sprintf("should %s language to the correct realm type", tc.l), func(t *testing.T) {
			assert.Equal(t, tc.expected, languageDisplay(tc.l))
		})
	}
}
