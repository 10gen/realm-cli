package realm

import (
	"fmt"
	"net/http"

	"github.com/10gen/realm-cli/internal/utils/api"
)

// ErrServerNotRunning is an error representing the Realm server is not running at the specified url
func ErrServerNotRunning(baseURL string) error {
	return fmt.Errorf("Realm server is not running at %s", baseURL)
}

func (c *client) Status() error {
	res, err := c.do(http.MethodGet, privateAPI+"/version", api.RequestOptions{})
	if err != nil {
		return ErrServerNotRunning(c.baseURL)
	}
	if res.StatusCode != http.StatusOK {
		return ErrServerNotRunning(c.baseURL)
	}
	return nil
}
