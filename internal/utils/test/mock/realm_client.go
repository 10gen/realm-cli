package mock

import (
	"github.com/10gen/realm-cli/internal/cloud/realm"
)

// RealmClient is a mocked Realm client
type RealmClient struct {
	realm.Client
	AuthenticateFn   func(publicAPIKey, privateAPIKey string) (realm.Session, error)
	GetAuthProfileFn func() (realm.AuthProfile, error)
	GetAppsForUserFn func() ([]realm.App, error)
	GetAppsFn        func(groupID string) ([]realm.App, error)
	StatusFn         func() error
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

// GetAuthProfile calls the mocked GetAuthProfile implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) GetAuthProfile() (realm.AuthProfile, error) {
	if rc.GetAuthProfileFn != nil {
		return rc.GetAuthProfileFn()
	}
	return rc.Client.GetAuthProfile()
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

// Status calls the mocked Status implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) Status() error {
	if rc.StatusFn != nil {
		return rc.Status()
	}
	return rc.Client.Status()
}
