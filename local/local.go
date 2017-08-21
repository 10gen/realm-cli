package local

import "github.com/10gen/stitch-cli/app"

func GetApp() (a app.App, ok bool) {
	ok = true
	a = app.App{
		Group:    "group-1",
		Name:     "platespace-prod",
		ID:       "598dca3bede4017c35942841",
		ClientID: "platespace-prod-txplq",
		Clusters: []app.Cluster{
			{Name: "mongodb-atlas", URI: "mongodb://localhost:27017/"},
		},
		Services: []app.Service{
			{Type: "GitHub", Name: "my-github-service"},
			{Type: "HTTP", Name: "my-http-service"},
			{Type: "Slack", Name: "my-slack-service"},
			{Type: "Slack", Name: "my-other-slack-service"},
		},
		Pipelines: []app.Pipeline{
			{Name: "my-pipe1"},
			{Name: "my-pipe2"},
		},
		Values: []app.Value{
			{Name: "s3bucket", Value: "my-s3-bucket"},
			{Name: "admin-phone-number", Value: "1234567890"},
			{Name: "config", Value: map[string]interface{}{"foo": "yes", "bar": "no"}},
		},
		AuthProviders: []app.AuthProvider{
			{Name: "anonymous"},
			{Name: "email"},
			{Name: "facebook"},
			{Name: "api-keys"},
		},
	}
	return
}
