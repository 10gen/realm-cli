package commands

import (
	"os/exec"
	"strings"
	"testing"
	"time"

	u "github.com/10gen/stitch-cli/utils/test"
	"github.com/10gen/stitch-cli/utils/test/harness"
	"github.com/10gen/stitch-cli/utils/test/mdbcloud"
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
	groupID := cloudClient.GroupID()

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

	// setup atlas cluster
	atlasClient := mdbcloud.NewClient(u.MongoDBCloudPublicAPIBaseURL(), u.MongoDBCloudAtlasAPIBaseURL()).
		WithAuth(cloudClient.Username(), apiKey)
	testCluster := mdbcloud.CreateAtlasCluster{
		BackupEnabled: "false",
		AtlasCluster: mdbcloud.AtlasCluster{
			Name: "testCluster",
			ProviderSettings: mdbcloud.ProviderSettings{
				ProviderName: "AWS",
				RegionName:   "US_EAST_1",
				InstanceSize: "M10",
			},
		},
	}
	err = atlasClient.CreateAtlasCluster(groupID, testCluster)
	u.So(t, err, gc.ShouldBeNil)
	defer atlasClient.DeleteAtlasCluster(groupID, "testCluster")

	t.Logf("Waiting for cluster to deploy")
	var cluster *mdbcloud.AtlasCluster
	time.Sleep(5 * time.Minute)
	for i := 0; i < 30; i++ {
		cluster, err = atlasClient.AtlasCluster(groupID, "testCluster")
		u.So(t, err, gc.ShouldBeNil)

		t.Logf("Cluster status: %s", cluster.StateName)
		if cluster.StateName != "CREATING" {
			break
		}
		time.Sleep(30 * time.Second)
	}
	if cluster.StateName == "CREATING" {
		t.Fatal("atlas cluster did not deploy in time")
	}

	t.Logf("Cluster finished deploying")

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
		"../testdata/simple_app_with_cluster",
		"--project-id",
		groupID,
		"--yes",
	}
	out, err := exec.Command("go", importArgs...).Output()
	u.So(t, err, gc.ShouldBeNil)

	// check Atlas whitelist for Stitch entry
	whitelistResponse, err := atlasClient.AtlasIPWhitelistEntries(groupID)
	u.So(t, err, gc.ShouldBeNil)
	hasStitchIP := false
	for _, whitelistEntry := range whitelistResponse.Results {
		if strings.Contains(whitelistEntry.Comment, "For MongoDB Stitch; do not delete") {
			hasStitchIP = true
		}
	}
	u.So(t, hasStitchIP, gc.ShouldBeTrue)

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
