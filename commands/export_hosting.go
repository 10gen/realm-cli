package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"sync"

	"github.com/10gen/realm-cli/api"
	"github.com/10gen/realm-cli/hosting"
	"github.com/10gen/realm-cli/models"
	"github.com/10gen/realm-cli/utils"
)

func getAssetAtURL(url string) (io.ReadCloser, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("downloading asset (url: %s) failed: response status code was %d", url, resp.StatusCode)
	}
	return resp.Body, nil
}

func exportStaticHostingAssets(realmClient api.RealmClient, ec *ExportCommand, appPath string, app *models.App) error {
	assetMetadatas, err := realmClient.ListAssetsForAppID(app.GroupID, app.ID)
	if err != nil {
		return err
	}

	assetDescriptions := hosting.AssetMetadataToAssetDescriptions(assetMetadatas)
	assetDescriptionsData, err := json.Marshal(assetDescriptions)
	if err != nil {
		return err
	}
	assetDescriptionsReader := bytes.NewReader(assetDescriptionsData)
	err = ec.writeFileToDirectory(path.Join(appPath, utils.HostingAttributes), assetDescriptionsReader)
	if err != nil {
		return fmt.Errorf("failed to write static hosting asset attributes file at: %s", path.Join(appPath, utils.HostingAttributes))
	}

	// Variables for the parallelization below
	var wg sync.WaitGroup
	jobs := make(chan hosting.AssetMetadata)
	errs := make(chan error)
	errorsHandlerDone := make(chan struct{})

	var errors []error
	// function for the error checker to run
	errChecker := func(errs <-chan error, errsDone chan<- struct{}) {
		for err := range errs {
			errors = append(errors, err)
		}
		errsDone <- struct{}{}
	}

	// run the error handler
	go errChecker(errs, errorsHandlerDone)

	// Spawn the workers
	for n := 0; n < numWorkers; n++ {
		wg.Add(1)
		go assetDownloadWorker(jobs, &wg, errs, ec, appPath)
	}

	// Pass in the information
	for _, amd := range assetMetadatas {
		jobs <- amd
	}

	close(jobs)
	wg.Wait()
	close(errs)
	<-errorsHandlerDone

	if len(errors) > 0 {
		return fmt.Errorf("exporting hosted assets failed, %d downloads were unsuccessful with error: %s", len(errors), errors[0])
	}

	return nil
}

// function for workers to run
func assetDownloadWorker(jobs <-chan hosting.AssetMetadata, wg *sync.WaitGroup, errs chan<- error, ec *ExportCommand, appPath string) {
	defer wg.Done()

	for job := range jobs {
		if job.IsDir() {
			assetDir := path.Join(appPath, utils.HostingFilesDirectory, job.FilePath)
			if mkdirErr := os.MkdirAll(assetDir, os.ModePerm); mkdirErr != nil {
				errs <- fmt.Errorf("failed to create directory %q: %s", assetDir, mkdirErr)
			}
			continue
		}

		// Go get the asset at the given URL
		reader, err := ec.getAssetAtURL(job.URL)
		if err != nil {
			errs <- err
			continue
		}

		// Now store the asset at the proper filePath
		err = ec.writeFileToDirectory(path.Join(appPath, utils.HostingFilesDirectory, job.FilePath), reader)
		reader.Close()
		if err != nil {
			errs <- err
			continue
		}
	}
}
