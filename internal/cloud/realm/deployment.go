package realm

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/10gen/realm-cli/internal/utils/api"
)

const (
	deploymentPathPattern = appPathPattern + "/deployments/%s"
)

// AppDeployment is a Realm app deployment
type AppDeployment struct {
	ID     string           `json:"_id"`
	Status DeploymentStatus `json:"status"`
}

// DeploymentStatus is the Realm application deployment status
type DeploymentStatus string

// set of known deployment statuses
const (
	DeploymentStatusCreated    DeploymentStatus = "created"
	DeploymentStatusSuccessful DeploymentStatus = "successful"
	DeploymentStatusFailed     DeploymentStatus = "failed"
	DeploymentStatusPending    DeploymentStatus = "pending"
)

func (c *client) Deployment(groupID, appID, deploymentID string) (AppDeployment, error) {
	res, resErr := c.do(
		http.MethodGet,
		fmt.Sprintf(deploymentPathPattern, groupID, appID, deploymentID),
		api.RequestOptions{},
	)
	if resErr != nil {
		return AppDeployment{}, resErr
	}
	if res.StatusCode != http.StatusOK {
		return AppDeployment{}, api.ErrUnexpectedStatusCode{"get deployment", res.StatusCode}
	}
	defer res.Body.Close()

	var deployment AppDeployment
	if err := json.NewDecoder(res.Body).Decode(&deployment); err != nil {
		return AppDeployment{}, err
	}
	return deployment, nil
}
