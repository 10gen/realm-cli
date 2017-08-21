package app

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
}

type Pipeline struct {
	Name string
}

type Value struct {
	Name  string
	Value interface{} // something json.Unmarshal would create
}

type AuthProvider struct {
	Name string
}
