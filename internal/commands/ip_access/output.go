package ip_access

import "github.com/10gen/realm-cli/internal/cloud/realm"

const (
	headerIP      = "IP"
	headerComment = "Comment"
)

type allowedIPOutputs []allowedIPOutput

type allowedIPOutput struct {
	allowedIP realm.AllowedIP
	err       error
}
