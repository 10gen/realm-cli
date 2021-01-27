package app

import (
	"fmt"
)

// File represents a file of the Realm app structure
type File struct {
	Parent fmt.Stringer
	Name   string
	Ext    string
}

func (f File) String() string {
	name := f.Name + f.Ext
	if f.Parent == nil {
		return name
	}
	return f.Parent.String() + name
}

// Dir represents a directory of the Realm app structure
type Dir struct {
	Parent fmt.Stringer
	Name   string
}

func (d Dir) String() string {
	name := d.Name + "/"
	if d.Parent == nil {
		return name
	}
	return d.Parent.String() + name
}

// set of app structure filepath names
const (
	extJS   = ".js"
	extJSON = ".json"

	NameAuthProviders    = "auth_providers"
	NameConfig           = "config"
	NameCustomResolvers  = "custom_resolvers"
	NameFunctions        = "functions"
	NameGraphQL          = "graphql"
	NameIncomingWebhooks = "incoming_webhooks"
	NameRealmConfig      = "realm_config"
	NameRules            = "rules"
	NameSecrets          = "secrets"
	NameServices         = "services"
	NameSource           = "source"
	NameStitch           = "stitch"
	NameTriggers         = "triggers"
	NameValues           = "values"
)

// set of app structure files and directories
var (
	FileConfig      = File{Name: NameConfig, Ext: extJSON}
	FileRealmConfig = File{Name: NameRealmConfig, Ext: extJSON}
	FileSource      = File{Name: NameSource, Ext: extJS}
	FileStitch      = File{Name: NameStitch, Ext: extJSON}

	FileSecrets = File{Name: NameSecrets, Ext: extJSON}

	DirAuthProviders = Dir{Name: NameAuthProviders}

	DirFunctions = Dir{Name: NameFunctions}

	DirGraphQL                = Dir{Name: NameGraphQL}
	FileGraphQLConfig         = File{Parent: DirGraphQL, Name: NameConfig, Ext: extJSON}
	DirGraphQLCustomResolvers = Dir{Parent: DirGraphQL, Name: NameCustomResolvers}
)

// FileAuthProvider creates the auth provider config filepath
func FileAuthProvider(name string) File {
	return File{Parent: DirAuthProviders, Name: name, Ext: extJSON}
}

// TODO(REALMC-7865): resolve these two code paths based on how import requests work now
// v1 here is how the past CLI read the file system for app data
// v2 is meant to accommodate any necessary logic for the time being while using 20210101 (see REALMC-7653)

func readAppStructureV1(appDir Directory) (map[string]interface{}, error) {
	pkg := map[string]interface{}{}
	if err := unmarshalAppConfigInto(appDir, &pkg); err != nil {
		return nil, err
	}
	if err := unmarshalSecretsInto(appDir, pkg); err != nil {
		return nil, err
	}
	if err := unmarshalDirectoryInto(appDir.Path, NameValues, pkg); err != nil {
		return nil, err
	}
	if err := unmarshalDirectoryInto(appDir.Path, NameAuthProviders, pkg); err != nil {
		return nil, err
	}
	if err := unmarshalFunctionsInto(appDir.Path, NameFunctions, pkg); err != nil {
		return nil, err
	}
	if err := unmarshalDirectoryInto(appDir.Path, NameTriggers, pkg); err != nil {
		return nil, err
	}
	if err := unmarshalGraphQLInto(appDir, pkg); err != nil {
		return nil, err
	}
	if err := unmarshalServicesInto(appDir, pkg); err != nil {
		return nil, err
	}
	return pkg, nil
}

func readAppStructureV2(appDir Directory) (map[string]interface{}, error) {
	pkg := map[string]interface{}{}
	if err := unmarshalAppConfigInto(appDir, &pkg); err != nil {
		return nil, err
	}
	if err := unmarshalSecretsInto(appDir, pkg); err != nil {
		return nil, err
	}
	if err := unmarshalDirectoryInto(appDir.Path, NameValues, pkg); err != nil {
		return nil, err
	}
	if err := unmarshalDirectoryInto(appDir.Path, NameAuthProviders, pkg); err != nil {
		return nil, err
	}
	if err := unmarshalFunctionsInto(appDir.Path, NameFunctions, pkg); err != nil {
		return nil, err
	}
	if err := unmarshalDirectoryInto(appDir.Path, NameTriggers, pkg); err != nil {
		return nil, err
	}
	if err := unmarshalGraphQLInto(appDir, pkg); err != nil {
		return nil, err
	}
	if err := unmarshalServicesInto(appDir, pkg); err != nil {
		return nil, err
	}
	return pkg, nil
}
