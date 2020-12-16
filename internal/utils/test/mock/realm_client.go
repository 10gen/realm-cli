package mock

import (
	"github.com/10gen/realm-cli/internal/cloud/realm"
)

// RealmClient is a mocked Realm client
type RealmClient struct {
	realm.Client
	AuthenticateFn func(publicAPIKey, privateAPIKey string) (realm.Session, error)
	AuthProfileFn  func() (realm.AuthProfile, error)

	FindAppsFn func(filter realm.AppFilter) ([]realm.App, error)

	CreateAPIKeyFn func(groupID, appID, apiKeyName string) (realm.APIKey, error)
	CreateUserFn   func(groupID, appID, email, password string) (realm.User, error)

	StatusFn func() error
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

// AuthProfile calls the mocked AuthProfile implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) AuthProfile() (realm.AuthProfile, error) {
	if rc.AuthProfileFn != nil {
		return rc.AuthProfileFn()
	}
	return rc.Client.AuthProfile()
}

// FindApps calls the mocked FindApps implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) FindApps(filter realm.AppFilter) ([]realm.App, error) {
	if rc.FindAppsFn != nil {
		return rc.FindAppsFn(filter)
	}
	return rc.Client.FindApps(filter)
}

// CreateAPIKey calls the mocked CreateAPIKey implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) CreateAPIKey(groupID, appID, apiKeyName string) (realm.APIKey, error) {
	if rc.CreateAPIKeyFn != nil {
		return rc.CreateAPIKeyFn(groupID, appID, apiKeyName)
	}
	return rc.Client.CreateAPIKey(groupID, appID, apiKeyName)
}

// CreateUser calls the mocked CreateUser implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) CreateUser(groupID, appID, email, password string) (realm.User, error) {
	if rc.CreateUserFn != nil {
		return rc.CreateUserFn(groupID, appID, email, password)
	}
	return rc.Client.CreateUser(groupID, appID, email, password)
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
