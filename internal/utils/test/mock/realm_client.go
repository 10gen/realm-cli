package mock

import (
	"github.com/10gen/realm-cli/internal/cloud/realm"
)

// RealmClient is a mocked Realm client
type RealmClient struct {
	realm.Client
	AuthenticateFn                func(publicAPIKey, privateAPIKey string) (realm.AuthResponse, error)
	GetUserProfileFn              func() (realm.UserProfile, error)
	FindProjectAppByClientAppIDFn func(groupIDs []string, app string) ([]realm.App, error)
}

// Authenticate calls the mocked Authenticate implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) Authenticate(publicAPIKey, privateAPIKey string) (realm.AuthResponse, error) {
	if rc.AuthenticateFn != nil {
		return rc.AuthenticateFn(publicAPIKey, privateAPIKey)
	}
	return rc.Client.Authenticate(publicAPIKey, privateAPIKey)
}

// GetUserProfile calls the mocked GetUserProfile implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) GetUserProfile() (realm.UserProfile, error) {
	if rc.GetUserProfileFn != nil {
		return rc.GetUserProfileFn()
	}
	return rc.Client.GetUserProfile()
}

// FindProjectAppByClientAppID calls the mocked FindProjectAppByClientAppID implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) FindProjectAppByClientAppID(groupIDs []string, app string) ([]realm.App, error) {
	if rc.FindProjectAppByClientAppIDFn != nil {
		return rc.FindProjectAppByClientAppIDFn(groupIDs, app)
	}
	return rc.Client.FindProjectAppByClientAppID(groupIDs, app)
}
