package app

type errProjectExists struct{}

func (err errProjectExists) Error() string { return "a project already exists" }

func (err errProjectExists) DisableUsage() struct{} { return struct{}{} }
