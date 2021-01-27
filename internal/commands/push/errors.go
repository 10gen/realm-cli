package push

type errProjectNotFound struct {
}

func (err errProjectNotFound) Error() string {
	return "must specify --app-dir or run command from inside a Realm app directory"
}

func (err errProjectNotFound) DisableUsage() struct{} { return struct{}{} }
