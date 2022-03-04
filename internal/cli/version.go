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

	// OSArch represents the CLI os architecture; used for locating the correct CLI URL
	OSArch string // value will be injected at build-time
)

const (
	manifestURL = "https://s3.amazonaws.com/realm-clis/versions/cloud-prod/CURRENT"
)

type versionManifest struct {
	Version string               `json:"version"`
	Info    map[string]buildInfo `json:"info"`
}

type buildInfo struct {
	Semver string `json:"-"`
	URL    string `json:"url"`
}

// VersionManifestClient is a version manifest client
type VersionManifestClient interface {
	Get(url string) (*http.Response, error)
}

// checkVersion looks for and returns a URL for a new CLI version, if one exists
func checkVersion(client VersionManifestClient) (buildInfo, error) {
	res, err := client.Get(manifestURL)
	if err != nil {
		return buildInfo{}, err
	}
	if res.StatusCode != http.StatusOK {
		return buildInfo{}, api.ErrUnexpectedStatusCode{"get cli version manifest", res.StatusCode}
	}
	defer res.Body.Close()

	var manifest versionManifest
	if err := json.NewDecoder(res.Body).Decode(&manifest); err != nil {
		return buildInfo{}, err
	}

	versionNext, err := parseSemver(manifest.Version)
	if err != nil {
		return buildInfo{}, err
	}

	versionCurrent, err := parseSemver(Version)
	if err != nil {
		return buildInfo{}, err
	}

	if versionCurrent.GTE(versionNext) {
		return buildInfo{}, nil // version is up-to-date
	}

	osInfo, ok := manifest.Info[OSArch]
	if !ok {
		return buildInfo{}, fmt.Errorf("unrecognized CLI OS build: %s", OSArch)
	}

	return buildInfo{versionNext.String(), osInfo.URL}, nil
}

func parseSemver(version string) (semver.Version, error) {
	parsed, err := semver.Make(version)
	if err != nil {
		return semver.Version{}, fmt.Errorf("failed to parse version v%s", version)
	}
	return parsed, nil
}
