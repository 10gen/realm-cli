package flags

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/pflag"
)

const (
	//TypeString is the type of strings
	TypeString = "string"
)

// enumSetValue is modified copy of stringSliceValue from
// the cobra pflags package
type enumSetValue struct {
	value          *[]string
	changed        bool
	validValuesSet map[string]bool
}

// NewEnumSetValue creates a stringSliceValue that validates inputs and prevents duplicates
func NewEnumSetValue(p *[]string, val []string, validValues []string) pflag.Value {
	esv := new(enumSetValue)
	esv.value = p
	*esv.value = val
	esv.validValuesSet = make(map[string]bool)
	for _, validValue := range validValues {
		esv.validValuesSet[validValue] = true
	}
	return esv
}

func (esv *enumSetValue) Set(val string) error {
	v, err := readAsCSV(val)
	if err != nil {
		return err
	}
	if !esv.changed {
		*esv.value = v
	} else {
		*esv.value = append(*esv.value, v...)
	}
	esv.changed = true
	return esv.validateAndRemoveDuplicates()
}

func (esv *enumSetValue) Type() string {
	return "stringSlice"
}

func (esv *enumSetValue) String() string {
	str, _ := writeAsCSV(*esv.value)
	return "[" + str + "]"
}

func (esv *enumSetValue) Append(val string) error {
	*esv.value = append(*esv.value, val)
	return esv.validateAndRemoveDuplicates()
}

func (esv *enumSetValue) Replace(val []string) error {
	*esv.value = val
	return esv.validateAndRemoveDuplicates()
}

func (esv *enumSetValue) GetSlice() []string {
	return *esv.value
}

func (esv *enumSetValue) validateAndRemoveDuplicates() error {
	valueSet := make(map[string]bool)
	for _, value := range *esv.value {
		if !esv.validValuesSet[value] {
			return esv.errInvalidEnumValue()
		}
		valueSet[value] = true
	}

	newValues := make([]string, 0, len(valueSet))
	for value := range valueSet {
		newValues = append(newValues, value)
	}
	*esv.value = newValues
	return nil
}

func (esv *enumSetValue) errInvalidEnumValue() error {
	validEnumValues := make([]string, 0, len(esv.validValuesSet))
	for validEnumValue := range esv.validValuesSet {
		validEnumValues = append(validEnumValues, validEnumValue)
	}
	// sort to ensure a consistent error message
	sort.Strings(validEnumValues)
	return fmt.Errorf("unsupported value, use one of [%s] instead", strings.Join(validEnumValues, ", "))
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
