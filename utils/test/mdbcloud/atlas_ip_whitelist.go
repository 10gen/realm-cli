package mdbcloud

// AtlasIPWhitelistEntry represents an Atlas Group IP Whitelist entry
type AtlasIPWhitelistEntry struct {
	CIDRBlock string `json:"cidrBlock"`
	Comment   string `json:"comment"`
}

// AtlasIPWhitelistGetResponse represents the response to a Atlas Group IP Whitelist get request
type AtlasIPWhitelistGetResponse struct {
	Results []AtlasIPWhitelistEntry `json:"results"`
}
