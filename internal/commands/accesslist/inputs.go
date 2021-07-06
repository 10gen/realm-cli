package accesslist

// Flag names and usages across the accessList commands
const (
	flagIP            = "ip"
	flagIPUsageCreate = "Specify the IP address or CIDR block that you would like to add"

	flagComment            = "comment"
	flagCommentUsageCreate = "Add a comment to the IP address or CIDR block that is being added to the Access List (Note: This action is optional)"

	flagUseCurrent            = "use-current"
	flagUseCurrentUsageCreate = "Add your current IP address to your Access List"

	flagAllowAll            = "allow-all"
	flagAllowAllUsageCreate = "Allows all IP addresses to access your Realm app (i.e. “0.0.0.0/0” will be added as an entry in your Access List)"
)
