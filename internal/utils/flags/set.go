package flags

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"sort"
	"strings"
)

type stringSet struct {
	values        *[]string
	valueSet      map[string]struct{}
	allowedValues []string
	validValueSet map[string]struct{}
}

// NewStringSet returns a new string set that accepts unique values from a flag
// The values arg should be a pointer to a string array and if it has contents
// then each pre-existing entry MUST be unique
func newStringSet(values *[]string, validValues []string) *stringSet {
	set := stringSet{values: values}

	set.valueSet = map[string]struct{}{}
	if values != nil {
		for _, value := range *values {
			set.valueSet[value] = struct{}{}
		}
	}

	set.allowedValues = make([]string, 0, len(validValues))
	set.validValueSet = map[string]struct{}{}
	for _, validValue := range validValues {
		set.allowedValues = append(set.allowedValues, fmt.Sprintf("'%s'", validValue))
		set.validValueSet[validValue] = struct{}{}
	}

	return &set
}

func (set stringSet) Type() string { return "Set" }

func (set stringSet) String() string {
	out := new(bytes.Buffer)

	w := csv.NewWriter(out)
	if err := w.Write(*set.values); err != nil {
		return "[]"
	}
	w.Flush()
	return fmt.Sprintf("[%s]", strings.TrimSuffix(out.String(), "\n"))
}

func (set *stringSet) Append(val string) error {
	if val == "" {
		return nil
	}
	return set.add(val)
}

func (set *stringSet) Replace(vals []string) error {
	*set.values = nil
	set.valueSet = map[string]struct{}{}
	return set.add(vals...)
}

func (set *stringSet) Set(val string) error {
	if val == "" {
		return nil
	}

	vals, err := csv.NewReader(strings.NewReader(val)).Read()
	if err != nil {
		return err
	}

	return set.add(vals...)
}

func (set *stringSet) add(vals ...string) error {
	values := make([]string, 0, len(set.valueSet)+len(vals))
	values = append(values, *set.values...)

	for _, val := range vals {
		if len(set.allowedValues) > 0 {
			if _, ok := set.validValueSet[val]; !ok {
				return fmt.Errorf("'%s' is an unsupported value, try instead one of [%s]", val, strings.Join(set.allowedValues, ", "))
			}
		}
		if _, ok := set.valueSet[val]; ok {
			continue
		}
		values = append(values, val)
		set.valueSet[val] = struct{}{}
	}

	sort.Strings(values) // ensure deterministic set
	*set.values = values
	return nil
}
