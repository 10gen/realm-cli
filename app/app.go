package app

import "encoding/json"

type App struct {
	Group, Name, ID, ClientID string

	Clusters      []Cluster
	Services      []Service
	Pipelines     []Pipeline
	Values        []Value
	AuthProviders []AuthProvider
}

type Cluster struct {
	Name, URI string
}

type Service struct {
	Type, Name string
	Webhooks   []Webhook
	Rules      []ServiceRule
}

type Webhook struct {
	Name, ID, Output string
	Pipeline         json.RawMessage
}

type ServiceRule struct {
	Name, ID string
	Rule     json.RawMessage
}

type Pipeline struct {
	Name, ID, Output      string
	Private, SkipRules    bool
	Parameters            []PipelineParameter
	CanEvaluate, Pipeline json.RawMessage
}

type PipelineParameter struct {
	Name     string
	Required bool
}

type Value struct {
	Name  string
	Value interface{} // something json.Unmarshal would create
}

type AuthProvider struct {
	Type, Name, ID                             string
	Enabled                                    bool
	Metadata, DomainRestrictions, RedirectURIs []string
	Config                                     json.RawMessage
}
