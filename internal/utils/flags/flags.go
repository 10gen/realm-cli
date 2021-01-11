package flags

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"sort"
	"strings"
)

const (
	//TypeString is the type of strings
	TypeString = "string"
)

var setMember struct{}

// EnumSetValue is a modified copy of stringSliceValue from
// the cobra pflags package.  It validates a set of enum values.
// Note: it does NOT maintain order of the input arguments
type EnumSetValue struct {
	values         *[]string
	valuesSet      map[string]struct{}
	validValues    []string
	validValuesSet map[string]struct{}
}

// NewEnumSet creates an EnumSetValue
func NewEnumSet(p *[]string, defaultValues []string, validValues []string) *EnumSetValue {
	esv := new(EnumSetValue)
	esv.values = p
	*esv.values = defaultValues
	esv.validValues = validValues
	esv.validValuesSet = make(map[string]struct{}, len(validValues))
	for _, validValue := range validValues {
		esv.validValuesSet[validValue] = setMember
	}
	esv.valuesSet = make(map[string]struct{}, len(defaultValues))
	for _, value := range defaultValues {
		esv.valuesSet[value] = setMember
	}
	err := esv.validateAndRemoveDuplicates()
	if err != nil {
		// what should I do here?  I'm pretty sure this would be a great example of when
		// to panic (coding error in a single threaded, non-critical application) but I know
		// you don't like that so lmk what to do.
	}
	return esv
}

// Set adds new values to the EnumSetValue
func (esv *EnumSetValue) Set(val string) error {
	values, err := readAsCSV(val)
	if err != nil {
		return err
	}

	*esv.values = append(*esv.values, values...)
	for _, value := range values {
		esv.valuesSet[value] = setMember
	}
	return esv.validateAndRemoveDuplicates()
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
	*esv.values = append(*esv.values, val)
	esv.valuesSet[val] = setMember
	return esv.validateAndRemoveDuplicates()
}

// Replace replaces all values in an EnumSetValue
func (esv *EnumSetValue) Replace(values []string) error {
	*esv.values = values
	esv.valuesSet = make(map[string]struct{}, len(values))
	for _, value := range values {
		esv.valuesSet[value] = setMember
	}
	return esv.validateAndRemoveDuplicates()
}

// GetSlice returns the underlying slice of the set of the EnumSetValue
func (esv *EnumSetValue) GetSlice() []string {
	return *esv.values
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
	return fmt.Errorf(`unsupported value, use one of ["%s"] instead`, strings.Join(esv.validValues, `", "`))
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
