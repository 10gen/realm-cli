package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

var (
	// CLIVersion represents the current version of the CLI. This version is dynamically replaced at build-time
	CLIVersion               = "20180301"
	versionManifestURLFormat = "https://s3.amazonaws.com/stitch-clis/versions/%s/CURRENT"
	cliBuildEnv              = "cloud-prod"
	cliOSArch                string
)

type versionManifest struct {
	Version int64                  `json:"version"`
	Info    map[string]versionInfo `json:"info"`
}

type versionInfo struct {
	URL string `json:"url"`
}

// CheckForNewCLIVersion looks for and returns a url for a new version
// of the CLI, if one exists. Any errors are swallowed.
func CheckForNewCLIVersion() string {
	url := fmt.Sprintf(versionManifestURLFormat, cliBuildEnv)
	resp, err := http.DefaultClient.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		return ""
	}
	var manifest versionManifest
	dec := json.NewDecoder(resp.Body)
	defer resp.Body.Close()
	dec.UseNumber()
	if err := dec.Decode(&manifest); err != nil {
		return ""
	}

	currVersion, err := strconv.ParseInt(CLIVersion, 10, 64)
	if err != nil {
		return ""
	}
	if currVersion >= manifest.Version {
		return ""
	}

	info, ok := manifest.Info[cliOSArch]
	if !ok {
		return ""
	}

	return fmt.Sprintf("New version of CLI available at %s", info.URL)
}
