// +build !mock

// Package atlas is for interacting with the Atlas public API.
package atlas

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/10gen/stitch-cli/config"
)

const baseURL = "https://cloud.mongodb.com/api/atlas/v1.0/"

// GetAllClusters returns all Atlas groups with their list of cluster names.
func GetAllClusters() (clusters map[string][]string, err error) {
	// TODO: get groups using stitch admin SDK
	allGroups := []string{"group-1", "group-2", "group3"}

	groupsMutex := new(sync.Mutex)
	wg := new(sync.WaitGroup)
	for _, group := range allGroups {
		wg.Add(1)
		go func(group string) {
			defer wg.Done()
			results, err := GetClusters(group)
			if err != nil {
				return
			}
			clusterNames := make([]string, len(results))
			for i, cluster := range results {
				clusterNames[i] = cluster[0]
			}
			groupsMutex.Lock()
			clusters[group] = clusterNames
			groupsMutex.Unlock()
		}(group)
	}
	wg.Wait()
	return
}

// GetClusters returns the Atlas clusters for the given group in arrays of (name, URI)
func GetClusters(group string) (clusters [][2]string, err error) {
	url := fmt.Sprintf("groups/%s/clusters", group)
	var output struct {
		Results []struct {
			Name string `json:"name"`
			URI  string `json:"mongoURI"`
		} `json:"results"`
	}
	err = doAtlasReq(url, &output)
	if err != nil {
		return
	}
	for _, res := range output.Results {
		clusters = append(clusters, [2]string{res.Name, res.URI})
	}
	return
}

// GetClusterURI returns the URI corresponding to the specified Atlas cluster.
func GetClusterURI(group, cluster string) (uri string, err error) {
	url := fmt.Sprintf("groups/%s/clusters/%s", group, cluster)
	var output struct {
		URI string `json:"mongoURI"`
	}
	err = doAtlasReq(url, &output)
	uri = output.URI
	return
}

func doAtlasReq(url string, jsonOut interface{}) error {
	req, err := http.NewRequest(http.MethodGet, baseURL+url, http.NoBody)
	if err != nil {
		return err
	}
	user := config.User()
	req.SetBasicAuth(user.Username, user.APIKey)
	c := new(http.Client)
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		// TODO: more helpful errors for 403, etc
		// e.g. something like ErrNotLoggedIn, ErrNotInGroup, ErrClusterNotFound
		return errors.New(resp.Status)
	}
	dec := json.NewDecoder(resp.Body)
	return dec.Decode(jsonOut)
}
