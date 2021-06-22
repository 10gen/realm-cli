package realm

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/10gen/realm-cli/internal/utils/flags"

	"github.com/AlecAivazis/survey/v2/core"
)

var (
	// DefaultAppConfigVersion is the default app config version
	DefaultAppConfigVersion = AppConfigVersion20210101
)

// AppConfigVersion is the Realm application config version for import/export
type AppConfigVersion int

func (cv AppConfigVersion) String() string { return strconv.Itoa(int(cv)) }

// Type returns the DeploymentModel type
func (cv AppConfigVersion) Type() string { return flags.TypeInt }

// Set validates and sets the deployment model value
func (cv *AppConfigVersion) Set(val string) error {
	v, err := strconv.Atoi(val)
	if err != nil {
		return err
	}
	newConfigVersion := AppConfigVersion(v)

	if !isValidConfigVersion(newConfigVersion) {
		return errInvalidConfigVersion
	}

	*cv = newConfigVersion
	return nil
}

// set of supported app config versions
const (
	AppConfigVersionZero     AppConfigVersion = 0
	AppConfigVersion20210101 AppConfigVersion = 20210101
	AppConfigVersion20200603 AppConfigVersion = 20200603
	AppConfigVersion20180301 AppConfigVersion = 20180301
)

func isValidConfigVersion(cv AppConfigVersion) bool {
	switch cv {
	case
		AppConfigVersionZero, // allow ConfigVersion to be optional
		AppConfigVersion20180301,
		AppConfigVersion20200603,
		AppConfigVersion20210101:
		return true
	}
	return false
}

// set of known config version, location and deployment values
var (
	ConfigVersionValues = []string{
		AppConfigVersion20180301.String(),
		AppConfigVersion20200603.String(),
		AppConfigVersion20210101.String(),
	}
	DeploymentModelValues = []string{
		DeploymentModelGlobal.String(),
		DeploymentModelLocal.String(),
	}
	LocationValues = []string{
		LocationVirginia.String(),
		LocationOregon.String(),
		LocationFrankfurt.String(),
		LocationIreland.String(),
		LocationSydney.String(),
		LocationMumbai.String(),
		LocationSingapore.String(),
	}
	EnvironmentValues = []string{
		EnvironmentDevelopment.String(),
		EnvironmentTesting.String(),
		EnvironmentQA.String(),
		EnvironmentProduction.String(),
	}

	errInvalidConfigVersion   = fmt.Errorf("unsupported config version, use one of [%s] instead", strings.Join(ConfigVersionValues, ", "))
	errInvalidDeploymentModel = fmt.Errorf("unsupported deployment model, use one of [%s] instead", strings.Join(DeploymentModelValues, ", "))
	errInvalidLocation        = fmt.Errorf("unsupported location, use one of [%s] instead", strings.Join(LocationValues, ", "))
	errInvalidEnvironment     = fmt.Errorf("unsupported environment, use one of [%s] instead", strings.Join(EnvironmentValues, ", "))
)

// DeploymentModel is the Realm app deployment model
type DeploymentModel string

// String returns the deployment model display
func (dm DeploymentModel) String() string { return string(dm) }

// Type returns the DeploymentModel type
func (dm DeploymentModel) Type() string { return flags.TypeString }

// Set validates and sets the deployment model value
func (dm *DeploymentModel) Set(val string) error {
	newDeploymentModel := DeploymentModel(strings.ToUpper(val))

	if !isValidDeploymentModel(newDeploymentModel) {
		return errInvalidDeploymentModel
	}

	*dm = newDeploymentModel
	return nil
}

// WriteAnswer validates and sets the deployment model value
func (dm *DeploymentModel) WriteAnswer(name string, value interface{}) error {
	var newDeploymentModel DeploymentModel

	switch v := value.(type) {
	case core.OptionAnswer:
		newDeploymentModel = DeploymentModel(v.Value)
	}

	if !isValidDeploymentModel(newDeploymentModel) {
		return errInvalidDeploymentModel
	}
	*dm = newDeploymentModel
	return nil
}

// set of supported Realm app deployment models
const (
	DeploymentModelEmpty  DeploymentModel = ""
	DeploymentModelGlobal DeploymentModel = "GLOBAL"
	DeploymentModelLocal  DeploymentModel = "LOCAL"
)

func isValidDeploymentModel(dm DeploymentModel) bool {
	switch dm {
	case
		DeploymentModelEmpty, // allow DeploymentModel to be optional
		DeploymentModelGlobal,
		DeploymentModelLocal:
		return true
	}
	return false
}

// Location is the Realm app location
type Location string

// String returns the Location display
func (l Location) String() string { return string(l) }

// Type returns the Location type
func (l Location) Type() string { return flags.TypeString }

// Set validates and sets the Location value
func (l *Location) Set(val string) error {
	newLocation := Location(strings.ToUpper(val))

	if !isValidLocation(newLocation) {
		return errInvalidLocation
	}

	*l = newLocation
	return nil
}

// WriteAnswer validates and sets the Location value
func (l *Location) WriteAnswer(name string, value interface{}) error {
	var newLocation Location

	switch v := value.(type) {
	case core.OptionAnswer:
		newLocation = Location(v.Value)
	}

	if !isValidLocation(newLocation) {
		return errInvalidLocation
	}
	*l = newLocation
	return nil
}

// set of supported Realm app locations
const (
	LocationEmpty     Location = ""
	LocationVirginia  Location = "US-VA"
	LocationOregon    Location = "US-OR"
	LocationFrankfurt Location = "DE-FF"
	LocationIreland   Location = "IE"
	LocationSydney    Location = "AU"
	LocationMumbai    Location = "IN-MB"
	LocationSingapore Location = "SG"
)

func isValidLocation(l Location) bool {
	switch l {
	case
		LocationEmpty, // allow Location to be optional
		LocationVirginia,
		LocationOregon,
		LocationFrankfurt,
		LocationIreland,
		LocationSydney,
		LocationMumbai,
		LocationSingapore:
		return true
	}
	return false
}

// Environment is the Realm app environment
type Environment string

// String returns the Environment display
func (e Environment) String() string { return string(e) }

// Type returns the Environment type
func (e Environment) Type() string { return flags.TypeString }

// Set validates and sets the Environment value
func (e *Environment) Set(val string) error {
	newEnvironment := Environment(strings.ToLower(val))

	if !isValidEnvironment(newEnvironment) {
		return errInvalidEnvironment
	}

	*e = newEnvironment
	return nil
}

// WriteAnswer validates and sets the Location value
func (e *Environment) WriteAnswer(name string, value interface{}) error {
	var newEnvironment Environment

	switch v := value.(type) {
	case core.OptionAnswer:
		newEnvironment = Environment(v.Value)
	}

	if !isValidEnvironment(newEnvironment) {
		return errInvalidEnvironment
	}
	*e = newEnvironment
	return nil
}

// set of supported Realm app environments
const (
	EnvironmentNone        Environment = ""
	EnvironmentDevelopment Environment = "development"
	EnvironmentTesting     Environment = "testing"
	EnvironmentQA          Environment = "qa"
	EnvironmentProduction  Environment = "production"
)

func isValidEnvironment(e Environment) bool {
	switch e {
	case
		EnvironmentNone, // no Environment is default
		EnvironmentDevelopment,
		EnvironmentTesting,
		EnvironmentQA,
		EnvironmentProduction:
		return true
	}
	return false
}
