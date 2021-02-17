package realm

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"

	"github.com/10gen/realm-cli/internal/utils/api"
)

const (
	exportPathPattern = appPathPattern + "/export"

	exportQueryForSourceControl = "source_control"
	exportQueryIsTemplated      = "template"
	exportQueryVersion          = "version"

	mediaParamFilename = "filename"

	trueVal = "true"
)

// ExportRequest is a Realm application export request
type ExportRequest struct {
	ConfigVersion AppConfigVersion
	IsTemplated   bool
}

func (c *client) Export(groupID, appID string, req ExportRequest) (string, *zip.Reader, error) {
	options := api.RequestOptions{Query: map[string]string{
		exportQueryVersion: DefaultAppConfigVersion.String(),
	}}

	if req.ConfigVersion != AppConfigVersionZero {
		options.Query[exportQueryVersion] = req.ConfigVersion.String()
	}
	if req.IsTemplated {
		options.Query[exportQueryIsTemplated] = trueVal
	} else {
		options.Query[exportQueryForSourceControl] = trueVal
	}

	res, resErr := c.do(http.MethodGet, fmt.Sprintf(exportPathPattern, groupID, appID), options)
	if resErr != nil {
		return "", nil, resErr
	}
	if res.StatusCode != http.StatusOK {
		return "", nil, api.ErrUnexpectedStatusCode{"export", res.StatusCode}
	}

	_, mediaParams, mediaErr := mime.ParseMediaType(res.Header.Get(api.HeaderContentDisposition))
	if mediaErr != nil {
		return "", nil, mediaErr
	}

	filename := mediaParams[mediaParamFilename]
	if filename == "" {
		return "", nil, errors.New("export response is missing filename")
	}

	defer res.Body.Close()
	body, bodyErr := ioutil.ReadAll(res.Body)
	if bodyErr != nil {
		return "", nil, bodyErr
	}

	zipPkg, zipErr := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if zipErr != nil {
		return "", nil, zipErr
	}

	return filename, zipPkg, nil
}
