package flags

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"
)

// set of known flag types
const (
	TypeInt    = "Integer"
	TypeString = "String"
)

// Flag is capable of registering a CLI flag
type Flag interface {
	Register(fs *pflag.FlagSet)
}

// Meta represents the metadata of a flag
type Meta struct {
	Name      string
	Shorthand string
	Usage     Usage
	Hidden    bool
	Deprecate string
}

// Usage represents the details of a flag's usage
type Usage struct {
	// Description should be a brief message about the flag's intended usage
	// It should not necessitate the use of multiple sentences nor end with punctuation
	// It should begin with a capitalized letter
	Description string
	// AllowedFormat should be the flag's accepted value pattern
	// (e.g. with Dates what 'yyyy-mm-dd' pattern is supported)
	// Note: this value will be wrapped with quotes
	AllowedFormat string
	// Default value should be the flag's default value, wrapped with ""
	// Alternatively, enclosing a description of the value with <> can suffice (e.g. <none>)
	DefaultValue string
	// AllowedValues should be the flag's array of allowed values, each wrapped with ""
	// Alternatively, using <> to describe a value can suffice (e.g. <none>, "one", "two")
	AllowedValues []string
	// Note is a separate message that can be used to describe any important caveats or nuances
	// It should follow similar conventions to Description, where it beings with a capitalized letter
	// and consists of only a brief message
	Note string
	// DoscLink is a link to MongoDB CLI docummentation for the flag
	// These should be leveraged when more than just a short blurb would suffice for explaining
	// the flags' usage (in case it is difficult to keep Description to a brief message)
	DocsLink string
}

func (u Usage) String() string {
	parts := make([]string, 0, 4)

	parts = append(parts, u.Description)

	if u.Note != "" {
		parts = append(parts, fmt.Sprintf("(Note: %s)", u.Note))
	}

	if u.DefaultValue != "" || len(u.AllowedValues) > 0 {
		valueParts := make([]string, 0, 3)

		if u.AllowedFormat != "" {
			valueParts = append(valueParts, fmt.Sprintf("Allowed format: %q", u.AllowedFormat))
		}

		if u.DefaultValue != "" {
			valueParts = append(valueParts, "Default value: "+u.DefaultValue)
		}

		if len(u.AllowedValues) > 0 {
			valueParts = append(valueParts, "Allowed values: "+strings.Join(u.AllowedValues, ", "))
		}

		parts = append(parts, fmt.Sprintf("(%s)", strings.Join(valueParts, "; ")))
	}

	if u.DocsLink != "" {
		parts = append(parts, fmt.Sprintf("[Learn more: %s]", u.DocsLink))
	}

	return strings.Join(parts, " ")
}

// BoolFlag is a boolean flag
type BoolFlag struct {
	Meta
	Value        *bool
	DefaultValue bool
}

// Register registers the boolean flag with the provided flag set
func (f BoolFlag) Register(fs *pflag.FlagSet) {
	if f.Shorthand == "" {
		fs.BoolVar(f.Value, f.Name, f.DefaultValue, f.Usage.String())
	} else {
		fs.BoolVarP(f.Value, f.Name, f.Shorthand, f.DefaultValue, f.Usage.String())
	}

	registerFlag(fs, f.Meta)
}

// CustomFlag is a custom flag
type CustomFlag struct {
	Meta
	Value pflag.Value
}

// Register registers the custom flag with the provided flag set
func (f CustomFlag) Register(fs *pflag.FlagSet) {
	if f.Shorthand == "" {
		fs.Var(f.Value, f.Name, f.Usage.String())
	} else {
		fs.VarP(f.Value, f.Name, f.Shorthand, f.Usage.String())
	}

	registerFlag(fs, f.Meta)
}

// StringSetOptions are the options available when creating a new string set flag.
// The DefaultValue and AllowedValues fields of these options (found in Meta.Usage)
// are ignored and instead are derived from the initially provided values and
// ValidValues, respectively
type StringSetOptions struct {
	Meta
	ValidValues []string
}

const (
	valueNone = "<none>"
)

// NewStringSetFlag returns a new string set flag from the provided options
func NewStringSetFlag(values *[]string, opts StringSetOptions) CustomFlag {
	var defaultValue string
	if values == nil || len(*values) == 0 {
		defaultValue = valueNone
	} else {
		defaultValues := make([]string, 0, len(*values))
		for _, value := range *values {
			defaultValues = append(defaultValues, fmt.Sprintf("%q", value))
		}
		defaultValue = fmt.Sprintf("[%s]", strings.Join(defaultValues, ","))
	}

	allowedValues := make([]string, 0, len(opts.ValidValues)+1)
	if defaultValue == valueNone {
		allowedValues = append(allowedValues, valueNone)
	}
	for _, validValue := range opts.ValidValues {
		allowedValues = append(allowedValues, fmt.Sprintf("%q", validValue))
	}

	return CustomFlag{
		Value: newStringSet(values, opts.ValidValues),
		Meta: Meta{
			Name:      opts.Name,
			Shorthand: opts.Shorthand,
			Usage: Usage{
				Description:   opts.Usage.Description,
				DefaultValue:  defaultValue,
				AllowedValues: allowedValues,
				Note:          opts.Usage.Note,
				DocsLink:      opts.Usage.DocsLink,
			},
			Hidden: opts.Hidden,
		},
	}
}

// StringFlag is a string flag
type StringFlag struct {
	Meta
	Value        *string
	DefaultValue string
}

// Register registers the string flag with the provided flag set
func (f StringFlag) Register(fs *pflag.FlagSet) {
	if f.Shorthand == "" {
		fs.StringVar(f.Value, f.Name, f.DefaultValue, f.Usage.String())
	} else {
		fs.StringVarP(f.Value, f.Name, f.Shorthand, f.DefaultValue, f.Usage.String())
	}

	registerFlag(fs, f.Meta)
}

// StringArrayFlag is a string array flag
type StringArrayFlag struct {
	Meta
	Value        *[]string
	DefaultValue []string
}

// Register registers the string array flag with the provided flag set
func (f StringArrayFlag) Register(fs *pflag.FlagSet) {
	if f.Shorthand == "" {
		fs.StringArrayVar(f.Value, f.Name, f.DefaultValue, f.Usage.String())
	} else {
		fs.StringArrayVarP(f.Value, f.Name, f.Shorthand, f.DefaultValue, f.Usage.String())
	}

	registerFlag(fs, f.Meta)
}

// StringSliceFlag is a string slice flag
type StringSliceFlag struct {
	Meta
	Value        *[]string
	DefaultValue []string
}

// Register registers the string slice flag with the provided flag set
func (f StringSliceFlag) Register(fs *pflag.FlagSet) {
	if f.Shorthand == "" {
		fs.StringSliceVar(f.Value, f.Name, f.DefaultValue, f.Usage.String())
	} else {
		fs.StringSliceVarP(f.Value, f.Name, f.Shorthand, f.DefaultValue, f.Usage.String())
	}

	registerFlag(fs, f.Meta)
}

func registerFlag(fs *pflag.FlagSet, f Meta) {
	if f.Deprecate != "" {
		MarkDeprecated(fs, f.Name, f.Deprecate)
	} else if f.Hidden {
		MarkHidden(fs, f.Name)
	}
}

// MarkHidden marks the specified flag as hidden from the provided flag set
// TODO(REALMC-8369): this method should go away if/when we can get
// golangci-lint to play nicely with errcheck and our exclude .errcheck file
// For now, we use this to isolate and minimize the nolint directives in this repo
func MarkHidden(fs *pflag.FlagSet, name string) {
	fs.MarkHidden(name) //nolint: errcheck
}

// MarkDeprecated marks the specified flag as deprecated from the provided flag set
// TODO(REALMC-8369): this method should go away if/when we can get
// golangci-lint to play nicely with errcheck and our exclude .errcheck file
// For now, we use this to isolate and minimize the nolint directives in this repo
func MarkDeprecated(fs *pflag.FlagSet, name, message string) {
	fs.MarkDeprecated(name, message) //nolint: errcheck
}
