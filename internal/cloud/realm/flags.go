package realm

import "github.com/10gen/realm-cli/internal/telemetry"

// set of default flag values
const (
	DefaultBaseURL       = "https://realm.mongodb.com"
	DefaultTelemetryType = string(telemetry.OnDefault)
)
