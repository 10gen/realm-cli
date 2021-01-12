package realm

import (
	"errors"
	"net/http"

	"github.com/10gen/realm-cli/internal/utils/api"
)

const (
	statusPath = privateAPI + "/version"
)

// set of supported status errors
var (
	ErrServerUnavailable = errors.New("Realm server is not available")
)

func (c *client) Status() error {
	res, err := c.do(http.MethodGet, statusPath, api.RequestOptions{NoAuth: true})
	if err != nil {
		return ErrServerUnavailable
	}
	if res.StatusCode != http.StatusOK {
		return ErrServerUnavailable
	}
	return nil
}
