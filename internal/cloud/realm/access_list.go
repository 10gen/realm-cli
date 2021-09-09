package realm

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/10gen/realm-cli/internal/utils/api"
)

const (
	allowedIPsPathPattern = appPathPattern + "/security/access_list"
	allowedIPPathPattern  = allowedIPsPathPattern + "/%s"
)

// AccessList is a list of allowed IPs stored in a Realm app
type AccessList struct {
	AllowedIPs []AllowedIP `json:"allowed_ips"`
}

// AllowedIP is an IP Access address stored in a Realm app
type AllowedIP struct {
	ID              string `json:"_id"`
	Address         string `json:"address"`
	Comment         string `json:"comment,omitempty"`
	IncludesCurrent bool   `json:"includes_current"`
}

type allowedIPRequest struct {
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
		return nil, api.ErrUnexpectedStatusCode{"get allowed ips", res.StatusCode}
	}

	defer res.Body.Close()

	var accessList AccessList
	if err := json.NewDecoder(res.Body).Decode(&accessList); err != nil {
		return nil, err
	}
	return accessList.AllowedIPs, nil
}

func (c *client) AllowedIPCreate(groupID, appID, address, comment string, useCurrent bool) (AllowedIP, error) {
	res, resErr := c.doJSON(
		http.MethodPost,
		fmt.Sprintf(allowedIPsPathPattern, groupID, appID),
		allowedIPRequest{
			address,
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

func (c *client) AllowedIPUpdate(groupID, appID, allowedIPID, newAddress, newComment string) error {
	res, err := c.doJSON(
		http.MethodPatch,
		fmt.Sprintf(allowedIPPathPattern, groupID, appID, allowedIPID),
		allowedIPRequest{
			Address: newAddress,
			Comment: newComment,
		},
		api.RequestOptions{},
	)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return api.ErrUnexpectedStatusCode{"update allowed ip", res.StatusCode}
	}

	return nil
}

func (c *client) AllowedIPDelete(groupID, appID, allowedIPID string) error {
	res, err := c.do(
		http.MethodDelete,
		fmt.Sprintf(allowedIPPathPattern, groupID, appID, allowedIPID),
		api.RequestOptions{},
	)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusNoContent {
		return api.ErrUnexpectedStatusCode{"delete allowed ip", res.StatusCode}
	}

	return nil
}
