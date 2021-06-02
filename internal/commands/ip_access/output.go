package ip_access

const (
	headerIP      = "IP"
	headerComment = "Comment"
)

func tableHeaders(additionalHeaders ...string) []string {
	return append([]string{headerIP, headerComment}, additionalHeaders...)
}
