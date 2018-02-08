package storage

// User stores the user's login credentials and some metadata.
type User struct {
	APIKey       string `yaml:"api_key"`
	Username     string `yaml:"username"`
	RefreshToken string `yaml:"refresh_token"`
	AccessToken  string `yaml:"access_token"`
}
