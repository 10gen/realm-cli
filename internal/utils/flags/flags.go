package flags

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"
)

const (
	//TypeString is the type of strings
	TypeString = "string"
)

type sliceValue interface {
	pflag.Value
	pflag.SliceValue
}

type enumSliceValue struct {
	stringSliceValue sliceValue
	validEnumValues  map[string]bool
	enumValues       *[]string
}

// NewEnumSliceValue creates a pflag value that validates enums
func NewEnumSliceValue(p *[]string, validEnumValues []string, defaultValues []string) pflag.Value {
	esv := new(enumSliceValue)
	fs := pflag.FlagSet{}
	name := "name"
	fs.StringSliceVar(p, name, defaultValues, "")
	esv.stringSliceValue = fs.Lookup(name).Value.(sliceValue)
	esv.validEnumValues = make(map[string]bool)
	for _, validEnumValue := range validEnumValues {
		esv.validEnumValues[validEnumValue] = true
	}
	esv.enumValues = p
	return esv
}

func (esv *enumSliceValue) Set(val string) error {
	err := esv.stringSliceValue.Set(val)
	if err != nil {
		return err
	}
	for _, enumValue := range *esv.enumValues {
		if !esv.validEnumValues[enumValue] {
			return esv.errInvalidEnumValue()
		}
	}
	return nil
}

func (esv *enumSliceValue) Type() string {
	return "stringSlice"
}

func (esv *enumSliceValue) String() string {
	return esv.stringSliceValue.String()
}

func (esv *enumSliceValue) Append(val string) error {
	if !esv.validEnumValues[val] {
		return esv.errInvalidEnumValue()
	}
	return esv.stringSliceValue.Append(val)
}

func (esv *enumSliceValue) Replace(val []string) error {
	err := esv.stringSliceValue.Replace(val)
	if err != nil {
		return err
	}
	for _, enumValue := range *esv.enumValues {
		if !esv.validEnumValues[enumValue] {
			return esv.errInvalidEnumValue()
		}
	}
	return nil
}

func (esv *enumSliceValue) GetSlice() []string {
	return esv.stringSliceValue.GetSlice()
}

func (esv *enumSliceValue) errInvalidEnumValue() error {
	validEnumValues := make([]string, 0, len(esv.validEnumValues))
	for validEnumValue := range esv.validEnumValues {
		validEnumValues = append(validEnumValues, validEnumValue)
	}
	return fmt.Errorf("unsupported value, use one of [%s] instead", strings.Join(validEnumValues, ", "))
}
