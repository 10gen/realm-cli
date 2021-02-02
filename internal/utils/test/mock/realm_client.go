package mock

import (
	"archive/zip"

	"github.com/10gen/realm-cli/internal/cloud/realm"
)

// RealmClient is a mocked Realm client
type RealmClient struct {
	realm.Client
	AuthenticateFn func(publicAPIKey, privateAPIKey string) (realm.Session, error)
	AuthProfileFn  func() (realm.AuthProfile, error)

	DiffFn   func(groupID, appID string, pkg map[string]interface{}) ([]string, error)
	ExportFn func(groupID, appID string, req realm.ExportRequest) (string, *zip.Reader, error)
	ImportFn func(groupID, appID string, pkg map[string]interface{}) error

	CreateAppFn func(groupID, name string, meta realm.AppMeta) (realm.App, error)
	FindAppsFn  func(filter realm.AppFilter) ([]realm.App, error)

	CreateDraftFn  func(groupID, appID string) (realm.AppDraft, error)
	DiffDraftFn    func(groupID, appID, draftID string) (realm.AppDraftDiff, error)
	DiscardDraftFn func(groupID, appID, draftID string) error
	DraftFn        func(groupID, appID string) (realm.AppDraft, error)

	DeployDraftFn func(groupID, appID, draftID string) (realm.AppDeployment, error)
	DeploymentFn  func(groupID, appID, deploymentID string) (realm.AppDeployment, error)

	SecretsFn      func(groupID, appID string) ([]realm.Secret, error)
	CreateSecretFn func(groupID, appID, name, value string) (realm.Secret, error)

	CreateAPIKeyFn func(groupID, appID, apiKeyName string) (realm.APIKey, error)
	CreateUserFn   func(groupID, appID, email, password string) (realm.User, error)
	DeleteUserFn   func(groupID, appID, userID string) error
	FindUsersFn    func(groupID, appID string, filter realm.UserFilter) ([]realm.User, error)

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

// Export calls the mocked Export implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) Export(groupID, appID string, req realm.ExportRequest) (string, *zip.Reader, error) {
	if rc.ExportFn != nil {
		return rc.ExportFn(groupID, appID, req)
	}
	return rc.Client.Export(groupID, appID, req)
}

// Import calls the mocked Import implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) Import(groupID, appID string, pkg map[string]interface{}) error {
	if rc.ImportFn != nil {
		return rc.ImportFn(groupID, appID, pkg)
	}
	return rc.Client.Import(groupID, appID, pkg)
}

// Diff calls the mocked Diff implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) Diff(groupID, appID string, pkg map[string]interface{}) ([]string, error) {
	if rc.DiffFn != nil {
		return rc.DiffFn(groupID, appID, pkg)
	}
	return rc.Client.Diff(groupID, appID, pkg)
}

// CreateApp calls the mocked CreateApp implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) CreateApp(groupID, name string, meta realm.AppMeta) (realm.App, error) {
	if rc.CreateAppFn != nil {
		return rc.CreateAppFn(groupID, name, meta)
	}
	return rc.Client.CreateApp(groupID, name, meta)
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

// CreateDraft calls the mocked CreateDraft implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) CreateDraft(groupID, appID string) (realm.AppDraft, error) {
	if rc.CreateDraftFn != nil {
		return rc.CreateDraftFn(groupID, appID)
	}
	return rc.Client.CreateDraft(groupID, appID)
}

// DeployDraft calls the mocked DeployDraft implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) DeployDraft(groupID, appID, draftID string) (realm.AppDeployment, error) {
	if rc.DeployDraftFn != nil {
		return rc.DeployDraftFn(groupID, appID, draftID)
	}
	return rc.Client.DeployDraft(groupID, appID, draftID)
}

// DiffDraft calls the mocked DiffDraft implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) DiffDraft(groupID, appID, draftID string) (realm.AppDraftDiff, error) {
	if rc.DiffDraftFn != nil {
		return rc.DiffDraftFn(groupID, appID, draftID)
	}
	return rc.Client.DiffDraft(groupID, appID, draftID)
}

// DiscardDraft calls the mocked DiscardDraft implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) DiscardDraft(groupID, appID, draftID string) error {
	if rc.DiscardDraftFn != nil {
		return rc.DiscardDraftFn(groupID, appID, draftID)
	}
	return rc.Client.DiscardDraft(groupID, appID, draftID)
}

// Draft calls the mocked Draft implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) Draft(groupID, appID string) (realm.AppDraft, error) {
	if rc.DraftFn != nil {
		return rc.DraftFn(groupID, appID)
	}
	return rc.Client.Draft(groupID, appID)
}

// Deployment calls the mocked Deployment implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) Deployment(groupID, appID, deploymentID string) (realm.AppDeployment, error) {
	if rc.DeploymentFn != nil {
		return rc.DeploymentFn(groupID, appID, deploymentID)
	}
	return rc.Client.Deployment(groupID, appID, deploymentID)
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

// Secrets calls the mocked Secrets implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) Secrets(groupID, appID string) ([]realm.Secret, error) {
	if rc.SecretsFn != nil {
		return rc.SecretsFn(groupID, appID)
	}
	return rc.Client.Secrets(groupID, appID)
}

// CreateSecret calls the mocked CreateSecret implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) CreateSecret(groupID, appID, name, value string) (realm.Secret, error) {
	if rc.CreateSecretFn != nil {
		return rc.CreateSecretFn(groupID, appID, name, value)
	}
	return rc.Client.CreateSecret(groupID, appID, name, value)
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

// DeleteUser calls the mocked DeleteUser implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) DeleteUser(groupID, appID, userID string) error {
	if rc.DeleteUserFn != nil {
		return rc.DeleteUserFn(groupID, appID, userID)
	}
	return rc.Client.DeleteUser(groupID, appID, userID)
}

// FindUsers calls the mocked FindUsers implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) FindUsers(groupID, appID string, filter realm.UserFilter) ([]realm.User, error) {
	if rc.FindUsersFn != nil {
		return rc.FindUsersFn(groupID, appID, filter)
	}
	return rc.Client.FindUsers(groupID, appID, filter)
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
