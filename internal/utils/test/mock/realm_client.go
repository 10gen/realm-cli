package mock

import (
	"github.com/10gen/realm-cli/internal/cloud/realm"
)

// RealmClient is a mocked Realm client
type RealmClient struct {
	realm.Client
	AuthenticateFn   func(publicAPIKey, privateAPIKey string) (realm.Session, error)
	GetUserProfileFn func() (realm.UserProfile, error)
	GetAppsForUserFn func() ([]realm.App, error)
	GetAppsFn        func(groupID string) ([]realm.App, error)
}

// Authenticate calls the mocked Authenticate implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) Authenticate(publicAPIKey, privateAPIKey string) (realm.Session, error) {
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

// GetAppsForUser calls the mocked GetAppsForUser implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) GetAppsForUser() ([]realm.App, error) {
	if rc.GetAppsForUserFn != nil {
		return rc.GetAppsForUserFn()
	}
	return rc.Client.GetAppsForUser()
}

// GetApps calls the mocked GetApps implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) GetApps(groupID string) ([]realm.App, error) {
	if rc.GetAppsFn != nil {
		return rc.GetAppsFn(groupID)
	}
	return rc.Client.GetApps(groupID)
}
