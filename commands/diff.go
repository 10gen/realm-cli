package commands

import (
	"encoding/json"
	"fmt"

	"github.com/10gen/stitch-cli/app"
	"github.com/10gen/stitch-cli/config"
	"github.com/10gen/stitch-cli/local"

	flag "github.com/ogier/pflag"
)

var diff = &Command{
	Run:  diffRun,
	Name: "diff",
	ShortUsage: `
USAGE:
    stitch diff [-C <CONFIG>] [--help]
`,
	LongUsage: `See the difference between the local app configuration and its remote version.

OPTIONS:
    -C, --local-config <CONFIG>
            Set the stitch config file. Defaults to looking for stitch.json
            recursively up from the current working directory.
`,
}

var (
	diffFlagSet *flag.FlagSet
)

func init() {
	diffFlagSet = diff.initFlags()
}

func diffRun() error {
	args := diffFlagSet.Args()
	if len(args) > 0 {
		return errUnknownArg(args[0])
	}
	if !config.LoggedIn() {
		return config.ErrNotLoggedIn
	}
	localApp, ok := local.GetApp()
	if !ok {
		return errorf("could not find local app config (use -C to use a file other than \"stitch.json\")")
	}

	// TODO: use admin SDK to export and create an app.App
	// return errorf("could not find app with group %q and ID %q", localApp.Group, localApp.ID)
	remoteApp := app.App{
		// must not change
		Group:    localApp.Group,
		ID:       localApp.ID,
		ClientID: localApp.ClientID,
	}
	{ // mock data
		remoteApp.Name = "staging-" + localApp.Name
		for i, svc := range localApp.Services {
			switch i % 3 {
			case 1:
				continue
			case 0:
				svc.Name += "-old"
				if len(svc.Config) == 0 {
					localApp.Services[i].Config = json.RawMessage(`{"key":"value"}`)
				}
				svc.Config = json.RawMessage(`{"key":"old-value"}`)
				var whs []app.Webhook
				for j, wh := range svc.Webhooks {
					switch j % 3 {
					case 0:
						continue
					case 1:
						wh.Name += "-old"
					}
					whs = append(whs, wh)
				}
				svc.Webhooks = whs
			}
			remoteApp.Services = append(remoteApp.Services, svc)
		}
		remoteApp.Services = append(remoteApp.Services, app.Service{
			ID:   "599ed92c27e5d6207f0c7aca",
			Name: "my-deprecated-service",
		})
		for i, pipe := range localApp.Pipelines {
			switch i % 3 {
			case 1:
				continue
			case 0:
				pipe.Name += "-old"
				if len(pipe.Pipeline) == 0 {
					localApp.Pipelines[i].Pipeline = json.RawMessage(`[
						{"service":"my-favorite-service", "action": "my-favorite-action"}
					]`)
				}
				pipe.Pipeline = json.RawMessage(`[{"service":"old-service", "action": "dull-action"}]`)
				var pps []app.PipelineParameter
				for j, pp := range pipe.Parameters {
					switch j % 3 {
					case 0:
						pp.Required = !pp.Required
					case 1:
						pp.Name += "-old"
					case 2:
						continue
					}
					pps = append(pps, pp)
				}
				pipe.Parameters = pps
			}
			remoteApp.Pipelines = append(remoteApp.Pipelines, pipe)
		}
		remoteApp.Pipelines = append(remoteApp.Pipelines, app.Pipeline{
			ID:   "599ed92c27e5d6207f0c7acb",
			Name: "my-deprecated-pipeline",
		})
		for i, v := range localApp.Values {
			switch i % 3 {
			case 0:
				continue
			case 1:
				v.Value = json.RawMessage(`"old-value"`)
			}
			remoteApp.Values = append(remoteApp.Values, v)
		}
		for i, ap := range localApp.AuthProviders {
			switch i % 3 {
			case 0:
				continue
			case 1:
				ap.Name += "-old"
				ap.Enabled = !ap.Enabled
				md := ap.Metadata
				if len(ap.Metadata) > 0 {
					md = ap.Metadata[1:]
				}
				ap.Metadata = md
				ap.DomainRestrictions = append(ap.DomainRestrictions, "foo.com")
			}
			remoteApp.AuthProviders = append(remoteApp.AuthProviders, ap)
		}
	}

	if localApp.ClientID != remoteApp.ClientID {
		return errorf("client id is immutable, it was changed from %q to %q", remoteApp.ClientID, localApp.ClientID)
	}

	d := app.Diff(remoteApp, localApp)
	fmt.Println(d)
	return nil
}
