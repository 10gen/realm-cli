package list

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"

	"github.com/10gen/realm-cli/internal/cloud/realm"
)

const (
	flagState      = "state"
	flagStateUsage = `select the state of users to list, available options: ["enabled", "disabled"]`

	flagStatus      = "status-pending"
	flagStatusUsage = `select the state of users to list, available options: ["enabled", "disabled"]`

	flagProviderTypes      = "provider-types"
	flagProviderTypesUsage = `todo add description`
)

type stateValue string

// String returns the state filter
func (sv stateValue) String() string { return string(sv) }

// Type returns the state type
func (sv stateValue) Type() string { return "string" }

// Set validates and sets the user type value
func (sv *stateValue) Set(val string) error {
	newStateValue := stateValue(val)

	if !isValidState(newStateValue) {
		return errInvalidState
	}

	*sv = newStateValue
	return nil
}

func (sv stateValue) getUserState() realm.UserState {
	switch sv {
	case stateEnabled:
		return realm.UserStateEnabled
	case stateDisabled:
		return realm.UserStateDisabled
	}
	return realm.UserStateNil
}

const (
	stateNil      stateValue = ""
	stateEnabled  stateValue = "enabled"
	stateDisabled stateValue = "disabled"
)

var (
	errInvalidState = func() error {
		allStateTypes := []string{stateEnabled.String(), stateDisabled.String()}
		return fmt.Errorf("unsupported value, use one of [%s] instead", strings.Join(allStateTypes, ", "))
	}()
)

func isValidState(sv stateValue) bool {
	switch sv {
	case
		stateNil, // allow state to be optional
		stateEnabled,
		stateDisabled:
		return true
	}
	return false
}

type providerTypesValue struct {
	value   *[]string
	changed bool
}

func newProviderTypesValue(p *[]string) *providerTypesValue {
	ptv := new(providerTypesValue)
	ptv.value = p
	*ptv.value = []string{}
	return ptv
}

func readAsCSV(val string) ([]string, error) {
	if val == "" {
		return []string{}, nil
	}
	stringReader := strings.NewReader(val)
	csvReader := csv.NewReader(stringReader)
	providerTypes, err := csvReader.Read()
	if err != nil {
		return nil, err
	}
	for _, providerType := range providerTypes {
		if !isValidProviderType(providerType) {
			return nil, errInvalidProviderType
		}
	}
	return providerTypes, nil
}

func writeAsCSV(vals []string) (string, error) {
	b := &bytes.Buffer{}
	w := csv.NewWriter(b)
	err := w.Write(vals)
	if err != nil {
		return "", err
	}
	w.Flush()
	return strings.TrimSuffix(b.String(), "\n"), nil
}

func (p *providerTypesValue) Set(val string) error {
	v, err := readAsCSV(val)
	if err != nil {
		return err
	}
	if !p.changed {
		*p.value = v
	} else {
		*p.value = append(*p.value, v...)
	}
	p.changed = true
	return nil
}

func (p *providerTypesValue) Type() string {
	return "stringSlice"
}

func (p *providerTypesValue) String() string {
	str, _ := writeAsCSV(*p.value)
	return "[" + str + "]"
}

func (p *providerTypesValue) Append(val string) error {
	if !isValidProviderType(val) {
		return errInvalidProviderType
	}
	*p.value = append(*p.value, val)
	return nil
}

func (p *providerTypesValue) Replace(val []string) error {
	for _, providerType := range val {
		if !isValidProviderType(providerType) {
			return errInvalidProviderType
		}
	}
	*p.value = val
	return nil
}

func (p *providerTypesValue) GetSlice() []string {
	return *p.value
}

const (
	providerTypeLocalUserPass string = "local-userpass"
	providerTypeAPIKey        string = "api-key"
	providerTypeFacebook      string = "oauth2-facebook"
	providerTypeGoogle        string = "oauth2-google"
	providerTypeAnonymous     string = "anon-user"
	providerTypeCustom        string = "custom-token"
)

var (
	errInvalidProviderType = func() error {
		allProviderTypes := []string{
			providerTypeLocalUserPass,
			providerTypeAPIKey,
			providerTypeFacebook,
			providerTypeGoogle,
			providerTypeAnonymous,
			providerTypeCustom,
		}
		return fmt.Errorf("unsupported value, use one of [%s] instead", strings.Join(allProviderTypes, ", "))
	}()
)

func isValidProviderType(provider string) bool {
	switch provider {
	case
		providerTypeLocalUserPass,
		providerTypeAPIKey,
		providerTypeFacebook,
		providerTypeGoogle,
		providerTypeAnonymous,
		providerTypeCustom:
		return true
	}
	return false
}
