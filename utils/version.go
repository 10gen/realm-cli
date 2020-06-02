package utils

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/blang/semver"
)

var (
	// CLIVersion represents the current version of the CLI. This version is dynamically replaced at build-time
	CLIVersion = "1.0.0"

	// CLIOSArch represents the OS Architecture of the CLI. It is used to select the appropriate URL for the CLI
	// binary based on the user's OS
	CLIOSArch string

	versionManifestURLFormat = "https://s3.amazonaws.com/realm-clis/versions/%s/CURRENT"
	cliBuildEnv              = "cloud-prod"
)

type versionManifest struct {
	Version string                 `json:"version"`
	Info    map[string]versionInfo `json:"info"`
}

type versionInfo struct {
	URL string `json:"url"`
}

// HTTPClient represents the minimum HTTP client required to check for version information
type HTTPClient interface {
	Get(url string) (*http.Response, error)
}

// CheckForNewCLIVersion looks for and returns a url for a new version
// of the CLI, if one exists. Any errors are swallowed.
func CheckForNewCLIVersion(client HTTPClient) string {
	url := fmt.Sprintf(versionManifestURLFormat, cliBuildEnv)
	resp, err := client.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		return ""
	}
	var manifest versionManifest
	dec := json.NewDecoder(resp.Body)
	defer resp.Body.Close()
	dec.UseNumber()
	if decodeErr := dec.Decode(&manifest); decodeErr != nil {
		return ""
	}

	manifestVersion, err := semver.Make(manifest.Version)
	if err != nil {
		return ""
	}

	currVersion, err := semver.Make(CLIVersion)
	if err != nil {
		return ""
	}

	if currVersion.GTE(manifestVersion) {
		return ""
	}

	info, ok := manifest.Info[CLIOSArch]
	if !ok {
		return ""
	}

	return fmt.Sprintf("New version (v%s) of CLI available at %s", manifestVersion, info.URL)
}
