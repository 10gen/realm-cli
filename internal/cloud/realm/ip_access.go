package realm

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/10gen/realm-cli/internal/utils/api"
)

const (
	allowedIPsPathPattern = appPathPattern + "/allowed_ips"
	allowedIPPathPattern  = allowedIPsPathPattern + "/%s"
)

// AllowedIP is an IP Access address stored in a Realm app
type AllowedIP struct {
	ID        string
	IPAddress string
	Comment   string
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
		return nil, api.ErrUnexpectedStatusCode{"allowed ips", res.StatusCode}
	}

	defer res.Body.Close()

	var allowedIPs []AllowedIP
	if err := json.NewDecoder(res.Body).Decode(&allowedIPs); err != nil {
		return nil, err
	}
	return allowedIPs, nil
}

type allowedIPsPayload struct{}

func (c *client) CreateAllowedIP(groupID, appID, allowedIPAddress, comment string) (AllowedIP, error) {
	res, resErr := c.doJSON(
		http.MethodPost,
		fmt.Sprintf(allowedIPsPathPattern, groupID, appID),
		allowedIPsPayload{},
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

func (c *client) DeleteAllowedIP(groupID, appID, allowedIPID string) error {
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

func (c *client) UpdateAllowedIP(groupID, appID, allowedIPID, allowedIPAddress, comment string) error {
	res, err := c.doJSON(
		http.MethodPut,
		fmt.Sprintf(allowedIPPathPattern, groupID, appID, allowedIPID),
		allowedIPsPayload{},
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
