package cli

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/10gen/realm-cli/internal/utils/api"
	"github.com/blang/semver"
)

var (
	// Name represents the CLI name; used for invoking the CLI commands
	Name = "realm-cli"

	// Version represents the CLI version
	Version = "0.0.0" // value will be injected at build-time

	// osArch represents the CLI os architecture; used for locating the correct CLI URL
	osArch string // value will be injected at build-time
)

const (
	manifestURL = "https://s3.amazonaws.com/realm-clis/versions/cloud-prod/CURRENT"
)

type versionManifest struct {
	Version string                 `json:"version"`
	Info    map[string]versionInfo `json:"info"`
}

type versionInfo struct {
	URL string `json:"url"`
}

// VersionManifestClient is a version manifest client
type VersionManifestClient interface {
	Get(url string) (*http.Response, error)
}

// checkVersion looks for and returns a URL for a new CLI version, if one exists
func checkVersion(client VersionManifestClient) (string, error) {
	res, err := client.Get(manifestURL)
	if err != nil {
		return "", err
	}
	if res.StatusCode != http.StatusOK {
		return "", api.ErrUnexpectedStatusCode{"get cli version manifest", res.StatusCode}
	}
	defer res.Body.Close()

	var manifest versionManifest
	if err := json.NewDecoder(res.Body).Decode(&manifest); err != nil {
		return "", err
	}

	versionNext, err := parseSemver(manifest.Version)
	if err != nil {
		return "", err
	}

	versionCurrent, err := parseSemver(Version)
	if err != nil {
		return "", err
	}

	if versionCurrent.GTE(versionNext) {
		return "", nil // version is up-to-date
	}

	osInfo, ok := manifest.Info[osArch]
	if !ok {
		return "", fmt.Errorf("unrecognized CLI OS build: %s", osArch)
	}

	return fmt.Sprintf(`New version (v%s) of CLI available: %s`, versionNext, osInfo.URL), nil
}

func parseSemver(version string) (semver.Version, error) {
	parsed, err := semver.Make(version)
	if err != nil {
		return semver.Version{}, fmt.Errorf("failed to parse version v%s", version)
	}
	return parsed, nil
}
