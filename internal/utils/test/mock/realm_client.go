package mock

import (
	"archive/zip"
	"io"

	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/cloud/realm"
)

// RealmClient is a mocked Realm client
type RealmClient struct {
	realm.Client

	AuthenticateFn func(authType string, creds user.Credentials) (realm.Session, error)
	AuthProfileFn  func() (realm.AuthProfile, error)

	DiffFn   func(groupID, appID string, appData interface{}) ([]string, error)
	ExportFn func(groupID, appID string, req realm.ExportRequest) (string, *zip.Reader, error)
	ImportFn func(groupID, appID string, appData interface{}) error

	ExportDependenciesFn        func(groupID, appID string) (string, io.ReadCloser, error)
	ExportDependenciesArchiveFn func(groupID, appID string) (string, io.ReadCloser, error)
	ImportDependenciesFn        func(groupID, appID, uploadPath string) error
	DiffDependenciesFn          func(groupID, appID, uploadPath string) (realm.DependenciesDiff, error)
	DependenciesStatusFn        func(groupID, appID string) (realm.DependenciesStatus, error)

	CreateAppFn      func(groupID, name string, meta realm.AppMeta) (realm.App, error)
	DeleteAppFn      func(groupID, appID string) error
	FindAppFn        func(groupID, appID string) (realm.App, error)
	FindAppsFn       func(filter realm.AppFilter) ([]realm.App, error)
	AppDescriptionFn func(groupID, appID string) (realm.AppDescription, error)

	CreateDraftFn  func(groupID, appID string) (realm.AppDraft, error)
	DiffDraftFn    func(groupID, appID, draftID string) (realm.AppDraftDiff, error)
	DiscardDraftFn func(groupID, appID, draftID string) error
	DraftFn        func(groupID, appID string) (realm.AppDraft, error)

	DeployDraftFn func(groupID, appID, draftID string) (realm.AppDeployment, error)
	DeploymentFn  func(groupID, appID, deploymentID string) (realm.AppDeployment, error)

	SecretsFn      func(groupID, appID string) ([]realm.Secret, error)
	CreateSecretFn func(groupID, appID, name, value string) (realm.Secret, error)
	DeleteSecretFn func(groupID, appID, secretID string) error
	UpdateSecretFn func(groupID, appID, secretID, name, value string) error

	CreateAPIKeyFn      func(groupID, appID, apiKeyName string) (realm.APIKey, error)
	CreateUserFn        func(groupID, appID, email, password string) (realm.User, error)
	DeleteUserFn        func(groupID, appID, userID string) error
	DisableUserFn       func(groupID, appID, userID string) error
	EnableUserFn        func(groupID, appID, userID string) error
	FindUsersFn         func(groupID, appID string, filter realm.UserFilter) ([]realm.User, error)
	RevokeUserSessionFn func(groupID, appID, userID string) error

	HostingAssetsFn                func(groupID, appID string) ([]realm.HostingAsset, error)
	HostingAssetUploadFn           func(groupID, appID, rootDir string, asset realm.HostingAsset) error
	HostingAssetRemoveFn           func(groupID, appID, path string) error
	HostingAssetAttributesUpdateFn func(groupID, appID, path string, attrs ...realm.HostingAssetAttribute) error
	HostingCacheInvalidateFn       func(groupID, appID, path string) error

	FunctionsFn               func(groupID, appID string) ([]realm.Function, error)
	AppDebugExecuteFunctionFn func(groupID, appID, userID, name string, args []interface{}) (realm.ExecutionResults, error)

	LogsFn func(groupID, appID string, opts realm.LogsOptions) (realm.Logs, error)

	SchemaModelsFn func(groupID, appID, language string) ([]realm.SchemaModel, error)

	AllTemplatesFn        func() ([]realm.Template, error)
	ClientTemplateFn      func(groupID, appID, templateID string) (*zip.Reader, bool, error)
	CompatibleTemplatesFn func(groupID, appID string) ([]realm.Template, error)

	AllowedIPsFn      func(groupID, appID string) ([]realm.AllowedIP, error)
	AllowedIPCreateFn func(groupID, appID, address, comment string, useCurrent bool) (realm.AllowedIP, error)
	AllowedIPUpdateFn func(groupID, appID, allowedIPID, newAddress, newComment string) error
	AllowedIPDeleteFn func(groupID, appID, allowedIPID string) error

	StatusFn func() error
}

// Authenticate calls the mocked Authenticate implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) Authenticate(authType string, creds user.Credentials) (realm.Session, error) {
	if rc.AuthenticateFn != nil {
		return rc.AuthenticateFn(authType, creds)
	}
	return rc.Client.Authenticate(authType, creds)
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
func (rc RealmClient) Import(groupID, appID string, appData interface{}) error {
	if rc.ImportFn != nil {
		return rc.ImportFn(groupID, appID, appData)
	}
	return rc.Client.Import(groupID, appID, appData)
}

// Diff calls the mocked Diff implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) Diff(groupID, appID string, appData interface{}) ([]string, error) {
	if rc.DiffFn != nil {
		return rc.DiffFn(groupID, appID, appData)
	}
	return rc.Client.Diff(groupID, appID, appData)
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

// DeleteApp calls the mocked DeleteApp implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) DeleteApp(groupID, appID string) error {
	if rc.DeleteAppFn != nil {
		return rc.DeleteAppFn(groupID, appID)
	}
	return rc.Client.DeleteApp(groupID, appID)
}

// FindApp calls the mocked FindApp implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) FindApp(groupID, appID string) (realm.App, error) {
	if rc.FindAppFn != nil {
		return rc.FindAppFn(groupID, appID)
	}
	return rc.Client.FindApp(groupID, appID)
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

// AppDescription calls the mocked AppDescription implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) AppDescription(groupID, appID string) (realm.AppDescription, error) {
	if rc.AppDescriptionFn != nil {
		return rc.AppDescriptionFn(groupID, appID)
	}
	return rc.Client.AppDescription(groupID, appID)
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

// DependenciesStatus calls the mocked DependenciesStatus implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) DependenciesStatus(groupID, appID string) (realm.DependenciesStatus, error) {
	if rc.DependenciesStatusFn != nil {
		return rc.DependenciesStatusFn(groupID, appID)
	}
	return rc.Client.DependenciesStatus(groupID, appID)
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

// DeleteSecret calls the mocked DeleteSecret implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) DeleteSecret(groupID, appID, secretID string) error {
	if rc.DeleteSecretFn != nil {
		return rc.DeleteSecretFn(groupID, appID, secretID)
	}
	return rc.Client.DeleteSecret(groupID, appID, secretID)
}

// UpdateSecret calls the mocked UpdateSecret implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) UpdateSecret(groupID, appID, secretID, name, value string) error {
	if rc.UpdateSecretFn != nil {
		return rc.UpdateSecretFn(groupID, appID, secretID, name, value)
	}
	return rc.Client.UpdateSecret(groupID, appID, secretID, name, value)
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

// DisableUser calls the mocked DisableUser implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) DisableUser(groupID, appID, userID string) error {
	if rc.DisableUserFn != nil {
		return rc.DisableUserFn(groupID, appID, userID)
	}
	return rc.Client.DisableUser(groupID, appID, userID)
}

// EnableUser calls the mocked EnableUser implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) EnableUser(groupID, appID, userID string) error {
	if rc.EnableUserFn != nil {
		return rc.EnableUserFn(groupID, appID, userID)
	}
	return rc.Client.EnableUser(groupID, appID, userID)
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

// RevokeUserSessions calls the mocked RevokeUserSessions implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) RevokeUserSessions(groupID, appID, userID string) error {
	if rc.RevokeUserSessionFn != nil {
		return rc.RevokeUserSessionFn(groupID, appID, userID)
	}
	return rc.Client.RevokeUserSessions(groupID, appID, userID)
}

// ExportDependencies calls the mocked ExportDependencies implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) ExportDependencies(groupID, appID string) (string, io.ReadCloser, error) {
	if rc.ExportDependenciesFn != nil {
		return rc.ExportDependenciesFn(groupID, appID)
	}
	return rc.Client.ExportDependencies(groupID, appID)
}

// ExportDependenciesArchive calls the mocked ExportDependenciesArchive implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) ExportDependenciesArchive(groupID, appID string) (string, io.ReadCloser, error) {
	if rc.ExportDependenciesArchiveFn != nil {
		return rc.ExportDependenciesArchiveFn(groupID, appID)
	}
	return rc.Client.ExportDependenciesArchive(groupID, appID)
}

// ImportDependencies calls the mocked ImportDependencies implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) ImportDependencies(groupID, appID, uploadPath string) error {
	if rc.ImportDependenciesFn != nil {
		return rc.ImportDependenciesFn(groupID, appID, uploadPath)
	}
	return rc.Client.ImportDependencies(groupID, appID, uploadPath)
}

// DiffDependencies calls the mocked DiffDependencies implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) DiffDependencies(groupID, appID, uploadPath string) (realm.DependenciesDiff, error) {
	if rc.DiffDependenciesFn != nil {
		return rc.DiffDependenciesFn(groupID, appID, uploadPath)
	}
	return rc.Client.DiffDependencies(groupID, appID, uploadPath)
}

// HostingAssets calls the mocked HostingAssets implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) HostingAssets(groupID, appID string) ([]realm.HostingAsset, error) {
	if rc.HostingAssetsFn != nil {
		return rc.HostingAssetsFn(groupID, appID)
	}
	return rc.Client.HostingAssets(groupID, appID)
}

// HostingAssetUpload calls the mocked HostingAssetUpload implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) HostingAssetUpload(groupID, appID, rootDir string, asset realm.HostingAsset) error {
	if rc.HostingAssetUploadFn != nil {
		return rc.HostingAssetUploadFn(groupID, appID, rootDir, asset)
	}
	return rc.Client.HostingAssetUpload(groupID, appID, rootDir, asset)
}

// HostingAssetRemove calls the mocked HostingAssetRemove implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) HostingAssetRemove(groupID, appID, path string) error {
	if rc.HostingAssetRemoveFn != nil {
		return rc.HostingAssetRemoveFn(groupID, appID, path)
	}
	return rc.Client.HostingAssetRemove(groupID, appID, path)
}

// HostingAssetAttributesUpdate calls the mocked HostingAssetAttributesUpdate implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) HostingAssetAttributesUpdate(groupID, appID, path string, attrs ...realm.HostingAssetAttribute) error {
	if rc.HostingAssetAttributesUpdateFn != nil {
		return rc.HostingAssetAttributesUpdateFn(groupID, appID, path, attrs...)
	}
	return rc.Client.HostingAssetAttributesUpdate(groupID, appID, path, attrs...)
}

// HostingCacheInvalidate calls the mocked HostingCacheInvalidate implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) HostingCacheInvalidate(groupID, appID, path string) error {
	if rc.HostingCacheInvalidateFn != nil {
		return rc.HostingCacheInvalidateFn(groupID, appID, path)
	}
	return rc.Client.HostingCacheInvalidate(groupID, appID, path)
}

// Functions calls the mocked Functions implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) Functions(groupID, appID string) ([]realm.Function, error) {
	if rc.FunctionsFn != nil {
		return rc.FunctionsFn(groupID, appID)
	}
	return rc.Client.Functions(groupID, appID)
}

// AppDebugExecuteFunction calls the mocked AppDebugExecuteFunction implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) AppDebugExecuteFunction(groupID, appID, userID, name string, args []interface{}) (realm.ExecutionResults, error) {
	if rc.AppDebugExecuteFunctionFn != nil {
		return rc.AppDebugExecuteFunctionFn(groupID, appID, userID, name, args)
	}
	return rc.Client.AppDebugExecuteFunction(groupID, appID, userID, name, args)
}

// Logs calls the mocked Logs implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) Logs(groupID, appID string, opts realm.LogsOptions) (realm.Logs, error) {
	if rc.LogsFn != nil {
		return rc.LogsFn(groupID, appID, opts)
	}
	return rc.Client.Logs(groupID, appID, opts)
}

// SchemaModels calls the mocked SchemaModels implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) SchemaModels(groupID, appID, language string) ([]realm.SchemaModel, error) {
	if rc.SchemaModelsFn != nil {
		return rc.SchemaModelsFn(groupID, appID, language)
	}
	return rc.Client.SchemaModels(groupID, appID, language)
}

// AllTemplates calls the mocked AllTemplates implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) AllTemplates() (realm.Templates, error) {
	if rc.AllTemplatesFn != nil {
		return rc.AllTemplatesFn()
	}
	return rc.Client.AllTemplates()
}

// ClientTemplate calls the mocked ClientTemplate implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) ClientTemplate(groupID, appID, templateID string) (*zip.Reader, bool, error) {
	if rc.ClientTemplateFn != nil {
		return rc.ClientTemplateFn(groupID, appID, templateID)
	}
	return rc.Client.ClientTemplate(groupID, appID, templateID)
}

// CompatibleTemplates calls the mocked CompatibleTemplates implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) CompatibleTemplates(groupID, appID string) (realm.Templates, error) {
	if rc.CompatibleTemplatesFn != nil {
		return rc.CompatibleTemplatesFn(groupID, appID)
	}
	return rc.Client.CompatibleTemplates(groupID, appID)
}

// AllowedIPs calls the mocked AllowedIPs implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) AllowedIPs(groupID, appID string) ([]realm.AllowedIP, error) {
	if rc.AllowedIPsFn != nil {
		return rc.AllowedIPsFn(groupID, appID)
	}
	return rc.Client.AllowedIPs(groupID, appID)
}

// AllowedIPCreate calls the mocked AllowedIPCreate implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) AllowedIPCreate(groupID, appID, address, comment string, useCurrent bool) (realm.AllowedIP, error) {
	if rc.AllowedIPCreateFn != nil {
		return rc.AllowedIPCreateFn(groupID, appID, address, comment, useCurrent)
	}
	return rc.AllowedIPCreate(groupID, appID, address, comment, useCurrent)
}

// AllowedIPUpdate calls the mocked AllowedIPUpdate implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) AllowedIPUpdate(groupID, appID, allowedIPID, newAddress, newComment string) error {
	if rc.AllowedIPUpdateFn != nil {
		return rc.AllowedIPUpdateFn(groupID, appID, allowedIPID, newAddress, newComment)
	}
	return rc.Client.AllowedIPUpdate(groupID, appID, allowedIPID, newAddress, newComment)
}

// AllowedIPDelete calls the mocked AllowedIPDelete implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) AllowedIPDelete(groupID, appID, allowedIPID string) error {
	if rc.AllowedIPDeleteFn != nil {
		return rc.AllowedIPDeleteFn(groupID, appID, allowedIPID)
	}
	return rc.Client.AllowedIPDelete(groupID, appID, allowedIPID)
}

// Status calls the mocked Status implementation if provided,
// otherwise the call falls back to the underlying realm.Client implementation.
// NOTE: this may panic if the underlying realm.Client is left undefined
func (rc RealmClient) Status() error {
	if rc.StatusFn != nil {
		return rc.StatusFn()
	}
	return rc.Client.Status()
}
