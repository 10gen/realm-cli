package flags

// set of supported flags
const (
	// globals
	AutoConfirm   = "yes"
	DisableColors = "disable-colors"
	OutputFormat  = "output-format"
	OutputTarget  = "output-target"
	Profile       = "profile"
	TelemetryMode = "telemetry"

	// global shorthands
	OutputTargetShort  = "o"
	ProfileShort       = "p"
	OutputFormatShort  = "t"
	AutoConfirmShort   = "y"
	TelemetryModeShort = "m"

	// auth
	App     = "app"
	Project = "project"

	PublicAPIKey  = "api-key"
	PrivateAPIKey = "private-api-key"

	// external
	RealmBaseURL = "base-url"
	CloudBaseURL = "cloud-base-url"
)

// set of supported flags' usages
// TODO(REALMC-7429): fill out the flag usages
const (
	PublicAPIKeyUsage  = "this is the --api-key usage"
	PrivateAPIKeyUsage = "this is the --private-api-key usage"

	AppUsage      = "this is the --app usage"
	ProjectUseage = "this is the --project usage"

	ProfileUsage = "this is the --profile, -p usage"

	RealmBaseURLUsage = "specify the base Realm server URL"
	CloudBaseURLUsage = "specify the base Atlas server URL"

	AutoConfirmUsage   = "set to automatically proceed through command confirmations"
	DisableColorsUsage = "disable output styling"
	OutputFormatUsage  = `set the output format, available options: [json]`
	OutputTargetUsage  = "write output to the specified filepath"
	TelemetryModeUsage = `enable or disable telemetry (this setting is remembered), available options: ["off", "on"]`
)
