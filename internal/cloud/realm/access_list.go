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

// AccessList is a list of allowed IPs stored in a Realm app
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

type allowedIPCreatePayload struct {
	Address    string `json:"address"`
	Comment    string `json:"comment,omitempty"`
	UseCurrent bool   `json:"use_current,omitempty"`
}

func (c *client) AllowedIPs(groupID, appID string) ([]AllowedIP, error) {
	res, resErr := c.do(
		http.MethodGet,
		fmt.Sprintf(allowedIPsPathPattern, groupID, appID),
		api.RequestOptions{},
	)
	if resErr != nil {
		return nil, resErr
	}

	if res.StatusCode != http.StatusOK {
		return nil, api.ErrUnexpectedStatusCode{"get allowed ips and/or CIDR blocks", res.StatusCode}
	}

	defer res.Body.Close()

	var accessList AccessList
	if err := json.NewDecoder(res.Body).Decode(&accessList); err != nil {
		return nil, err
	}
	return accessList.AllowedIPs, nil
}

func (c *client) AllowedIPCreate(groupID, appID, ipAddress, comment string, useCurrent bool) (AllowedIP, error) {
	res, resErr := c.doJSON(
		http.MethodPost,
		fmt.Sprintf(allowedIPsPathPattern, groupID, appID),
		allowedIPCreatePayload{
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
		return AllowedIP{}, api.ErrUnexpectedStatusCode{"create allowed ip or CIDR block", res.StatusCode}
	}

	defer res.Body.Close()

	var allowedIP AllowedIP
	if err := json.NewDecoder(res.Body).Decode(&allowedIP); err != nil {
		return AllowedIP{}, err
	}
	return allowedIP, nil
}
