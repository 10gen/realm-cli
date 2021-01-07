package list

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"
)

const (
	flagState      = "state"
	flagStateShort = "s"
	flagStateUsage = `select the state of users to list, available options: ["enabled", "disabled"]`

	flagStatus      = "status-pending"
	flagStatusShort = "p"
	flagStatusUsage = `select the state of users to list, available options: ["enabled", "disabled"]`

	flagProviderTypes      = "provider"
	flagProviderTypesShort = "t"
	flagProviderTypesUsage = `todo add description`

	flagUsers      = "users"
	flagUsersShort = "u" //idk what you want this to be
	flagUsersUsage = `todo add description`
)

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

func writeAsCSV(vals []string) string {
	b := &bytes.Buffer{}
	w := csv.NewWriter(b)
	err := w.Write(vals)
	if err != nil {
		return ""
	}
	w.Flush()
	return strings.TrimSuffix(b.String(), "\n")
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
	str := writeAsCSV(*p.value)
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
