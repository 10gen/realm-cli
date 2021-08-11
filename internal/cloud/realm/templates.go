package realm

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/10gen/realm-cli/internal/utils/api"
)

// AllowedTemplates is an array of supported templates.
var AllowedTemplates = []string{
	"web.graphql.todo",
	"web.mql.todo",
	"triggers",
	"ios.swift.todo",
	"android.kotlin.todo",
	"react-native.todo",
	"xamarin.todo",
}

// Template represents an available Realm app template
type Template struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Templates is a slice of Template structs
type Templates []Template

// MapByID converts an array of templates into a map whose keys are the template ids and values are the template
func (templates Templates) MapByID() map[string]Template {
	out := make(map[string]Template, len(templates))
	for _, t := range templates {
		out[t.ID] = t
	}
	return out
}

const (
	allTemplatesPath               = adminAPI + "/templates"
	clientTemplatePathPattern      = appPathPattern + "/templates/%s/client"
	compatibleTemplatesPathPattern = appPathPattern + "/templates"
)

func (c *client) AllTemplates() (Templates, error) {
	res, resErr := c.do(
		http.MethodGet,
		allTemplatesPath,
		api.RequestOptions{},
	)
	if resErr != nil {
		return nil, resErr
	}
	if res.StatusCode != http.StatusOK {
		return nil, api.ErrUnexpectedStatusCode{"get templates", res.StatusCode}
	}
	defer res.Body.Close()

	var templates Templates
	if err := json.NewDecoder(res.Body).Decode(&templates); err != nil {
		return nil, err
	}
	return templates, nil
}

func (c *client) ClientTemplate(groupID, appID, templateID string) (*zip.Reader, bool, error) {
	res, resErr := c.do(http.MethodGet, fmt.Sprintf(clientTemplatePathPattern, groupID, appID, templateID), api.RequestOptions{})
	if resErr != nil {
		return nil, false, resErr
	}
	if res.StatusCode == http.StatusNoContent {
		// No client exists for this template so there is nothing to return
		return nil, false, nil
	}
	if res.StatusCode != http.StatusOK {
		return nil, false, api.ErrUnexpectedStatusCode{"get client template", res.StatusCode}
	}

	defer res.Body.Close()
	body, bodyErr := ioutil.ReadAll(res.Body)
	if bodyErr != nil {
		return nil, false, bodyErr
	}

	zipPkg, zipErr := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if zipErr != nil {
		return nil, false, zipErr
	}

	return zipPkg, true, nil
}

func (c *client) CompatibleTemplates(groupID, appID string) (Templates, error) {
	res, resErr := c.do(http.MethodGet, fmt.Sprintf(compatibleTemplatesPathPattern, groupID, appID), api.RequestOptions{})
	if resErr != nil {
		return nil, resErr
	}
	if res.StatusCode == http.StatusBadRequest {
		// 400 implies app was not created with template so we return no compatible templates
		return nil, nil
	}
	if res.StatusCode != http.StatusOK {
		return nil, api.ErrUnexpectedStatusCode{"get compatible templates", res.StatusCode}
	}
	defer res.Body.Close()

	var templates Templates
	if err := json.NewDecoder(res.Body).Decode(&templates); err != nil {
		return nil, err
	}
	return templates, nil
}
