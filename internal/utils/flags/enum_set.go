package flags

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"sort"
	"strings"
)

// set of known flag types
const (
	TypeInt    = "int"
	TypeString = "string"
)

var setMember struct{}

// EnumSetValue is a modified copy of stringSliceValue from
// the cobra pflags package.  It validates a set of enum values.
// Note: it does NOT maintain order of the input arguments
type EnumSetValue struct {
	values         *[]string
	valuesSet      map[string]struct{}
	validValues    []interface{}
	validValuesSet map[string]struct{}
}

// NewEnumSet creates an EnumSetValue.  Expects p to point to an empty
// slice and will clear it otherwise.
func NewEnumSet(p *[]string, validValues []interface{}) *EnumSetValue {
	esv := new(EnumSetValue)
	esv.values = p
	*esv.values = nil
	esv.validValues = validValues
	esv.validValuesSet = make(map[string]struct{}, len(validValues))
	for _, validValue := range validValues {
		sVal := fmt.Sprintf("%v", validValue)
		esv.validValuesSet[sVal] = setMember
	}
	esv.valuesSet = make(map[string]struct{})
	return esv
}

// Set adds new values to the EnumSetValue
func (esv *EnumSetValue) Set(val string) error {
	values, err := readAsCSV(val)
	if err != nil {
		return err
	}

	return esv.set(values...)
}

// Type returns the type string of EnumSetValue
func (esv *EnumSetValue) Type() string {
	return "enumSet"
}

// String returns a string representation of an EnumSetValue
func (esv *EnumSetValue) String() string {
	str, err := writeAsCSV(*esv.values)
	if err != nil {
		return "[]"
	}
	return "[" + str + "]"
}

// Append appends a single value to an EnumSetValue
func (esv *EnumSetValue) Append(val string) error {
	return esv.set(val)
}

// Replace replaces all values in an EnumSetValue
func (esv *EnumSetValue) Replace(values []string) error {
	esv.valuesSet = make(map[string]struct{}, len(values))
	return esv.set(values...)
}

// GetSlice returns the underlying slice of the set of the EnumSetValue
func (esv *EnumSetValue) GetSlice() []string {
	return *esv.values
}

func (esv *EnumSetValue) set(values ...string) error {
	*esv.values = append(*esv.values, values...)
	for _, value := range values {
		esv.valuesSet[value] = setMember
	}
	return esv.validateAndRemoveDuplicates()
}

func (esv *EnumSetValue) validateAndRemoveDuplicates() error {
	for _, value := range *esv.values {
		if _, exists := esv.validValuesSet[value]; !exists {
			return esv.errInvalidEnumValue()
		}
	}

	newValues := make([]string, 0, len(esv.valuesSet))
	for value := range esv.valuesSet {
		newValues = append(newValues, value)
	}
	// sort to ensure determinism
	sort.Strings(newValues)

	*esv.values = newValues
	return nil
}

func (esv *EnumSetValue) errInvalidEnumValue() error {
	var sb strings.Builder
	sb.WriteString(`unsupported value, use one of ["`)
	for i, v := range esv.validValues {
		if i != 0 {
			sb.WriteString(`", "`)
		}
		sb.WriteString(fmt.Sprintf("%v", v))
	}
	sb.WriteString(`"] instead`)
	return errors.New(sb.String())
}

// readAsCSV is copied from the cobra pflags package
func readAsCSV(val string) ([]string, error) {
	if val == "" {
		return []string{}, nil
	}
	stringReader := strings.NewReader(val)
	csvReader := csv.NewReader(stringReader)
	return csvReader.Read()
}

// writeAsCSV is copied from the cobra pflags package
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
