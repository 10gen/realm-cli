package atlas

import (
	"errors"
	"net/http"

	"github.com/10gen/realm-cli/internal/utils/api"
)

// set of known MongoDB Cloud Atlas status errors
var (
	ErrServerUnavailable = errors.New("Atlas server is not available")
)

func (c *client) Status() error {
	res, err := c.doWithBaseURL(http.MethodGet, publicAPI, api.RequestOptions{})
	if err != nil {
		return ErrServerUnavailable
	}
	if res.StatusCode != http.StatusOK {
		return ErrServerUnavailable
	}
	return nil
}
