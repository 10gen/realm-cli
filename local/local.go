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
			{Type: "GitHub", Name: "my-github-service", Webhooks: []app.Webhook{
				{
					Name:     "my-github-webhook-1",
					ID:       "5678asdf7890zxcv",
					Output:   "object",
					Pipeline: `[{"service": "my-slack-service", "action": "post"}]`,
				},
				{Name: "my-github-webhook-2"},
			}, Rules: []app.ServiceRule{
				{
					Name: "my-github-rule",
					ID:   "1234tyui5678ghjk",
					Rule: `{"%%true": true}`,
				},
			}},
			{Type: "HTTP", Name: "my-http-service", Webhooks: []app.Webhook{
				{Name: "my-github-webhook-1"},
			}},
			{Type: "Slack", Name: "my-slack-service"},
			{Type: "Slack", Name: "my-other-slack-service"},
		},
		Pipelines: []app.Pipeline{
			{
				Name:        "my-pipe1",
				ID:          "3456erty8765kjhg",
				Output:      "array",
				Private:     true,
				SkipRules:   false,
				CanEvaluate: `{"%%true": true}`,
				Pipeline: `[{
	"service": "my-slack-service",
	"action": "post",
	"args": {
		"channel": {"%cond": {
			"if": {"%neq": ["%%args.channel", ""]},
			"then": "%%args.channel",
			"else": "general"
		}},
		"username": "dummybot",
		"text": "%%args.message"
	}
}]`,
				Parameters: []app.PipelineParameter{
					{"message", true},
					{"channel", false},
				},
			},
			{Name: "my-pipe2"},
		},
		Values: []app.Value{
			{Name: "s3bucket", Value: "my-s3-bucket"},
			{Name: "admin-phone-number", Value: "1234567890"},
			{Name: "config", Value: map[string]interface{}{"foo": "yes", "bar": "no"}},
		},
		AuthProviders: []app.AuthProvider{
			{
				Type:    "anon",
				Name:    "anonymous",
				ID:      "0987fdsa7654gfsd",
				Enabled: true,
			},
			{
				Type:               "email",
				Name:               "email",
				ID:                 "1289fhdj0921yuio",
				Enabled:            true,
				Metadata:           []string{"email"},
				DomainRestrictions: []string{"example.com"},
			},
			{
				Type:               "oauth",
				Name:               "facebook",
				ID:                 "0987fdsa7654gfsd",
				Enabled:            false,
				Metadata:           []string{"name", "picture", "email"},
				RedirectURIs:       []string{"domain.com/endpoint"},
				DomainRestrictions: []string{"facebook.com"},
				Config:             `{"clientId": "1234zxcv", "clientSecret": "qwer5678"}`,
			},
			{
				Type:    "apikey",
				Name:    "api-keys",
				Enabled: false,
			},
		},
	}
	return
}
