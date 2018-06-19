package commands

import (
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

	err := exec.Command("go", loginArgs...).Run()
	u.So(t, err, gc.ShouldBeNil)
	err = exec.Command("ls", "../cli_conf").Run()
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
	out, err := exec.Command("go", importArgs...).Output()
	u.So(t, err, gc.ShouldBeNil)

	importOut := string(out)
	appID := importOut[strings.Index(importOut, "'simple-app-")+1 : len(importOut)-2]

	atlasClient := mdbcloud.NewClient(cloudEnv.CloudAPIBaseURL).
		WithAuth(cloudEnv.Username, cloudEnv.APIKey)

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
	out, _ = exec.Command(
		"diff",
		"../testdata/simple_app_with_cluster/stitch.json",
		"../exported_app/stitch.json",
	).Output()
	u.So(t, out, gc.ShouldHaveLength, 0)

	out, _ = exec.Command("cat", "../exported_app/services/mongodb-atlas/config.json").Output()
	u.So(t, string(out), gc.ShouldContainSubstring, "\"id\":")
	out, _ = exec.Command(
		"diff",
		"../testdata/simple_app_with_cluster/services/mongodb-atlas/config.json",
		"../exported_app/services/mongodb-atlas/config.json",
	).Output()
	u.So(t, out, gc.ShouldHaveLength, 0)
}
