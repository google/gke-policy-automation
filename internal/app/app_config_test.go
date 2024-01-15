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
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"
	"time"

	cfg "github.com/google/gke-policy-automation/internal/config"
	"gopkg.in/yaml.v3"
)

func TestLoadCliConfig_file(t *testing.T) {
	testConfigPath := "./test-fixtures/test_config.yaml"
	cliConfig := &CliConfig{ConfigFile: testConfigPath}
	pa := PolicyAutomationApp{ctx: context.Background()}
	err := pa.LoadCliConfig(cliConfig, nil, nil)
	if err != nil {
		t.Fatalf("err is not nil; want nil; err = %s", err)
	}

	data, err := os.ReadFile(testConfigPath)
	if err != nil {
		t.Fatalf("read test config file err is not nil; want nil; err = %s", err)
	}
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	config := &cfg.Config{}
	if err := decoder.Decode(&config); err != nil && err != io.EOF {
		t.Fatalf("unmarshal test config file err is not nil; want nil; err = %s", err)
	}
	if !reflect.DeepEqual(pa.config, config) {
		t.Errorf("policyAutomation config does not match test config")
	}
}

func TestLoadCliConfig_with_validation(t *testing.T) {
	validationErrMsg := "wrong validation"
	validateFnMock := func(config cfg.Config) error {
		return fmt.Errorf(validationErrMsg)
	}
	cliConfig := &CliConfig{}
	pa := PolicyAutomationApp{ctx: context.Background()}
	err := pa.LoadCliConfig(cliConfig, nil, validateFnMock)
	if err == nil {
		t.Fatalf("expected error for loadCliConfig; got nil")
	}
	if err.Error() != validationErrMsg {
		t.Fatalf("error msg = %v; want %v", err.Error(), validationErrMsg)
	}
}

func TestLoadCliConfig_checkDefaults(t *testing.T) {
	cliConfig := &CliConfig{
		CredentialsFile: "./test-fixtures/test_credentials.json",
		ClusterName:     "test",
		ClusterLocation: "europe-central2",
		ProjectName:     "my-project",
	}
	pa := PolicyAutomationApp{ctx: context.Background()}
	err := pa.LoadCliConfig(cliConfig, cfg.SetCheckConfigDefaults, nil)
	if err != nil {
		t.Fatalf("err is not nil; want nil; err = %s", err)
	}
	if len(pa.config.Policies) != 1 {
		t.Fatalf("len of config policies is %d; want %d", len(pa.config.Policies), 1)
	}
	policy := pa.config.Policies[0]
	defaultPolicy := cfg.ConfigPolicy{
		GitRepository: cfg.DefaultGitRepository,
		GitBranch:     cfg.DefaultGitBranch,
		GitDirectory:  cfg.DefaultGitPolicyDir,
	}
	if !reflect.DeepEqual(policy, defaultPolicy) {
		t.Error("config policy is not same as default policy")
	}
}

func TestLoadConfig(t *testing.T) {
	config := &cfg.Config{
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
	err = pa.Close()
	if err != nil {
		t.Errorf("err on close is not nil; want nil; err = %s", err)
	}
}

func TestNewConfigFromCli(t *testing.T) {
	input := &CliConfig{
		SilentMode:      true,
		JSONOutput:      true,
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
	if config.JSONOutput != input.JSONOutput {
		t.Errorf("jsonOutput = %v; want %v", config.JSONOutput, input.JSONOutput)
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

func TestAddDatetimePrefix(t *testing.T) {
	testDate := time.Date(1994, 7, 20, 5, 20, 0, 0, time.UTC)
	expectedResult := "19940720_0520_value"

	result := addDateTimePrefix("value", testDate)
	if result != expectedResult {
		t.Errorf("%s should be %s", result, expectedResult)
	}
}
