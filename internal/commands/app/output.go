package app

import "github.com/10gen/realm-cli/internal/cloud/realm"

type appOutput struct {
	app realm.App
	err error
}

const (
	headerID      = "ID"
	headerName    = "Name"
	headerDeleted = "Deleted"
	headerDetails = "Details"
)

type newAppOutputs struct {
	AppID               string              `json:"client_app_id"`
	Filepath            string              `json:"filepath"`
	URL                 string              `json:"url,omitempty"`
	Backend             string              `json:"backend,omitempty"`
	Frontends           string              `json:"frontends,omitempty"`
	Clusters            []dataSourceOutputs `json:"clusters,omitempty"`
	ServerlessInstances []dataSourceOutputs `json:"serverless_instances,omitempty"`
	Datalakes           []dataSourceOutputs `json:"datalakes,omitempty"`
}

type dataSourceOutputs struct {
	Name string `json:"name"`
}
