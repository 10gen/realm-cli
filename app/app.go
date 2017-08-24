// Package app defined the App type which represents the configuration of a
// stitch application, and provides means for importing and exporting such a
// configuration.
package app

import "encoding/json"

// App contains all data relevant to the configuration of a stitch application.
type App struct {
	Group, Name, ID, ClientID string

	Services      []Service
	Pipelines     []Pipeline
	Values        []Value
	AuthProviders []AuthProvider
}

// Service contains data relevant to the configuration of any particular
// service associated with an App.
type Service struct {
	Type, Name, ID string
	Config         json.RawMessage
	Webhooks       []Webhook
	Rules          []ServiceRule
}

// Webhook contains data relevant to the configuration of any particular
// incoming webhook associated with a Service.
type Webhook struct {
	Name, ID, Output string
	Pipeline         json.RawMessage
}

// ServiceRule contains data relevant to the configuration of any particular
// rule associated with a Service.
type ServiceRule struct {
	Name, ID string
	Rule     json.RawMessage
}

// Pipeline contains data relevant to the configuration of any particular named
// pipeline associated with an App.
type Pipeline struct {
	Name, ID, Output      string
	Private, SkipRules    bool
	Parameters            []PipelineParameter
	CanEvaluate, Pipeline json.RawMessage
}

// PipelineParameter contains data relevant to the configuration of any
// particular argument for a named pipeline.
type PipelineParameter struct {
	Name     string
	Required bool
}

// Value contains data relevant to the configuration of any particular
// admin-defined "value" associated an App.
type Value struct {
	Name  string
	Value json.RawMessage
}

// AuthProvider contains data relevant to the configuration of any
// particular authentication provider associated with an App.
// Some of these fields are not applicable to all auth providers.
type AuthProvider struct {
	Type, Name, ID                             string
	Enabled                                    bool
	Metadata, DomainRestrictions, RedirectURIs []string
	Config                                     json.RawMessage
}
