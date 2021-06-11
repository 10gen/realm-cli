package realm

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/10gen/realm-cli/internal/utils/api"
)

const (
	allowedIPsPathPattern = appPathPattern + "/security/allowed_ips"
	allowedIPPathPattern  = allowedIPsPathPattern + "/%s"
)

type AccessList struct {
	AllowedIPs []AllowedIP `json:"allowed_ips"`
	CurrentIP  string      `json:"current_ip"`
}

// AllowedIP is an IP Access address stored in a Realm app
type AllowedIP struct {
	ID                string `json:"_id"`
	IPAddress         string `json:"ip_address"`
	Comment           string `json:"comment"`
	IncludesCurrentIP bool   `json:"includes_current_ip"`
}

func (c *client) AllowedIPs(groupID, appID string) (AccessList, error) {
	res, resErr := c.do(
		http.MethodGet,
		fmt.Sprintf(allowedIPsPathPattern, groupID, appID),
		api.RequestOptions{},
	)
	if resErr != nil {
		return AccessList{}, resErr
	}

	if res.StatusCode != http.StatusOK {
		return AccessList{}, api.ErrUnexpectedStatusCode{"get allowed ips", res.StatusCode}
	}

	defer res.Body.Close()

	var accessList AccessList
	if err := json.NewDecoder(res.Body).Decode(&accessList); err != nil {
		return AccessList{}, err
	}
	return accessList, nil
}

type allowedIPsPayload struct {
	IPAddress  string `json:"ip_address"`
	Comment    string `json:"comment"`
	UseCurrent bool   `json:"use_current"`
	AllowAll   bool   `json:"allow_all"`
}

func (c *client) AllowedIPCreate(groupID, appID, ipAddress, comment string, useCurrent, allowAll bool) (AllowedIP, error) {
	res, resErr := c.doJSON(
		http.MethodPost,
		fmt.Sprintf(allowedIPsPathPattern, groupID, appID),
		allowedIPsPayload{
			ipAddress,
			comment,
			useCurrent,
			allowAll,
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

func (c *client) AllowedIPUpdate(groupID, appID, allowedIPID, newIPAddress, comment string) error {
	res, err := c.doJSON(
		http.MethodPut,
		fmt.Sprintf(allowedIPPathPattern, groupID, appID, allowedIPID),
		allowedIPsPayload{
			newIPAddress,
			comment,
			false,
			false,
		},
		api.RequestOptions{},
	)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusNoContent {
		return api.ErrUnexpectedStatusCode{"update allowed ip", res.StatusCode}
	}

	return nil
}
