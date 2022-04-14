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
	"context"
	"os"
	"reflect"
	"testing"

	"gopkg.in/yaml.v2"
)

func TestNewPolicyAutomationApp(t *testing.T) {
	pa := NewPolicyAutomationApp()
	paApp, ok := pa.(*PolicyAutomationApp)
	if !ok {
		t.Fatalf("Result of NewPolicyAutomationApp is not *PolicyAutomationApp")
	}
	if paApp.ctx == nil {
		t.Fatalf("policyAutomationApp ctx is nil")
	}
	if paApp.out == nil {
		t.Fatalf("policyAutomationApp output is nil")
	}
}

func TestLoadCliConfig_file(t *testing.T) {
	testConfigPath := "./test-fixtures/test_config.yaml"
	cliConfig := &CliConfig{ConfigFile: testConfigPath}
	pa := PolicyAutomationApp{ctx: context.Background()}
	err := pa.LoadCliConfig(cliConfig)
	if err != nil {
		t.Fatalf("err is not nil; want nil; err = %s", err)
	}

	data, err := os.ReadFile(testConfigPath)
	if err != nil {
		t.Fatalf("read test config file err is not nil; want nil; err = %s", err)
	}
	config := &ConfigNg{}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		t.Fatalf("unmarshal test config file err is not nil; want nil; err = %s", err)
	}
	if !reflect.DeepEqual(pa.config, config) {
		t.Errorf("policyAutomation config does not match test config")
	}
}

func TestLoadConfig(t *testing.T) {
	config := &ConfigNg{
		CredentialsFile: "./test-fixtures/test_credentials.json",
	}
	pa := PolicyAutomationApp{ctx: context.Background()}
	err := pa.LoadConfig(config)
	if err != nil {
		t.Fatalf("err is not nil; want nil; err = %s", err)
	}
	if !reflect.DeepEqual(config, pa.config) {
		t.Errorf("pa.config is not same as input config")
	}
	if pa.gke == nil {
		t.Errorf("pa.gke is nil; want gke.GKEClient")
	}
	err = pa.Close()
	if err != nil {
		t.Errorf("err on close is not nil; want nil; err = %s", err)
	}
}

func TestNewConfigFromCli(t *testing.T) {
	input := &CliConfig{
		SilentMode:      true,
		CredentialsFile: "/path/to/creds.json",
		ClusterName:     "testCluster",
		ClusterLocation: "europe-central2",
		LocalDirectory:  "/path/to/policies",
		GitRepository:   "https://github.com/test/test",
		GitBranch:       "main",
		GitDirectory:    "policies",
	}
	config := newConfigFromCli(input)
	if config.SilentMode != input.SilentMode {
		t.Errorf("silentMode = %v; want %v", config.SilentMode, input.SilentMode)
	}
	if config.CredentialsFile != input.CredentialsFile {
		t.Errorf("credentialsFile = %v; want %v", config.CredentialsFile, input.CredentialsFile)
	}
	if len(config.Clusters) != 1 {
		t.Fatalf("len(clusters) = %v; want %v", len(config.Clusters), 1)
	}
	if config.Clusters[0].Name != input.ClusterName {
		t.Errorf("cluster[0] name = %v; want %v", config.Clusters[0].Name, input.ClusterName)
	}
	if config.Clusters[0].Location != input.ClusterLocation {
		t.Errorf("cluster[0] location = %v; want %v", config.Clusters[0].Location, input.ClusterLocation)
	}
	if config.Clusters[0].Project != input.ProjectName {
		t.Errorf("cluster[0] project = %v; want %v", config.Clusters[0].Project, input.ProjectName)
	}
	if len(config.Policies) != 2 {
		t.Fatalf("len(policies) = %v; want %v", len(config.Policies), 2)
	}
	if config.Policies[0].LocalDirectory != input.LocalDirectory {
		t.Errorf("policies[0] localDirectory = %v; want %v", config.Policies[0].LocalDirectory, input.LocalDirectory)
	}
	if config.Policies[1].GitRepository != input.GitRepository {
		t.Errorf("policies[1] gitRepository = %v; want %v", config.Policies[1].GitRepository, input.GitRepository)
	}
	if config.Policies[1].GitBranch != input.GitBranch {
		t.Errorf("policies[1] gitBranch = %v; want %v", config.Policies[1].GitBranch, input.GitBranch)
	}
	if config.Policies[1].GitDirectory != input.GitDirectory {
		t.Errorf("policies[1] gitDirectory = %v; want %v", config.Policies[1].GitDirectory, input.GitDirectory)
	}
}

func TestGetClusterName(t *testing.T) {
	input := []ConfigCluster{
		{ID: "projects/myproject/locations/europe-central2/clusters/testCluster"},
		{Name: "testClusterTwo", Location: "europe-east2", Project: "testProject"},
	}
	expected := []string{
		"projects/myproject/locations/europe-central2/clusters/testCluster",
		"projects/testProject/locations/europe-east2/clusters/testClusterTwo",
	}
	for i := range input {
		name, _ := getClusterName(input[i])
		if name != expected[i] {
			t.Errorf("clusterName = %v; want %v", name, expected[i])
		}
	}
}

func TestGetClusterName_negative(t *testing.T) {
	input := ConfigCluster{Name: "test", Location: "europe-east2"}
	_, err := getClusterName(input)
	if err == nil {
		t.Errorf("error is nil; want error")
	}
}

/*
func TestCreateReviewApp(t *testing.T) {
	clusterName := "testCluster"
	clusterLocation := "europe-warsaw2"
	projectName := "testProject"
	credsFile := "./creds"
	gitRepo := "https://github.com/user/repo"
	gitBranch := "my-branch"
	gitDirectory := "rego-remote"
	localDirectory := "rego-local"

	args := []string{"gke-review",
		"-c", clusterName, "-l", clusterLocation,
		"-p", projectName,
		"-creds", credsFile,
		"-git-policy-repo", gitRepo,
		"-git-policy-branch", gitBranch,
		"-git-policy-dir", gitDirectory,
		"-local-policy-dir", localDirectory,
		"-s",
	}
	reviewMock := func(c *Config) {
		if c.ClusterName != clusterName {
			t.Errorf("clusterName = %s; want %s", c.ClusterName, clusterName)
		}
		if c.ClusterLocation != clusterLocation {
			t.Errorf("clusterLocation = %s; want %s", c.ClusterLocation, clusterLocation)
		}
		if c.ProjectName != projectName {
			t.Errorf("projectName = %s; want %s", c.ProjectName, projectName)
		}
		if c.CredentialsFile != credsFile {
			t.Errorf("CredentialsFile = %s; want %s", c.CredentialsFile, credsFile)
		}
		if !c.SilentMode {
			t.Errorf("SilentMode = %v; want true", c.SilentMode)
		}
		if c.GitRepository != gitRepo {
			t.Errorf("GitRepository = %s; want %s", c.GitRepository, gitRepo)
		}
		if c.GitBranch != gitBranch {
			t.Errorf("GitBranch = %s; want %s", c.GitBranch, gitBranch)
		}
		if c.GitDirectory != gitDirectory {
			t.Errorf("GitDirectory = %s; want %s", c.GitDirectory, gitDirectory)
		}
		if c.LocalDirectory != localDirectory {
			t.Errorf("LocalDirectory = %s; want %s", c.LocalDirectory, localDirectory)
		}
	}
	err := CreateReviewApp(reviewMock).Run(args)
	if err != nil {
		t.Fatalf("error when running the review application: %v", err)
	}
}

func TestCreateReviewApp_Defaults(t *testing.T) {
	args := []string{"gke-review",
		"-c", "testCluster", "-l", "europe-warsaw2",
		"-p", "testProject"}

	reviewMock := func(c *Config) {
		if c.GitRepository != DefaultGitRepository {
			t.Errorf("GitRepository = %s; want %s", c.GitRepository, DefaultGitRepository)
		}
		if c.GitBranch != DefaultGitBranch {
			t.Errorf("GitBranch = %s; want %s", c.GitBranch, DefaultGitBranch)
		}
		if c.GitDirectory != DefaultGitPolicyDir {
			t.Errorf("GitDirectory = %s; want %s", c.GitDirectory, DefaultGitPolicyDir)
		}
	}
	err := CreateReviewApp(reviewMock).Run(args)
	if err != nil {
		t.Fatalf("error when running the review application: %v", err)
	}
}

func TestGetPolicySource(t *testing.T) {
	c := &Config{
		GitRepository: DefaultGitRepository,
		GitBranch:     DefaultGitBranch,
		GitDirectory:  DefaultGitPolicyDir,
	}
	src := getPolicySource(c)
	if _, ok := src.(*policy.GitPolicySource); !ok {
		t.Errorf("policySource is not *GitPolicySource; want not *GitPolicySource")
	}
}

func TestGetPolicySource_local(t *testing.T) {
	c := &Config{
		GitRepository:  DefaultGitRepository,
		GitBranch:      DefaultGitBranch,
		GitDirectory:   DefaultGitPolicyDir,
		LocalDirectory: "some-local-dir",
	}
	src := getPolicySource(c)
	if _, ok := src.(*policy.LocalPolicySource); !ok {
		t.Errorf("policySource is not *LocalPolicySource; want not *LocalPolicySource")
	}
}
*/
