package app

import (
	"fmt"
	"testing"
)

func TestReadConfig(t *testing.T) {
	filePath := "/some/test/path/file.txt"
	silent := true
	credsFile := "/path/to/creds.json"
	cluster1Name := "clusterOne"
	cluster1Location := "clusterOneLocation"
	cluster1Project := "clusterOneProject"
	cluster2Id := "projects/testProject/locations/europe-central2/clusters/clusterTwo"
	policy1Directory := "/my/test/policies"
	policy2Repository := "https://github.com/test/test"
	policy2Branch := "test"
	policy2Directory := "policies"
	fileData := fmt.Sprintf("silent: %t\n"+
		"credentialsFile: %s\n"+
		"clusters:\n"+
		"- name: %s\n"+
		"  location: %s\n"+
		"  project: %s\n"+
		"- id: %s\n"+
		"policies:\n"+
		"- local: %s\n"+
		"- repository: %s\n"+
		"  branch: %s\n"+
		"  directory: %s\n",
		silent, credsFile,
		cluster1Name, cluster1Location, cluster1Project, cluster2Id,
		policy1Directory, policy2Repository, policy2Branch, policy2Directory,
	)
	readFn := func(path string) ([]byte, error) {
		if path != filePath {
			t.Fatalf("file path = %v; want %v", path, filePath)
		}
		return []byte(fileData), nil
	}

	config, err := ReadConfig(filePath, readFn)
	if err != nil {
		t.Fatalf("got error want nil")
	}
	if config.SilentMode != silent {
		t.Errorf("config silent = %v; want %v", config.SilentMode, silent)
	}
	if config.CredentialsFile != credsFile {
		t.Errorf("config credentialsFile = %v; want %v", config.CredentialsFile, credsFile)
	}
	if len(config.Clusters) < 2 {
		t.Fatalf("config cluster length = %v; want %v", len(config.Clusters), 2)
	}
	if config.Clusters[0].Name != cluster1Name {
		t.Errorf("config cluster[0] name = %v; want %v", config.Clusters[0].Name, cluster1Name)
	}
	if config.Clusters[0].Location != cluster1Location {
		t.Errorf("config cluster[0] location = %v; want %v", config.Clusters[0].Location, cluster1Location)
	}
	if config.Clusters[0].Project != cluster1Project {
		t.Errorf("config cluster[0] project = %v; want %v", config.Clusters[0].Project, cluster1Project)
	}
	if config.Clusters[1].ID != cluster2Id {
		t.Errorf("config cluster[1] id = %v; want %v", config.Clusters[1].ID, cluster2Id)
	}
	if len(config.Policies) < 2 {
		t.Fatalf("config policies length = %v; want %v", len(config.Policies), 2)
	}
	if config.Policies[0].LocalDirectory != policy1Directory {
		t.Errorf("config policies[0] local = %v; want %v", config.Policies[0].LocalDirectory, policy1Directory)
	}
	if config.Policies[1].GitRepository != policy2Repository {
		t.Errorf("config policies[1] repository = %v; want %v", config.Policies[1].GitRepository, policy2Repository)
	}
	if config.Policies[1].GitBranch != policy2Branch {
		t.Errorf("config policies[1] gitBranch = %v; want %v", config.Policies[1].GitBranch, policy2Branch)
	}
	if config.Policies[1].GitDirectory != policy2Directory {
		t.Errorf("config policies[1] gitDirectory = %v; want %v", config.Policies[1].GitDirectory, policy2Directory)
	}
}
