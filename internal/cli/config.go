package cli

// Config is the CLI config
type Config struct {
	PublicAPIKey  string `yaml:"public_api_key,omitempty"`
	PrivateAPIKey string `yaml:"private_api_key,omitempty"`
	RefreshToken  string `yaml:"refresh_token,omitempty"`
	AccessToken   string `yaml:"access_token,omitempty"`
}
