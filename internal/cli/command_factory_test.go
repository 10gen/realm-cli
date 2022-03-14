package cli

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/10gen/realm-cli/internal/telemetry"
	"github.com/10gen/realm-cli/internal/utils/api"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"
)

type capturedEvent struct {
	EventType telemetry.EventType
	Data      []telemetry.EventData
}

func TestCommandFactoryVersionCheck(t *testing.T) {
	now := time.Now()

	t.Run("should not warn the user nor update the last version check when client fails to get build info", func(t *testing.T) {
		profile := mock.NewProfile(t)

		lastVersionCheck := profile.LastVersionCheck().AddDate(0, 0, -2)
		profile.SetLastVersionCheck(lastVersionCheck)

		var events []capturedEvent
		var telemetryService mock.TelemetryService
		telemetryService.TrackEventFn = func(eventType telemetry.EventType, data ...telemetry.EventData) {
			events = append(events, capturedEvent{eventType, data})
		}

		out, ui := mock.NewUI()

		factory := &CommandFactory{
			profile:          profile,
			telemetryService: telemetryService,
			ui:               ui,
		}
		factory.checkForNewVersion(mockVersionClient{status: http.StatusNotFound})

		assert.Equal(t, 1, len(events))
		assert.Equal(t,
			capturedEvent{
				telemetry.EventTypeCommandError,
				telemetry.EventDataError(api.ErrUnexpectedStatusCode{"get cli version manifest", 404}),
			},
			events[0],
		)

		assert.Equal(t, "", out.String())
		assert.Equal(t, lastVersionCheck, profile.LastVersionCheck())
	})

	t.Run("should not warn the user but should update the last version check when client gets build info equal to current", func(t *testing.T) {
		profile := mock.NewProfile(t)
		lastVersionCheck := profile.LastVersionCheck()

		var events []capturedEvent
		var telemetryService mock.TelemetryService
		telemetryService.TrackEventFn = func(eventType telemetry.EventType, data ...telemetry.EventData) {
			events = append(events, capturedEvent{eventType, data})
		}

		out, ui := mock.NewUI()

		factory := &CommandFactory{
			profile:          profile,
			telemetryService: telemetryService,
			ui:               ui,
		}
		factory.checkForNewVersion(mockVersionClient{url: "http://somewhere.com"})

		assert.Equal(t, 0, len(events))
		assert.Equal(t, "", out.String())
		assert.True(t, lastVersionCheck.Before(profile.LastVersionCheck()), "version check time should be updated")
	})

	t.Run("should warn the user and update the last version check when client gets new build info", func(t *testing.T) {
		profile := mock.NewProfile(t)
		lastVersionCheck := profile.LastVersionCheck()

		var events []capturedEvent
		var telemetryService mock.TelemetryService
		telemetryService.TrackEventFn = func(eventType telemetry.EventType, data ...telemetry.EventData) {
			events = append(events, capturedEvent{eventType, data})
		}

		out, ui := mock.NewUI()

		factory := &CommandFactory{
			profile:          profile,
			telemetryService: telemetryService,
			ui:               ui,
		}
		factory.checkForNewVersion(mockVersionClient{version: "0.1.0", url: "http://somewhere.com"})

		assert.Equal(t, 1, len(events))
		assert.Equal(t, capturedEvent{EventType: telemetry.EventTypeCommandVersionCheck}, events[0])

		assert.Equal(t, `New version (v0.1.0) of CLI available: http://somewhere.com
Note: This is the only time this alert will display today
To install
  npm install -g mongodb-realm-cli@v0.1.0
  curl -o ./realm-cli http://somewhere.com && chmod +x ./realm-cli
`, out.String())

		assert.True(t, lastVersionCheck.Before(profile.LastVersionCheck()), "version check time should be updated")
	})

	t.Run("should not warn user nor update the last version check if the check has already occurred that day", func(t *testing.T) {
		lastVersionCheck := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

		profile := mock.NewProfile(t)
		profile.SetLastVersionCheck(lastVersionCheck)

		var events []capturedEvent
		var telemetryService mock.TelemetryService
		telemetryService.TrackEventFn = func(eventType telemetry.EventType, data ...telemetry.EventData) {
			events = append(events, capturedEvent{eventType, data})
		}

		out, ui := mock.NewUI()

		factory := &CommandFactory{
			profile:          profile,
			telemetryService: telemetryService,
			ui:               ui,
		}
		factory.checkForNewVersion(mockVersionClient{})

		assert.Equal(t, 0, len(events))
		assert.Equal(t, "", out.String())
		assert.Equal(t, lastVersionCheck, profile.LastVersionCheck())
	})
}

type mockVersionClient struct {
	status  int
	version string
	url     string
}

func (client mockVersionClient) Get(_ string) (*http.Response, error) {
	status := client.status
	if status == 0 {
		status = http.StatusOK
	}

	version := client.version
	if version == "" {
		version = Version
	}

	url := client.url
	if url == "" {
		url = "http://testurl.com"
	}

	return &http.Response{
		StatusCode: status,
		Body: ioutil.NopCloser(strings.NewReader(fmt.Sprintf(`{
  "version": %q,
  "info": { %q: { "url": %q } }
}`, version, OSArch, url))),
	}, nil
}
