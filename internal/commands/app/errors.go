package app

type errProjectExists struct {
	details string
}

func (err errProjectExists) Error() string {
	if err.details != "" {
		return err.details
	}
	return "a project already exists"
}

func (err errProjectExists) DisableUsage() struct{} { return struct{}{} }
