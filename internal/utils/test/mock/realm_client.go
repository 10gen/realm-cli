package mock

import (
	"github.com/10gen/realm-cli/internal/realm"
)

// RealmClient is a mocked Realm client
type RealmClient struct {
	realm.Client
	LoginFn func(publicAPIKey, privateAPIKey string) (realm.AuthResponse, error)
}

// Login calls the mocked Login implementation if provided, otherwise the call
// falls back to the underlying realm.Client
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) Login(publicAPIKey, privateAPIKey string) (realm.AuthResponse, error) {
	if rc.LoginFn != nil {
		return rc.LoginFn(publicAPIKey, privateAPIKey)
	}
	return rc.Client.Login(publicAPIKey, privateAPIKey)
}
