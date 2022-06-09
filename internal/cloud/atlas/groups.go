package atlas

import (
	"encoding/json"
	"net/http"

	"github.com/10gen/realm-cli/internal/utils/api"
)

const (
	groupsPath = publicAPI + "/groups"

	linkRelNext = "next"
)

// Group is an Atlas group
type Group struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Link contains additional data related to the response
type Link struct {
	Href string `json:"href"`
	Rel  string `json:"rel"`
}

// Groups represents the groups response from Atlas
type Groups struct {
	Results []Group `json:"results"`
	Links   []Link  `json:"links"`
}

func (c *client) Groups(url string, useBaseURL bool) (Groups, error) {
	if useBaseURL {
		url = c.baseURL + url
	}
	res, err := c.doWithURL(
		http.MethodGet,
		url,
		api.RequestOptions{},
	)
	if err != nil {
		return Groups{}, err
	}
	if res.StatusCode != http.StatusOK {
		return Groups{}, api.ErrUnexpectedStatusCode{"get groups", res.StatusCode}
	}
	defer res.Body.Close()

	var groups Groups
	if err := json.NewDecoder(res.Body).Decode(&groups); err != nil {
		return Groups{}, err
	}
	return groups, nil
}

// AllGroups fetches all atlas groups
func AllGroups(c Client) ([]Group, error) {
	groups, err := c.Groups(groupsPath, true)
	if err != nil {
		return nil, err
	}
	return fetchNextGroups(c, groups)
}

func fetchNextGroups(c Client, groups Groups) ([]Group, error) {
	allGroups := groups.Results
	for _, link := range groups.Links {
		if link.Rel != linkRelNext {
			continue
		}
		res, err := c.Groups(link.Href, false)
		if err != nil {
			return nil, err
		}
		nextGroups, err := fetchNextGroups(c, res)
		if err != nil {
			return nil, err
		}
		allGroups = append(allGroups, nextGroups...)
		break
	}
	return allGroups, nil
}
