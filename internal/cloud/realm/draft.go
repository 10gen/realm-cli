package realm

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/10gen/realm-cli/internal/utils/api"
)

const (
	draftsPathPattern      = appPathPattern + "/drafts"
	draftPathPattern       = draftsPathPattern + "/%s"
	draftDeployPathPattern = draftPathPattern + "/deployment"
	draftDiffPathPattern   = draftPathPattern + "/diff"
)

// AppDraft is a Realm app draft
type AppDraft struct {
	ID string `json:"_id"`
}

func (c *client) CreateDraft(groupID, appID string) (AppDraft, error) {
	res, resErr := c.do(
		http.MethodPost,
		fmt.Sprintf(draftsPathPattern, groupID, appID),
		api.RequestOptions{},
	)
	if resErr != nil {
		return AppDraft{}, resErr
	}
	if res.StatusCode != http.StatusCreated {
		return AppDraft{}, api.ErrUnexpectedStatusCode{"create draft", res.StatusCode}
	}
	defer res.Body.Close()

	var draft AppDraft
	if err := json.NewDecoder(res.Body).Decode(&draft); err != nil {
		return AppDraft{}, err
	}
	return draft, nil
}

func (c *client) DeployDraft(groupID, appID, draftID string) (AppDeployment, error) {
	res, resErr := c.do(
		http.MethodPost,
		fmt.Sprintf(draftDeployPathPattern, groupID, appID, draftID),
		api.RequestOptions{},
	)
	if resErr != nil {
		return AppDeployment{}, resErr
	}
	if res.StatusCode != http.StatusCreated {
		return AppDeployment{}, api.ErrUnexpectedStatusCode{"deploy draft", res.StatusCode}
	}
	defer res.Body.Close()

	var deployment AppDeployment
	if err := json.NewDecoder(res.Body).Decode(&deployment); err != nil {
		return AppDeployment{}, err
	}
	return deployment, nil
}

func (c *client) DiffDraft(groupID, appID, draftID string) (AppDraftDiff, error) {
	res, resErr := c.do(
		http.MethodGet,
		fmt.Sprintf(draftDiffPathPattern, groupID, appID, draftID),
		api.RequestOptions{},
	)
	if resErr != nil {
		return AppDraftDiff{}, resErr
	}
	if res.StatusCode != http.StatusOK {
		return AppDraftDiff{}, api.ErrUnexpectedStatusCode{"diff draft", res.StatusCode}
	}
	defer res.Body.Close()

	var draftDiff AppDraftDiff
	if err := json.NewDecoder(res.Body).Decode(&draftDiff); err != nil {
		return AppDraftDiff{}, err
	}
	return draftDiff, nil
}

func (c *client) DiscardDraft(groupID, appID, draftID string) error {
	res, resErr := c.do(
		http.MethodDelete,
		fmt.Sprintf(draftPathPattern, groupID, appID, draftID),
		api.RequestOptions{},
	)
	if resErr != nil {
		return resErr
	}
	if res.StatusCode != http.StatusNoContent {
		return api.ErrUnexpectedStatusCode{"discard draft", res.StatusCode}
	}
	return nil
}

func (c *client) Draft(groupID, appID string) (AppDraft, error) {
	res, resErr := c.do(
		http.MethodGet,
		fmt.Sprintf(draftsPathPattern, groupID, appID),
		api.RequestOptions{},
	)
	if resErr != nil {
		return AppDraft{}, resErr
	}
	if res.StatusCode != http.StatusOK {
		return AppDraft{}, api.ErrUnexpectedStatusCode{"get drafts", res.StatusCode}
	}
	defer res.Body.Close()

	var drafts []AppDraft
	if err := json.NewDecoder(res.Body).Decode(&drafts); err != nil {
		return AppDraft{}, err
	}

	if len(drafts) == 0 || len(drafts) > 1 {
		return AppDraft{}, ErrDraftNotFound
	}

	return drafts[0], nil
}
