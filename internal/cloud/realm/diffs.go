package realm

import (
	"fmt"

	"github.com/10gen/realm-cli/internal/terminal"
)

// AppDraftDiff are the diffs for a Realm app draft and its corresponding app
type AppDraftDiff struct {
	Diffs             []string          `json:"diffs"`
	HostingFilesDiff  HostingFilesDiff  `json:"hosting_files_diff"`
	DependenciesDiff  DependenciesDiff  `json:"dependencies_diff"`
	GraphQLConfigDiff GraphQLConfigDiff `json:"graphql_config_diff"`
	SchemaOptionsDiff SchemaOptionsDiff `json:"schema_options_diff"`
}

// DiffList returns the diffs as a list of interface{}
func (d AppDraftDiff) DiffList() []interface{} {
	diffs := make([]interface{}, len(d.Diffs))
	for i, diff := range d.Diffs {
		diffs[i] = diff
	}
	return diffs
}

// HasChanges returns whether the diff has any changes
func (d AppDraftDiff) HasChanges() bool {
	return d.Len() > 0
}

// Len returns the number of changes the diff has
func (d AppDraftDiff) Len() int {
	return len(d.Diffs) +
		d.HostingFilesDiff.Len() +
		d.DependenciesDiff.Len() +
		d.GraphQLConfigDiff.Len() +
		d.SchemaOptionsDiff.Len()
}

// HostingFilesDiff are the diffs for a Realm app's static hosting setup
type HostingFilesDiff struct {
	Added    []string `json:"added"`
	Deleted  []string `json:"deleted"`
	Modified []string `json:"modified"`
}

// DiffList returns the diffs as a list of interface{}
func (d HostingFilesDiff) DiffList() []interface{} {
	diffs := make([]interface{}, 0, d.Len())
	for _, added := range d.Added {
		diffs = append(diffs, "added: "+added)
	}
	for _, deleted := range d.Deleted {
		diffs = append(diffs, "deleted: "+deleted)
	}
	for _, modified := range d.Modified {
		diffs = append(diffs, "modified: "+modified)
	}
	return diffs
}

// HasChanges returns whether the diff has any changes
func (d HostingFilesDiff) HasChanges() bool {
	return d.Len() > 0
}

// Len returns the number of changes the diff has
func (d HostingFilesDiff) Len() int {
	return len(d.Added) + len(d.Deleted) + len(d.Modified)
}

// DependenciesDiff are the diffs for a Realm app's dependencies
type DependenciesDiff struct {
	Added    []DependencyData     `json:"added"`
	Deleted  []DependencyData     `json:"deleted"`
	Modified []DependencyDiffData `json:"modified"`
}

// DiffList returns the diffs as a list of interface{}
func (d DependenciesDiff) DiffList() []interface{} {
	diffs := make([]interface{}, 0, d.Len())
	for _, added := range d.Added {
		diffs = append(diffs, "+ "+added.String())
	}
	for _, deleted := range d.Deleted {
		diffs = append(diffs, "- "+deleted.String())
	}
	for _, modified := range d.Modified {
		diffs = append(diffs, modified)
	}
	return diffs
}

// Cap returns the dependencies diffs' total capacity
func (d DependenciesDiff) Cap() int {
	return d.Len() + 3
}

// Strings returns the diffs as a list of strings
func (d DependenciesDiff) Strings() []string {
	diffs := make([]string, 0, d.Cap())
	if len(d.Added) > 0 {
		diffs = append(diffs, "Added Dependencies")
		for _, dep := range d.Added {
			diffs = append(diffs, terminal.Indent+"+ "+dep.String())
		}
	}
	if len(d.Deleted) > 0 {
		diffs = append(diffs, "Removed Dependencies")
		for _, dep := range d.Deleted {
			diffs = append(diffs, terminal.Indent+"- "+dep.String())
		}
	}
	if len(d.Modified) > 0 {
		diffs = append(diffs, "Modified Dependencies")
		for _, dep := range d.Modified {
			diffs = append(diffs, terminal.Indent+"* "+dep.String())
		}
	}
	return diffs
}

// HasChanges returns whether the diff has any changes
func (d DependenciesDiff) HasChanges() bool {
	return d.Len() > 0
}

// Len returns the number of changes the diff has
func (d DependenciesDiff) Len() int {
	return len(d.Added) + len(d.Deleted) + len(d.Modified)
}

// GraphQLConfigDiff are the diffs for a Realm app's GraphQL setup
type GraphQLConfigDiff struct {
	FieldDiffs []FieldDiff `json:"field_diffs"`
}

// DiffList returns the diffs as a list of interface{}
func (d GraphQLConfigDiff) DiffList() []interface{} {
	diffs := make([]interface{}, d.Len())
	for i, diff := range d.FieldDiffs {
		diffs[i] = diff
	}
	return diffs
}

// HasChanges returns whether the diff has any changes
func (d GraphQLConfigDiff) HasChanges() bool {
	return d.Len() > 0
}

// Len returns the number of changes the diff has
func (d GraphQLConfigDiff) Len() int {
	return len(d.FieldDiffs)
}

// SchemaOptionsDiff are the diffs for a Realm app's schema
type SchemaOptionsDiff struct {
	GraphQLValidationDiffs []FieldDiff `json:"graphql_validation_diff"`
	RestValidationDiffs    []FieldDiff `json:"rest_validation_diff"`
}

// DiffList returns the diffs as a list of interface{}
func (d SchemaOptionsDiff) DiffList() []interface{} {
	diffs := make([]interface{}, 0, d.Len())
	for _, diff := range d.GraphQLValidationDiffs {
		diffs = append(diffs, diff)
	}
	for _, diff := range d.RestValidationDiffs {
		diffs = append(diffs, diff)
	}
	return diffs
}

// HasChanges returns whether the diff has any changes
func (d SchemaOptionsDiff) HasChanges() bool {
	return d.Len() > 0
}

// Len returns the number of changes the diff has
func (d SchemaOptionsDiff) Len() int {
	return len(d.GraphQLValidationDiffs) + len(d.RestValidationDiffs)
}

// DependencyData is the data for an external Javascript dependency
type DependencyData struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func (d DependencyData) String() string {
	return fmt.Sprintf("%s@%s", d.Name, d.Version)
}

// DependencyDiffData is the data for an external Javascript dependency which had its version changed
type DependencyDiffData struct {
	DependencyData
	PreviousVersion string `json:"previous_version"`
}

func (d DependencyDiffData) String() string {
	return fmt.Sprintf("%s@%s -> %s", d.Name, d.PreviousVersion, d.DependencyData.String())
}

// FieldDiff is a the data for an updated field of a Schema or GraphQL config
type FieldDiff struct {
	Field         string      `json:"field_name"`
	PreviousValue interface{} `json:"previous"`
	UpdatedValue  interface{} `json:"updated"`
}

func (f FieldDiff) String() string {
	return fmt.Sprintf("%s: %s -> %s", f.Field, f.PreviousValue, f.UpdatedValue)
}
