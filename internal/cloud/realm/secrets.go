package realm

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/10gen/realm-cli/internal/utils/api"
)

const (
	secretsPathPattern = appPathPattern + "/secrets"
	secretPathPattern  = secretsPathPattern + "/%s"
)

// Secret is a secret stored in a Realm app
type Secret struct {
	ID   string `json:"_id"`
	Name string `json:"name"`
}

func (c *client) Secrets(groupID, appID string) ([]Secret, error) {
	res, resErr := c.do(
		http.MethodGet,
		fmt.Sprintf(secretsPathPattern, groupID, appID),
		api.RequestOptions{},
	)
	if resErr != nil {
		return nil, resErr
	}

	if res.StatusCode != http.StatusOK {
		return nil, api.ErrUnexpectedStatusCode{"secrets", res.StatusCode}
	}

	defer res.Body.Close()

	var secrets []Secret
	if err := json.NewDecoder(res.Body).Decode(&secrets); err != nil {
		return nil, err
	}
	return secrets, nil
}

type secretsPayload struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (c *client) CreateSecret(groupID, appID, name, value string) (Secret, error) {
	res, resErr := c.doJSON(
		http.MethodPost,
		fmt.Sprintf(secretsPathPattern, groupID, appID),
		secretsPayload{name, value},
		api.RequestOptions{},
	)
	if resErr != nil {
		return Secret{}, resErr
	}
	if res.StatusCode != http.StatusCreated {
		return Secret{}, api.ErrUnexpectedStatusCode{"create secret", res.StatusCode}
	}
	defer res.Body.Close()
	var secret Secret
	if err := json.NewDecoder(res.Body).Decode(&secret); err != nil {
		return Secret{}, err
	}
	return secret, nil
}

func (c *client) DeleteSecret(groupID, appID, secretID string) error {
	res, err := c.do(
		http.MethodDelete,
		fmt.Sprintf(secretPathPattern, groupID, appID, secretID),
		api.RequestOptions{},
	)

	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusNoContent {
		return api.ErrUnexpectedStatusCode{"delete secret", res.StatusCode}
	}

	return nil
}

func (c *client) UpdateSecret(groupID, appID, secretID, name, value string) error {
	res, err := c.doJSON(
		http.MethodPut,
		fmt.Sprintf(secretPathPattern, groupID, appID, secretID),
		secretsPayload{name, value},
		api.RequestOptions{},
	)

	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusNoContent {
		return api.ErrUnexpectedStatusCode{"update secret", res.StatusCode}
	}

	return nil
}
