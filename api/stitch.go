package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"path"

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

// AssetAttribute represents an attribute of a particular static hosting asset
type AssetAttribute struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// AssetDescription is the struct that contains the metadata we store for the CLI
type AssetDescription struct {
	FilePath string           `json:"path"`
	Attrs    []AssetAttribute `json:"attrs"`
}

// ValidAttributeNames stores the attribute names that Stitch static hosting supports
var ValidAttributeNames = map[string]bool{
	AttributeContentType:        true,
	AttributeContentDisposition: true,
	AttributeContentLanguage:    true,
	AttributeContentEncoding:    true,
	AttributeCacheControl:       true,
}

// ErrAppNotFound is used when an app cannot be found by client app ID
type ErrAppNotFound struct {
	ClientAppID string
}

func (eanf ErrAppNotFound) Error() string {
	return fmt.Sprintf("Unable to find app with ID: %q", eanf.ClientAppID)
}

// ErrStitchResponse represents a response from a Stitch API call
type ErrStitchResponse struct {
	data errStitchResponseData
}

// Error returns a stringified error message
func (esr ErrStitchResponse) Error() string {
	return fmt.Sprintf("error: %s", esr.data.Error)
}

// UnmarshalJSON unmarshals JSON data into an ErrStitchResponse
func (esr *ErrStitchResponse) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &esr.data)
}

type errStitchResponseData struct {
	Error string `json:"error"`
}

// UnmarshalStitchError unmarshals an *http.Response into an ErrStitchResponse. If the Body does not
// contain content it uses the provided Status
func UnmarshalStitchError(res *http.Response) error {
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(res.Body); err != nil {
		return err
	}

	str := buf.String()
	if str == "" {
		return ErrStitchResponse{
			data: errStitchResponseData{
				Error: res.Status,
			},
		}
	}

	var stitchResponse ErrStitchResponse
	if err := json.NewDecoder(&buf).Decode(&stitchResponse); err != nil {
		stitchResponse.data.Error = str
	}

	return stitchResponse
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
