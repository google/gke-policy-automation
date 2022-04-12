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
	"fmt"
	"io"
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
	if !reflect.DeepEqual(paApp.out.w, io.Discard) {
		t.Errorf("policyAutomationApp output is not io.Discard")
	}
}

func TestLoadCliConfig_file(t *testing.T) {
	testConfigPath := "./test-fixtures/test_config.yaml"
	cliConfig := &CliConfig{ConfigFile: testConfigPath}
	pa := PolicyAutomationApp{ctx: context.Background()}
	err := pa.LoadCliConfig(cliConfig, nil)
	if err != nil {
		t.Fatalf("err is not nil; want nil; err = %s", err)
	}

	data, err := os.ReadFile(testConfigPath)
	if err != nil {
		t.Fatalf("read test config file err is not nil; want nil; err = %s", err)
	}
	config := &Config{}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		t.Fatalf("unmarshal test config file err is not nil; want nil; err = %s", err)
	}
	if !reflect.DeepEqual(pa.config, config) {
		t.Errorf("policyAutomation config does not match test config")
	}
}

func TestLoadCliConfig_with_validation(t *testing.T) {
	validationErrMsg := "wrong validation"
	validateFnMock := func(config Config) error {
		return fmt.Errorf(validationErrMsg)
	}
	cliConfig := &CliConfig{}
	pa := PolicyAutomationApp{ctx: context.Background()}
	err := pa.LoadCliConfig(cliConfig, validateFnMock)
	if err == nil {
		t.Fatalf("expected error for loadCliConfig; got nil")
	}
	if err.Error() != validationErrMsg {
		t.Fatalf("error msg = %v; want %v", err.Error(), validationErrMsg)
	}
}

func TestLoadConfig(t *testing.T) {
	config := &Config{
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

func TestNewConfigFromCli_base(t *testing.T) {
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
	if len(config.Policies) != 1 {
		t.Fatalf("len(policies) = %v; want %v", len(config.Policies), 1)
	}
	policySrc := config.Policies[0]
	if policySrc.LocalDirectory != input.LocalDirectory {
		t.Errorf("policy localDirectory = %v; want %v", policySrc.LocalDirectory, input.LocalDirectory)
	}
}

func TestNewConfigFromCli_gitPolicySrc(t *testing.T) {
	input := &CliConfig{
		GitRepository: "https://github.com/test/test",
		GitBranch:     "main",
		GitDirectory:  "policies",
	}
	config := newConfigFromCli(input)
	if len(config.Policies) != 1 {
		t.Fatalf("len(policies) = %v; want %v", len(config.Policies), 1)
	}
	policySrc := config.Policies[0]
	if policySrc.GitRepository != input.GitRepository {
		t.Errorf("policy gitRepository = %v; want %v", policySrc.LocalDirectory, input.GitRepository)
	}
	if policySrc.GitBranch != input.GitBranch {
		t.Errorf("policy gitBranch = %v; want %v", policySrc.GitBranch, input.GitBranch)
	}
	if policySrc.GitDirectory != input.GitDirectory {
		t.Errorf("policy gitDirectory = %v; want %v", policySrc.GitDirectory, input.GitDirectory)
	}
}

func TestNewConfigFromCli_defaultPolicySrc(t *testing.T) {
	input := &CliConfig{}
	config := newConfigFromCli(input)
	if len(config.Policies) != 1 {
		t.Fatalf("len(policies) = %v; want %v", len(config.Policies), 1)
	}
	policySrc := config.Policies[0]
	if policySrc.GitRepository != DefaultGitRepository {
		t.Errorf("policy gitRepository = %v; want %v", policySrc.GitRepository, DefaultGitRepository)
	}
	if policySrc.GitBranch != DefaultGitBranch {
		t.Errorf("policy gitBranch = %v; want %v", policySrc.GitBranch, DefaultGitBranch)
	}
	if policySrc.GitDirectory != DefaultGitPolicyDir {
		t.Errorf("policy gitDirectory = %v; want %v", policySrc.GitDirectory, DefaultGitPolicyDir)
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
