package flags

// OptionalString is a value that can either be set with a user-defined value or unset and have a default value
type OptionalString struct {
	IsSet        bool
	Value        string
	DefaultValue string
}

// String returns the string representation of an optional string
func (o OptionalString) String() string {
	if o.IsSet {
		return o.Value
	}
	return o.DefaultValue
}

// Set determines how to set the value of an optional string
func (o *OptionalString) Set(s string) error {
	if s == "" {
		o.Value = o.DefaultValue
	} else {
		o.IsSet = true
		o.Value = s
	}

	return nil
}

func (o OptionalString) Type() string {
	return "OptionalString"
}


