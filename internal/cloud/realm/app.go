package realm

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/10gen/realm-cli/internal/utils/api"
	"github.com/AlecAivazis/survey/v2/core"
)

const (
	appsPathPattern = adminAPI + "/groups/%s/apps"
	appPathPattern  = appsPathPattern + "/%s"
)

// AppMeta is Realm application metadata
type AppMeta struct {
	Location        string `json:"location,omitempty"`
	DeploymentModel string `json:"deployment_model,omitempty"`
}

// App is a Realm application
type App struct {
	AppMeta
	ID           string `json:"_id"`
	ClientAppID  string `json:"client_app_id"`
	Name         string `json:"name"`
	DomainID     string `json:"domain_id"`
	GroupID      string `json:"group_id"`
	LastUsed     int64  `json:"last_used"`
	LastModified int64  `json:"last_modified"`
	Product      string `json:"product"`
}

func (app App) String() string {
	return fmt.Sprintf("%s (%s)", app.ClientAppID, app.GroupID)
}

type createAppRequest struct {
	Name string `json:"name"`
	AppMeta
}

func (c *client) CreateApp(groupID, name string, meta AppMeta) (App, error) {
	res, resErr := c.doJSON(
		http.MethodPost,
		fmt.Sprintf(appsPathPattern, groupID),
		createAppRequest{name, meta},
		api.RequestOptions{UseAuth: true},
	)
	if resErr != nil {
		return App{}, resErr
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		return App{}, unmarshalServerError(res)
	}

	var app App
	if err := json.NewDecoder(res.Body).Decode(&app); err != nil {
		return App{}, err
	}
	return app, nil
}

func (c *client) DeleteApp(groupID, appID string) error {
	res, resErr := c.do(
		http.MethodDelete,
		fmt.Sprintf(appPathPattern, groupID, appID),
		api.RequestOptions{UseAuth: true},
	)
	if resErr != nil {
		return resErr
	}
	if res.StatusCode != http.StatusNoContent {
		defer res.Body.Close()
		return unmarshalServerError(res)
	}
	return nil
}

// AppFilter represents the optional filter parameters available for lists of apps
type AppFilter struct {
	GroupID string
	App     string // can be client app id or name
}

func (c *client) FindApps(filter AppFilter) ([]App, error) {
	var apps []App
	if filter.GroupID == "" {
		arr, err := c.getAppsForUser()
		if err != nil {
			return nil, err
		}
		apps = arr
	} else {
		arr, err := c.getApps(filter.GroupID)
		if err != nil {
			return nil, err
		}
		apps = arr
	}

	if filter.App == "" {
		return apps, nil
	}

	var filtered = make([]App, 0, len(apps))
	for _, app := range apps {
		if strings.HasPrefix(app.ClientAppID, filter.App) {
			filtered = append(filtered, app)
		}
	}
	return filtered, nil
}

func (c *client) getAppsForUser() ([]App, error) {
	profile, profileErr := c.AuthProfile()
	if profileErr != nil {
		return nil, profileErr
	}

	var apps []App
	for _, groupID := range profile.AllGroupIDs() {
		projectApps, err := c.getApps(groupID)
		if err != nil {
			return nil, err
		}
		apps = append(apps, projectApps...)
	}
	return apps, nil
}

func (c *client) getApps(groupID string) ([]App, error) {
	res, resErr := c.do(
		http.MethodGet,
		fmt.Sprintf(appsPathPattern, groupID),
		api.RequestOptions{UseAuth: true},
	)
	if resErr != nil {
		return nil, resErr
	}
	if res.StatusCode == http.StatusNotFound {
		return nil, errors.New("group could not be found")
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, unmarshalServerError(res)
	}

	var apps []App
	if err := json.NewDecoder(res.Body).Decode(&apps); err != nil {
		return nil, err
	}
	return apps, nil
}

// DeploymentModel is the Realm app deployment model
type DeploymentModel string

// String returns the deployment model display
func (dm DeploymentModel) String() string { return string(dm) }

// Type returns the DeploymentModel type
func (dm DeploymentModel) Type() string { return "string" }

// Set validates and sets the deployment model value
func (dm *DeploymentModel) Set(val string) error {
	newDeploymentModel := DeploymentModel(val)

	if !isValidDeploymentModel(newDeploymentModel) {
		return errInvalidDeploymentModel
	}

	*dm = newDeploymentModel
	return nil
}

// WriteAnswer validates and sets the deployment model value
func (dm *DeploymentModel) WriteAnswer(name string, value interface{}) error {
	var newDeploymentModel DeploymentModel

	switch v := value.(type) {
	case core.OptionAnswer:
		newDeploymentModel = DeploymentModel(v.Value)
	}

	if !isValidDeploymentModel(newDeploymentModel) {
		return errInvalidDeploymentModel
	}
	*dm = newDeploymentModel
	return nil
}

// set of supported Realm app deployment models
const (
	DeploymentModelNil    DeploymentModel = ""
	DeploymentModelGlobal DeploymentModel = "GLOBAL"
	DeploymentModelLocal  DeploymentModel = "LOCAL"
)

var (
	errInvalidDeploymentModel = func() error {
		allDeploymentModels := []string{DeploymentModelGlobal.String(), DeploymentModelLocal.String()}
		return fmt.Errorf("unsupported value, use one of [%s] instead", strings.Join(allDeploymentModels, ", "))
	}()
)

func isValidDeploymentModel(dm DeploymentModel) bool {
	switch dm {
	case
		DeploymentModelNil, // allow DeploymentModel to be optional
		DeploymentModelGlobal,
		DeploymentModelLocal:
		return true
	}
	return false
}

// Location is the Realm app location
type Location string

// String returns the Location display
func (l Location) String() string { return string(l) }

// Type returns the Location type
func (l Location) Type() string { return "string" }

// Set validates and sets the Location value
func (l *Location) Set(val string) error {
	newLocation := Location(val)

	if !isValidLocation(newLocation) {
		return errInvalidLocation
	}

	*l = newLocation
	return nil
}

// WriteAnswer validates and sets the Location value
func (l *Location) WriteAnswer(name string, value interface{}) error {
	var newLocation Location

	switch v := value.(type) {
	case core.OptionAnswer:
		newLocation = Location(v.Value)
	}

	if !isValidLocation(newLocation) {
		return errInvalidLocation
	}
	*l = newLocation
	return nil
}

// set of supported Realm app locations
const (
	LocationNil       Location = ""
	LocationVirginia  Location = "US-VA"
	LocationOregon    Location = "US-OR"
	LocationFrankfurt Location = "DE-FF"
	LocationIreland   Location = "IE"
	LocationSydney    Location = "AU"
	LocationMumbai    Location = "IN-MB"
	LocationSingapore Location = "SG"
)

var (
	errInvalidLocation = func() error {
		allLocations := []string{
			LocationVirginia.String(),
			LocationOregon.String(),
			LocationFrankfurt.String(),
			LocationIreland.String(),
			LocationSydney.String(),
			LocationMumbai.String(),
			LocationSingapore.String(),
		}
		return fmt.Errorf("unsupported value, use one of [%s] instead", strings.Join(allLocations, ", "))
	}()
)

func isValidLocation(l Location) bool {
	switch l {
	case
		LocationNil, // allow Location to be optional
		LocationVirginia,
		LocationOregon,
		LocationFrankfurt,
		LocationIreland,
		LocationSydney,
		LocationMumbai,
		LocationSingapore:
		return true
	}
	return false
}
