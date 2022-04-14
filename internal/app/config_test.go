// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

func TestValidateClusterReviewConfig(t *testing.T) {
	config := Config{
		Clusters: []ConfigCluster{
			{ID: "some/cluster/id"},
			{Name: "cluster", Location: "region", Project: "project"},
		},
		Policies: []ConfigPolicy{
			{LocalDirectory: "./directory"},
			{GitRepository: "repo", GitBranch: "main", GitDirectory: "./dir"},
		},
	}
	if err := ValidateClusterReviewConfig(config); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidateClusterReviewConfig_negative(t *testing.T) {
	badConfigs := []Config{
		{
			Clusters: []ConfigCluster{
				{ID: "some/cluster/id", Name: "someName"},
				{Name: "someName"},
				{Name: "someName", Location: "location"},
				{Project: "project"},
			},
			Policies: []ConfigPolicy{
				{LocalDirectory: "./directory"},
				{GitRepository: "repo"},
				{GitRepository: "repo", GitBranch: "main"},
				{GitDirectory: "dir"},
				{LocalDirectory: "./directory", GitRepository: "somerepo"},
			},
		},
		{
			Clusters: []ConfigCluster{{ID: "Some/cluster"}},
		},
		{},
	}

	for i, config := range badConfigs {
		if err := ValidateClusterReviewConfig(config); err == nil {
			t.Errorf("expected error on invalid cluster config [%d]", i)
		}
	}
}

func TestValidatePolicyCheckConfig(t *testing.T) {
	config := Config{
		Policies: []ConfigPolicy{
			{LocalDirectory: "./directory"},
			{GitRepository: "repo", GitBranch: "main", GitDirectory: "./dir"},
		},
	}
	if err := ValidatePolicyCheckConfig(config); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidatePolicyCheckConfig_negative(t *testing.T) {
	badConfigs := []Config{
		{
			Policies: []ConfigPolicy{
				{LocalDirectory: "./directory"},
				{GitRepository: "repo"},
				{GitRepository: "repo", GitBranch: "main"},
				{GitDirectory: "dir"},
				{LocalDirectory: "./directory", GitRepository: "somerepo"},
			},
		},
		{},
	}

	for i, config := range badConfigs {
		if err := ValidatePolicyCheckConfig(config); err == nil {
			t.Errorf("expected error on invalid cluster config [%d]", i)
		}
	}
}
