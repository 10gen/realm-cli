package ipaccess

const (
	headerIP      = "IP Address"
	headerComment = "Comment"
)

func tableHeaders(additionalHeaders ...string) []string {
	return append([]string{headerIP, headerComment}, additionalHeaders...)
}
