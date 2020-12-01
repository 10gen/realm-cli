package flags

// set of supported flags
const (
	PublicAPIKey  = "api-key"
	PrivateAPIKey = "private-api-key"

	Profile      = "profile"
	ProfileShort = "p"

	RealmBaseURL = "base-url"
	CloudBaseURL = "cloud-base-url"

	AutoConfirm      = "yes"
	AutoConfirmShort = "y"
	DisableColors    = "disable-colors"
	OutputFormat     = "output-format"
	OutputTarget     = "output-target"
)

// set of supported flags' usages
// TODO(REALMC-7429): fill out the flag usages
const (
	PublicAPIKeyUsage  = "this is the --api-key usage"
	PrivateAPIKeyUsage = "this is the --private-api-key usage"

	ProfileUsage = "this is the --profile, -p usage"

	RealmBaseURLUsage = "this is the --base-url usage"
	CloudBaseURLUsage = "this is the --cloud-base-url usage"

	AutoConfirmUsage   = "this is the --yes, -y usage"
	DisableColorsUsage = "this is the --disable-colors usage"
	OutputFormatUsage  = "formats the output specified by type"
	OutputTargetUsage  = "this is the --output-target usage"
)
