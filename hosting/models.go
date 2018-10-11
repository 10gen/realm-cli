package hosting

import (
	"path"
	"sort"
	"strings"
	"time"

	"github.com/10gen/stitch-cli/utils"
)

// Valid attribute types names
var (
	AttributeContentType        = "Content-Type"
	AttributeContentDisposition = "Content-Disposition"
	AttributeContentLanguage    = "Content-Language"
	AttributeContentEncoding    = "Content-Encoding"
	AttributeCacheControl       = "Cache-Control"
)

// ValidAttributeNames stores the attribute names that Stitch static hosting supports
var ValidAttributeNames = map[string]bool{
	AttributeContentType:        true,
	AttributeContentDisposition: true,
	AttributeContentLanguage:    true,
	AttributeContentEncoding:    true,
	AttributeCacheControl:       true,
}

// AssetMetadata represents the metadata of a static hosted asset
type AssetMetadata struct {
	AppID        string           `json:"appId,omitempty"`
	FilePath     string           `json:"path"`
	FileHash     string           `json:"hash,omitempty"`
	FileSize     int64            `json:"size,omitempty"`
	Attrs        []AssetAttribute `json:"attrs"`
	AttrsHash    string           `json:"attrs_hash,omitempty"`
	LastModified int64            `json:"last_modified,omitempty"`
	URL          string           `json:"url,omitempty"`
}

// IsDir is true if the asset represents a directory
func (amd *AssetMetadata) IsDir() bool {
	return strings.HasSuffix(amd.FilePath, "/")
}

// NewAssetMetadata is a constructor for AssetMetadata
func NewAssetMetadata(appID string, filePath string, fileHash string, fileSize int64, attrs []AssetAttribute) *AssetMetadata {
	return &AssetMetadata{
		AppID:        appID,
		FilePath:     filePath,
		FileHash:     fileHash,
		FileSize:     fileSize,
		Attrs:        attrs,
		LastModified: time.Now().Unix(),
	}
}

// AssetAttribute represents an attribute of a particular static hosting asset
type AssetAttribute struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type byNameValue []AssetAttribute

func (b byNameValue) Len() int {
	return len(b)
}
func (b byNameValue) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}
func (b byNameValue) Less(i, j int) bool {
	if b[i].Name == b[j].Name {
		return b[i].Value < b[j].Value
	}
	return b[i].Name < b[j].Name
}

// AssetAttributesEqual determines whether the []AssetAttribute are the same
func AssetAttributesEqual(a, b []AssetAttribute) bool {
	sortA := byNameValue(a)
	sortB := byNameValue(b)
	sort.Sort(&sortA)
	sort.Sort(&sortB)

	if len(a) != len(b) {
		return false
	}

	for i, assetAttrA := range a {
		assetAttrB := b[i]
		if assetAttrB.Name != assetAttrA.Name || assetAttrB.Value != assetAttrA.Value {
			return false
		}
	}

	return true
}

// AssetDescription is the struct that contains the metadata we store for the CLI
type AssetDescription struct {
	FilePath string           `json:"path"`
	Attrs    []AssetAttribute `json:"attrs"`
}

// AssetMetadatasToAssetDescriptions takes AssetMetadatas and outputs the slice of AssetDescriptions
// that should be written into the metadata file
func AssetMetadatasToAssetDescriptions(assetMetadata []AssetMetadata) []AssetDescription {
	assetDescriptions := make([]AssetDescription, 0, len(assetMetadata))
	for _, amd := range assetMetadata {

		// If there are no attributes for the asset, we dont need to add it to the assetDescription file
		if len(amd.Attrs) == 0 {
			continue
		}

		// If the file's only attribute is "Content-Type" and the default type of its file extension
		// matches what appears in our default file type mappings then do not write any metadata entry for the file.
		if len(amd.Attrs) == 1 && amd.Attrs[0].Name == AttributeContentType {
			if extension := path.Ext(amd.FilePath); extension != "" {
				if ctype, found := utils.GetContentTypeByExtension(extension[1:]); found && ctype == amd.Attrs[0].Value {
					continue
				}
			}
		}

		// Save the values of the headers (Content-Type, Content-Disposition, Content-Language, Content-Encoding, Cache-Control)
		var assetAttributes []AssetAttribute
		for _, attribute := range amd.Attrs {
			if ValidAttributeNames[attribute.Name] {
				assetAttributes = append(assetAttributes, attribute)
			}
		}
		assetDescriptions = append(assetDescriptions, AssetDescription{FilePath: amd.FilePath, Attrs: assetAttributes})
	}
	return assetDescriptions
}

// ModifiedAssetMetadata represents a description of changes to assetMetadata
type ModifiedAssetMetadata struct {
	AssetMetadata AssetMetadata
	BodyModified  bool
	AttrModified  bool
}

// GetModifiedAssetMetadata returns a ModifiedAssetMetadata created from the
// diff between local and remote
func GetModifiedAssetMetadata(local, remote AssetMetadata) ModifiedAssetMetadata {
	bodyModified := local.FileHash != remote.FileHash
	attrModified := !AssetAttributesEqual(local.Attrs, remote.Attrs)

	return ModifiedAssetMetadata{
		local,
		bodyModified,
		attrModified,
	}
}

// AssetMetadataDiffs represents a set of
//locally deleted, locally added, and locally modified AssetMetadata
type AssetMetadataDiffs struct {
	AddedLocally    []AssetMetadata
	DeletedLocally  []AssetMetadata
	ModifiedLocally []ModifiedAssetMetadata
}

// NewAssetMetadataDiffs is a constructor for AssetMetadataDiffs
func NewAssetMetadataDiffs(added, deleted []AssetMetadata, modified []ModifiedAssetMetadata) *AssetMetadataDiffs {
	return &AssetMetadataDiffs{
		DeletedLocally:  deleted,
		AddedLocally:    added,
		ModifiedLocally: modified,
	}
}
