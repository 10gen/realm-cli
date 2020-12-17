package initialize

type errProjectExists struct{}

func (err errProjectExists) Error() string { return "a project already exists" }

func (err errProjectExists) PrintUsage() bool { return false }
