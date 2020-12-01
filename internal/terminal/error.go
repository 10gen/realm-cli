package terminal

const (
	logFieldErr = "err"
)

var (
	errorMessageFields = []string{logFieldErr}
)

type errorMessage struct {
	error
}

func (e errorMessage) Message() (string, error) {
	return e.Error(), nil
}

func (e errorMessage) Payload() ([]string, map[string]interface{}, error) {
	return errorMessageFields, map[string]interface{}{
		logFieldErr: e.Error(),
	}, nil
}
