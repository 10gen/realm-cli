package pull

type errProjectNotFound struct {
}

func (err errProjectNotFound) Error() string {
	return "must specify --remote or run command from inside a Realm app directory"
}

func (err errProjectNotFound) DisableUsage() struct{} { return struct{}{} }
