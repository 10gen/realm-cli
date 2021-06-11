package accesslist

const (
	flagIP            = "ip"
	flagIPUsageCreate = "the allowed IP address"
	flagIPUsageUpdate = "the current allowed IP address"
	flagIPUsageDelete = "delete the allowed IP address"

	flagNewIP            = "new-ip"
	flagNewIPUsageUpdate = "the new allowed IP address"

	flagComment            = "comment"
	flagCommentUsageCreate = "the comment of the allowd IP address"
	flagCommentUsageUpdate = "the new comment of the allowed IP"

	flagUseCurrent            = "use-current"
	flagUseCurrentUsageCreate = "use current IP address"

	flagAllowAll            = "allow-all"
	flagAllowAllUsageCreate = "allow access to Realm app from everywhere"
)
