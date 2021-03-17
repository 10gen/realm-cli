package app

type errProjectExists struct {
	path string
}

func (err errProjectExists) Error() string {
	var suffix string
	if err.path != "" {
		suffix = " at " + err.path
	}
	return "a project already exists" + suffix
}

func (err errProjectExists) DisableUsage() struct{} { return struct{}{} }
