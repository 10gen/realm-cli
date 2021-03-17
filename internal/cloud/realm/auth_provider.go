package realm

// AuthProvider is a Realm application auth provider
type AuthProvider struct {
	ID                 string                 `json:"id,omitempty"`
	Name               string                 `json:"name"`
	Type               string                 `json:"type"`
	Config             map[string]interface{} `json:"config,omitempty"`
	SecretConfig       map[string]interface{} `json:"secret_config,omitempty"`
	Disabled           bool                   `json:"disabled"`
	MetadataFields     []AuthMetdataField     `json:"metadata_fields,omitempty"`
	DomainRestrictions []string               `json:"domain_restrictions,omitempty"`
	RedirectURIs       []string               `json:"redirect_uris,omitempty"`
}

// AuthMetdataField is a metadata field used with Realm auth
type AuthMetdataField struct {
	Required  bool   `json:"required"`
	Name      string `json:"name"`
	FieldName string `json:"field_name,omitempty"`
}
