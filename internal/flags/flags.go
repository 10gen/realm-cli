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

// set of default flag values
const (
	DefaultOutputFormat = OutputFormatText
	DefaultRealmBaseURL = "https://realm.mongodb.com"
)

// set of supported output formats
const (
	OutputFormatJSON = "json"
	OutputFormatText = "text"
)
