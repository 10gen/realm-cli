package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/10gen/stitch-cli/api"
	"github.com/10gen/stitch-cli/hosting"
	"github.com/10gen/stitch-cli/utils"

	"github.com/mitchellh/cli"
	"github.com/mitchellh/go-homedir"
)

// checkErrs builds a list of errors from the error channel errChan and logs them
func checkErrs(errChan <-chan error, errDoneChan chan<- struct{}, ui cli.Ui, errors *[]error) {
	for err := range errChan {
		*errors = append(*errors, err)
		ui.Error(err.Error())
	}
	errDoneChan <- struct{}{}
}

// ImportHosting will push local Stitch hosting assets to the server
func ImportHosting(groupID, appID, rootDir string, assetMetadataDiffs *hosting.AssetMetadataDiffs, resetCache bool, client api.StitchClient, ui cli.Ui) error {
	// build a channel of hosting operations
	var opWG sync.WaitGroup
	opChan := make(chan hostingOp)
	errChan := make(chan error)
	errDoneChan := make(chan struct{})

	var errors []error
	go checkErrs(errChan, errDoneChan, ui, &errors)

	// create workers
	for n := 0; n < numWorkers; n++ {
		opWG.Add(1)
		go hostingOpHandler(opChan, &opWG, errChan)
	}

	baseOp := baseHostingOp{groupID, appID, rootDir, client}
	// create hosting Ops to be handled
	for _, added := range assetMetadataDiffs.AddedLocally {
		opChan <- &addOp{baseOp, added}
	}

	for _, deleted := range assetMetadataDiffs.DeletedLocally {
		opChan <- &deleteOp{baseOp, deleted}
	}

	for _, modified := range assetMetadataDiffs.ModifiedLocally {
		opChan <- &modifyOp{baseOp, modified}
	}

	close(opChan)
	opWG.Wait()
	close(errChan)
	<-errDoneChan

	if len(errors) > 0 {
		return fmt.Errorf("%v error(s) occurred while importing hosting assets", len(errors))
	}

	if resetCache {
		if err := client.InvalidateCache(groupID, appID, "/*"); err != nil {
			return err
		}
	}

	return nil
}

func hostingOpHandler(opChan <-chan hostingOp, opWG *sync.WaitGroup, errChan chan<- error) {
	defer opWG.Done()

	for op := range opChan {
		if doErr := op.Do(); doErr != nil {
			errChan <- doErr
			continue
		}
	}
}

type baseHostingOp struct {
	groupID string
	appID   string
	rootDir string
	client  api.StitchClient
}

// hostingOp represents an import operation done with hosting assets
type hostingOp interface {
	Do() error
}

type addOp struct {
	baseHostingOp
	assetMetadata hosting.AssetMetadata
}

// Do performs an add operation
func (op *addOp) Do() error {
	return doUpload(op.groupID, op.appID, op.rootDir, op.client, op.assetMetadata)
}

type deleteOp struct {
	baseHostingOp
	assetMetadata hosting.AssetMetadata
}

// DoRequest performs a delete operation
func (op *deleteOp) Do() error {
	fp := op.assetMetadata.FilePath
	if err := op.client.DeleteAsset(op.groupID, op.appID, fp); err != nil {
		return fmt.Errorf("deleting '%s' failed => %s", fp, err)
	}
	return nil
}

type modifyOp struct {
	baseHostingOp
	modifiedAssetMetadata hosting.ModifiedAssetMetadata
}

// DoRequest performs modify operation
func (op *modifyOp) Do() error {
	mAM := op.modifiedAssetMetadata
	// only the attributes were modified
	if mAM.AttrModified && !mAM.BodyModified {
		fp := op.modifiedAssetMetadata.AssetMetadata.FilePath
		if err :=
			op.client.SetAssetAttributes(
				op.groupID,
				op.appID,
				fp,
				mAM.AssetMetadata.Attrs...); err != nil {
			return fmt.Errorf("%s => %s", fp, err)
		}

		return nil
	}

	if uploadErr := doUpload(op.groupID, op.appID, op.rootDir, op.client, mAM.AssetMetadata); uploadErr != nil {
		return uploadErr
	}

	return nil
}

func doUpload(groupID, appID, rootDir string, client api.StitchClient, am hosting.AssetMetadata) error {
	errStrF := "uploading '%s' failed => %s"

	body, bodyErr := os.Open(filepath.Join(rootDir, am.FilePath))
	if bodyErr != nil {
		return fmt.Errorf(errStrF, am.FilePath, bodyErr)
	}
	defer body.Close()

	if uploadErr := client.UploadAsset(groupID, appID, am.FilePath, am.FileHash, am.FileSize, body, am.Attrs...); uploadErr != nil {
		return fmt.Errorf(errStrF, am.FilePath, uploadErr)
	}

	return nil
}

func getAssetCachePath(configPath string) (string, error) {
	cachePath, eErr := homedir.Expand(configPath)
	if eErr != nil {
		return "", eErr
	}

	if cachePath == "" {
		home, dirErr := homedir.Dir()
		if dirErr != nil {
			return "", dirErr
		}
		cachePath = filepath.Join(home, ".config", "stitch")
	} else {
		cachePath = filepath.Dir(cachePath)
	}

	return filepath.Join(cachePath, utils.HostingCacheFileName), nil
}
