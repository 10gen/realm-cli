package commands

import (
	"os/exec"
	"strings"
	"testing"

	u "github.com/10gen/stitch-cli/utils/test"
	"github.com/10gen/stitch-cli/utils/test/harness"
	gc "github.com/smartystreets/goconvey/convey"
)

func TestCloudCommands(t *testing.T) {
	u.SkipUnlessMongoDBCloudRunning(t)
	serverBaseURL := u.StitchServerBaseURL()

	// setup cloud
	cloudClient := harness.NewCloudPrivateAPIClient(t)
	err := cloudClient.RegisterUser()
	u.So(t, err, gc.ShouldBeNil)
	err = cloudClient.CreateGroup(harness.PlanTypeNDS)
	u.So(t, err, gc.ShouldBeNil)
	_, apiKey, err := cloudClient.CreateAPIKey()
	u.So(t, err, gc.ShouldBeNil)

	// test login
	loginArgs := []string{
		"run",
		"../main.go",
		"login",
		"--config-path",
		"../cli_conf",
		"--base-url",
		serverBaseURL,
		"--username",
		cloudClient.Username(),
		"--api-key",
		apiKey,
	}

	err = exec.Command("go", loginArgs...).Run()
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
		serverBaseURL,
		"--path",
		"../testdata/simple_app",
		"--project-id",
		cloudClient.GroupID(),
		"--yes",
	}
	out, err := exec.Command("go", importArgs...).Output()
	u.So(t, err, gc.ShouldBeNil)

	// test export
	importOut := string(out)
	appID := importOut[strings.Index(importOut, "'simple-app-")+1 : len(importOut)-2]
	exportArgs := []string{
		"run",
		"../main.go",
		"export",
		"--config-path",
		"../cli_conf",
		"--base-url",
		serverBaseURL,
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
	out, _ = exec.Command("diff", "../testdata/simple_app/stitch.json", "../exported_app/stitch.json").Output()
	u.So(t, out, gc.ShouldHaveLength, 0)

}
