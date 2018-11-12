package commands

import (
	"encoding/json"
	"io/ioutil"
	"os/exec"
	"strings"
	"testing"

	"github.com/10gen/stitch-cli/api/mdbcloud"
	u "github.com/10gen/stitch-cli/utils/test"
	gc "github.com/smartystreets/goconvey/convey"
)

func TestCloudCommands(t *testing.T) {
	u.SkipUnlessMongoDBCloudRunning(t)
	cloudEnv := u.ENV()

	// test login
	loginArgs := []string{
		"run",
		"../main.go",
		"login",
		"--config-path",
		"../cli_conf",
		"--base-url",
		cloudEnv.StitchServerBaseURL,
		"--username",
		cloudEnv.Username,
		"--api-key",
		cloudEnv.APIKey,
	}

	out, err := exec.Command("go", loginArgs...).Output()
	t.Logf("output from login command: %s", string(out))
	u.So(t, err, gc.ShouldBeNil)

	out, err = exec.Command("ls", "../cli_conf").Output()
	t.Logf("output from ls command: %s", string(out))
	u.So(t, err, gc.ShouldBeNil)

	// test import
	importArgs := []string{
		"run",
		"../main.go",
		"import",
		"--config-path",
		"../cli_conf",
		"--base-url",
		cloudEnv.StitchServerBaseURL,
		"--path",
		"../testdata/simple_app_with_cluster",
		"--project-id",
		cloudEnv.GroupID,
		"--yes",
	}
	out, err = exec.Command("go", importArgs...).Output()
	t.Logf("Output from import command: '%s'", string(out))
	u.So(t, err, gc.ShouldBeNil)

	importOut := string(out)
	appID := importOut[strings.Index(importOut, "'simple-app-")+1 : len(importOut)-2]

	atlasClient := mdbcloud.NewClient(cloudEnv.CloudAPIBaseURL).
		WithAuth(cloudEnv.AdminUsername, cloudEnv.AdminAPIKey)

	defer atlasClient.DeleteDatabaseUser(cloudEnv.GroupID, "mongodb-stitch-"+appID)

	// test export
	exportArgs := []string{
		"run",
		"../main.go",
		"export",
		"--config-path",
		"../cli_conf",
		"--base-url",
		cloudEnv.StitchServerBaseURL,
		"--app-id",
		appID,
		"-o",
		"../exported_app",
		"--yes",
	}
	err = exec.Command("go", exportArgs...).Run()
	u.So(t, err, gc.ShouldBeNil)

	out, _ = exec.Command("cat", "../exported_app/stitch.json").Output()
	u.So(t, string(out), gc.ShouldContainSubstring, "\"app_id\":")
	u.So(t, string(out), gc.ShouldContainSubstring, "\"name\":")
	diffFiles(
		t,
		"../testdata/simple_app_with_cluster/stitch.json",
		"../exported_app/stitch.json",
	)

	out, _ = exec.Command("cat", "../exported_app/services/mongodb-atlas/config.json").Output()
	u.So(t, string(out), gc.ShouldContainSubstring, "\"id\":")
	u.So(t, string(out), gc.ShouldContainSubstring, "\"name\":")
	diffFiles(
		t,
		"../testdata/simple_app_with_cluster/services/mongodb-atlas/config.json",
		"../exported_app/services/mongodb-atlas/config.json",
	)

	// test export template
	exportTemplateArgs := []string{
		"run",
		"../main.go",
		"export",
		"--config-path",
		"../cli_conf",
		"--base-url",
		cloudEnv.StitchServerBaseURL,
		"--app-id",
		appID,
		"-o",
		"../exported_tmpl",
		"--yes",
		"--as-template",
	}
	err = exec.Command("go", exportTemplateArgs...).Run()
	u.So(t, err, gc.ShouldBeNil)

	out, _ = exec.Command("cat", "../exported_tmpl/stitch.json").Output()
	u.So(t, string(out), gc.ShouldNotContainSubstring, "\"app_id\":")
	u.So(t, string(out), gc.ShouldContainSubstring, "\"name\":")

	diffFiles(
		t,
		"../testdata/template_app_with_cluster/stitch.json",
		"../exported_tmpl/stitch.json",
	)

	out, _ = exec.Command("cat", "../exported_tmpl/services/mongodb-atlas/config.json").Output()
	u.So(t, string(out), gc.ShouldNotContainSubstring, "\"id\":")
	u.So(t, string(out), gc.ShouldContainSubstring, "\"name\":")

	diffFiles(
		t,
		"../testdata/template_app_with_cluster/services/mongodb-atlas/config.json",
		"../exported_tmpl/services/mongodb-atlas/config.json",
	)
}

func diffFiles(t *testing.T, expectedFilePath, actualFilePath string) {
	var expectedConfig map[string]interface{}
	expectedData, err := ioutil.ReadFile(expectedFilePath)
	u.So(t, err, gc.ShouldBeNil)
	u.So(t, json.Unmarshal(expectedData, &expectedConfig), gc.ShouldBeNil)

	var actualConfig map[string]interface{}
	actualData, err := ioutil.ReadFile(actualFilePath)
	u.So(t, err, gc.ShouldBeNil)
	u.So(t, json.Unmarshal(actualData, &actualConfig), gc.ShouldBeNil)

	u.So(t, actualConfig, gc.ShouldResemble, expectedConfig)
}
