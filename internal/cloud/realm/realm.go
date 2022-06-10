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

// set of known config version, location, provider region and deployment values
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
	ProviderRegionValues = []string{
		AWSProviderRegionUSEast1.String(),
		AWSProviderRegionUSWest2.String(),
		AWSProviderRegionEUCentral1.String(),
		AWSProviderRegionEUWest1.String(),
		AWSProviderRegionAPSoutheast1.String(),
		AWSProviderRegionAPSoutheast2.String(),
		AWSProviderRegionAPSouth1.String(),
		AWSProviderRegionUSEast2.String(),
		AWSProviderRegionEUWest2.String(),
		AWSProviderRegionSAEast1.String(),
		AzureProviderRegionEastUS2.String(),
		AzureProviderRegionWestUS.String(),
		AzureProviderRegionWestEurope.String(),
		AzureProviderRegionEastAsia.String(),
		AzureProviderRegionSouthEastAsia.String(),
	}
	ProviderRegionValuesByCloudProvider = map[string][]string{
		CloudProviderAWS: {
			AWSProviderRegionUSEast1.Label(),
			AWSProviderRegionUSWest2.Label(),
			AWSProviderRegionEUCentral1.Label(),
			AWSProviderRegionEUWest1.Label(),
			AWSProviderRegionAPSoutheast1.Label(),
			AWSProviderRegionAPSoutheast2.Label(),
			AWSProviderRegionAPSouth1.Label(),
			AWSProviderRegionUSEast2.Label(),
			AWSProviderRegionEUWest2.Label(),
			AWSProviderRegionSAEast1.Label(),
		},

		CloudProviderAzure: {
			AzureProviderRegionEastUS2.Label(),
			AzureProviderRegionWestUS.Label(),
			AzureProviderRegionWestEurope.Label(),
			AzureProviderRegionEastAsia.Label(),
			AzureProviderRegionSouthEastAsia.Label(),
		},
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
	errInvalidProviderRegion  = fmt.Errorf("unsupported provider region, use one of [%s] instead", strings.Join(ProviderRegionValues, ", "))
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
	LocationSaoPaolo  Location = "BR-SP"
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
		LocationSaoPaolo,
		LocationSydney,
		LocationMumbai,
		LocationSingapore:
		return true
	}
	return false
}

// Set of supported cloud providers
const (
	CloudProviderAWS   = "aws"
	CloudProviderAzure = "azure"
	CloudProviderGCP   = "gcp"
)

// TODO(BAAS-xxxx) add ticket number here if we want to do gcp separately
var CloudProviderValues = []string{CloudProviderAWS, CloudProviderAzure}

// The provider regions that we currently support
const (
	ProviderRegionEmpty ProviderRegion = ""

	// AWS provider region that we currently support
	AWSProviderRegionUSEast1      ProviderRegion = "aws-us-east-1"
	AWSProviderRegionUSWest2      ProviderRegion = "aws-us-west-2"
	AWSProviderRegionEUCentral1   ProviderRegion = "aws-eu-central-1"
	AWSProviderRegionEUWest1      ProviderRegion = "aws-eu-west-1"
	AWSProviderRegionAPSoutheast1 ProviderRegion = "aws-ap-southeast-1"
	AWSProviderRegionAPSoutheast2 ProviderRegion = "aws-ap-southeast-2"
	AWSProviderRegionAPSouth1     ProviderRegion = "aws-ap-south-1"
	AWSProviderRegionUSEast2      ProviderRegion = "aws-us-east-2"
	AWSProviderRegionEUWest2      ProviderRegion = "aws-eu-west-2"
	AWSProviderRegionSAEast1      ProviderRegion = "aws-sa-east-1"

	// Azure provider regions that we currently support
	AzureProviderRegionEastUS2       ProviderRegion = "azure-eastus2"
	AzureProviderRegionWestUS        ProviderRegion = "azure-westus"
	AzureProviderRegionWestEurope    ProviderRegion = "azure-westeurope"
	AzureProviderRegionEastAsia      ProviderRegion = "azure-eastasia"
	AzureProviderRegionSouthEastAsia ProviderRegion = "azure-southeastasia"

	// GCP provider regions that we currently support
	GCPProviderRegionUSCentral1  ProviderRegion = "gcp-us-central1"
	GCPProviderRegionUSWest1     ProviderRegion = "gcp-us-west1"
	GCPProviderRegionEuropeWest1 ProviderRegion = "gcp-europe-west1"
	GCPProviderRegionAsiaSouth1  ProviderRegion = "gcp-asia-south1"
)

// ProviderRegionToLocation provides a quick lookup of a provider and region combination to
// location values. This is necessary as the backing clusters no longer match our current
// operating regions. In order to do backing cluster storage, we must leverage a lookup here.
var ProviderRegionToLocation = map[ProviderRegion]Location{
	ProviderRegionEmpty:              LocationVirginia,
	AWSProviderRegionUSEast1:         LocationVirginia,
	AWSProviderRegionUSWest2:         LocationOregon,
	AWSProviderRegionEUCentral1:      LocationFrankfurt,
	AWSProviderRegionEUWest1:         LocationIreland,
	AWSProviderRegionAPSoutheast1:    LocationSingapore,
	AWSProviderRegionAPSoutheast2:    LocationSydney,
	AWSProviderRegionAPSouth1:        LocationMumbai,
	AWSProviderRegionUSEast2:         LocationVirginia,
	AWSProviderRegionEUWest2:         LocationIreland,
	AWSProviderRegionSAEast1:         LocationSaoPaolo,
	AzureProviderRegionEastUS2:       LocationVirginia,
	AzureProviderRegionWestUS:        LocationOregon,
	AzureProviderRegionWestEurope:    LocationFrankfurt,
	AzureProviderRegionEastAsia:      LocationMumbai,
	AzureProviderRegionSouthEastAsia: LocationSingapore,
	GCPProviderRegionUSCentral1:      LocationVirginia,
	GCPProviderRegionUSWest1:         LocationOregon,
	GCPProviderRegionEuropeWest1:     LocationFrankfurt,
	GCPProviderRegionAsiaSouth1:      LocationMumbai,
}

// ProviderRegion is the Realm app provider region
type ProviderRegion string

// String returns the ProviderRegion display
func (p ProviderRegion) String() string { return string(p) }

// Label returns the ProviderRegion display without the cloud provider
func (p ProviderRegion) Label() string {
	return p.String()[strings.IndexByte(p.String(), '-')+1:]
}

// Type returns the ProviderRegion type
func (p ProviderRegion) Type() string { return flags.TypeString }

// Set validates and sets the ProviderRegion value
func (p *ProviderRegion) Set(val string) error {
	newProviderRegion := ProviderRegion(strings.ToLower(val))

	if !isValidProviderRegion(newProviderRegion) {
		return errInvalidProviderRegion
	}

	*p = newProviderRegion
	return nil
}

// WriteAnswer validates and sets the ProviderRegion value
func (p *ProviderRegion) WriteAnswer(name string, value interface{}) error {
	var newProviderRegion ProviderRegion

	switch v := value.(type) {
	case core.OptionAnswer:
		newProviderRegion = ProviderRegion(v.Value)
	}

	if !isValidProviderRegion(newProviderRegion) {
		return errInvalidProviderRegion
	}
	*p = newProviderRegion
	return nil
}

func isValidProviderRegion(p ProviderRegion) bool {
	switch p {
	case
		ProviderRegionEmpty,
		AWSProviderRegionUSEast1,
		AWSProviderRegionUSWest2,
		AWSProviderRegionEUCentral1,
		AWSProviderRegionEUWest1,
		AWSProviderRegionAPSoutheast1,
		AWSProviderRegionAPSoutheast2,
		AWSProviderRegionAPSouth1,
		AWSProviderRegionUSEast2,
		AWSProviderRegionEUWest2,
		AWSProviderRegionSAEast1,

		AzureProviderRegionEastUS2,
		AzureProviderRegionWestUS,
		AzureProviderRegionWestEurope,
		AzureProviderRegionEastAsia,
		AzureProviderRegionSouthEastAsia,

		GCPProviderRegionUSCentral1,
		GCPProviderRegionUSWest1,
		GCPProviderRegionEuropeWest1,
		GCPProviderRegionAsiaSouth1:
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

// WriteAnswer validates and sets the Environment value
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
