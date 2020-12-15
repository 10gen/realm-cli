package realm

import (
	"fmt"
	"strconv"
)

var (
	// DefaultAppConfigVersion is the default app config version
	// TODO(REALMC-7653): switch this default version to AppConfigVersion20210101
	DefaultAppConfigVersion = AppConfigVersion20200603
)

// AppConfigVersion is the Realm application config version for import/export
type AppConfigVersion int

func (v AppConfigVersion) String() string { return strconv.Itoa(int(v)) }

// set of supported app config versions
const (
	AppConfigVersionNil      AppConfigVersion = 0
	AppConfigVersion20210101 AppConfigVersion = 20210101
	AppConfigVersion20200603 AppConfigVersion = 20200603
	AppConfigVersion20180301 AppConfigVersion = 20180301
)

// set of app structure filepath parts
const (
	FileAppConfig = "config.json"

	DirAuthProviders = "auth_providers"

	DirGraphQL                = "graphql"
	DirGraphQLCustomResolvers = DirGraphQL + "/custom_resolvers"
	FileGraphQLConfig         = DirGraphQL + "/config.json"
)

// FileAuthProvider creates the auth provider config filepath
func FileAuthProvider(name string) string {
	return fmt.Sprintf("%s/%s.json", DirAuthProviders, name)
}
