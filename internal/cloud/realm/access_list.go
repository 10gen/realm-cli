package realm

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/10gen/realm-cli/internal/utils/api"
)

const (
	allowedIPsPathPattern = appPathPattern + "/security/allowed_ips"
)

// AccessList is the list of addresses stored in a Realm app
type AccessList struct {
	AllowedIPs []AllowedIP `json:"allowed_ips"`
	CurrentIP  string      `json:"current_ip"`
}

// AllowedIP is an IP Access address stored in a Realm app
type AllowedIP struct {
	ID              string `json:"_id"`
	Address         string `json:"address"`
	Comment         string `json:"comment,omitempty"`
	IncludesCurrent bool   `json:"includes_current"`
}

type allowedIPsPayload struct {
	Address    string `json:"address"`
	Comment    string `json:"comment,omitempty"`
	UseCurrent bool   `json:"use_current,omitempty"`
}

func (c *client) AllowedIPCreate(groupID, appID, ipAddress, comment string, useCurrent bool) (AllowedIP, error) {
	res, resErr := c.doJSON(
		http.MethodPost,
		fmt.Sprintf(allowedIPsPathPattern, groupID, appID),
		allowedIPsPayload{
			ipAddress,
			comment,
			useCurrent,
		},
		api.RequestOptions{},
	)
	if resErr != nil {
		return AllowedIP{}, resErr
	}

	if res.StatusCode != http.StatusCreated {
		return AllowedIP{}, api.ErrUnexpectedStatusCode{"create allowed ip", res.StatusCode}
	}

	defer res.Body.Close()

	var allowedIP AllowedIP
	if err := json.NewDecoder(res.Body).Decode(&allowedIP); err != nil {
		return AllowedIP{}, err
	}
	return allowedIP, nil
}
