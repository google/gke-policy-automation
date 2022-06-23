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
	"os"
	"reflect"
	"testing"
	"time"

	"gopkg.in/yaml.v2"

	cfg "github.com/google/gke-policy-automation/internal/config"
	"github.com/google/gke-policy-automation/internal/gke"
	"github.com/google/gke-policy-automation/internal/outputs"
)

type DiscoveryClientMock struct {
	GetClustersInProjectFn func(name string) ([]string, error)
	GetClustersInFolderFn  func(number string) ([]string, error)
	GetClustersInOrgFn     func(number string) ([]string, error)
	CloseFn                func() error
}

func (m DiscoveryClientMock) GetClustersInProject(name string) ([]string, error) {
	return m.GetClustersInProjectFn(name)
}

func (m DiscoveryClientMock) GetClustersInFolder(number string) ([]string, error) {
	return m.GetClustersInFolderFn(number)
}

func (m DiscoveryClientMock) GetClustersInOrg(number string) ([]string, error) {
	return m.GetClustersInOrgFn(number)
}

func (m DiscoveryClientMock) Close() error {
	return m.CloseFn()
}

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
	if len(paApp.collectors) == 0 {
		t.Fatalf("policyAutomationApp collector is nil")
	}
	if _, ok := paApp.collectors[0].(*outputs.ConsoleResultCollector); !ok {
		t.Fatalf("policyAutomationApp collector is not ConsoleResultCollector (default)")
	}
}

func TestGetClusters_config(t *testing.T) {
	clusters := []string{"cluster1", "cluster2"}
	pa := PolicyAutomationApp{
		config: &cfg.Config{
			Clusters: []cfg.ConfigCluster{
				{ID: clusters[0]},
				{ID: clusters[1]},
			},
		},
	}
	results, err := pa.getClusters()
	if err != nil {
		t.Fatalf("err is not nil; want nil; err = %s", err)
	}
	if !reflect.DeepEqual(results, clusters) {
		t.Errorf("results = %v; want %v", results, clusters)
	}
}

func TestGetClusters_discovery(t *testing.T) {
	pa := PolicyAutomationApp{
		out: outputs.NewSilentOutput(),
		ctx: context.Background(),
		config: &cfg.Config{
			CredentialsFile: "test-fixtures/test_credentials.json",
			ClusterDiscovery: cfg.ClusterDiscovery{
				Enabled: true,
			},
		},
	}
	_, err := pa.getClusters()
	if err != nil {
		t.Fatalf("err is not nil; want nil; err = %s", err)
	}
	if _, ok := pa.discovery.(*gke.AssetInventoryDiscoveryClient); !ok {
		t.Errorf("policy automation discovery client is not *gke.AssetInventoryDiscoveryClient")
	}
}

func TestDiscoverClusters_org(t *testing.T) {
	clusters := []string{"clusterOne", "clusterTwo"}
	orgNumber := "123456789"
	clusterInOrgFn := func(number string) ([]string, error) {
		if number != orgNumber {
			t.Errorf("received org number is %v; want %v", number, orgNumber)
		}
		return clusters, nil
	}
	pa := PolicyAutomationApp{
		out:       outputs.NewSilentOutput(),
		discovery: DiscoveryClientMock{GetClustersInOrgFn: clusterInOrgFn},
		config:    &cfg.Config{ClusterDiscovery: cfg.ClusterDiscovery{Enabled: true, Organization: orgNumber}},
	}
	results, err := pa.discoverClusters()
	if err != nil {
		t.Fatalf("err is not nil; want nil; err = %s", err)
	}
	if !reflect.DeepEqual(results, clusters) {
		t.Fatalf("results are %v; want %v", results, clusters)
	}
}

func TestDiscoverClusters_folders(t *testing.T) {
	folders := []string{"12345", "6789"}
	foldersContent := map[string][]string{
		folders[0]: {"clusterOne", "clusterTwo"},
		folders[1]: {"clusterThree", "clusterFour"},
	}
	clusterInFoldersFn := func(number string) ([]string, error) {
		clusters, ok := foldersContent[number]
		if !ok {
			t.Errorf("received folder number = %v; not defined in a test", number)
		}
		return clusters, nil
	}
	pa := PolicyAutomationApp{
		out:       outputs.NewSilentOutput(),
		discovery: DiscoveryClientMock{GetClustersInFolderFn: clusterInFoldersFn},
		config:    &cfg.Config{ClusterDiscovery: cfg.ClusterDiscovery{Enabled: true, Folders: folders}},
	}
	results, err := pa.discoverClusters()
	if err != nil {
		t.Fatalf("err is not nil; want nil; err = %s", err)
	}
	allFoldersContent := make([]string, 0)
	allFoldersContent = append(allFoldersContent, foldersContent[folders[0]]...)
	allFoldersContent = append(allFoldersContent, foldersContent[folders[1]]...)
	if !reflect.DeepEqual(results, allFoldersContent) {
		t.Fatalf("results are %v; want %v", results, allFoldersContent)
	}
}

func TestDiscoverClusters_projects(t *testing.T) {
	projects := []string{"projectOne", "projectTwo"}
	projectsContent := map[string][]string{
		projects[0]: {"clusterOne", "clusterTwo"},
		projects[1]: {"clusterThree", "clusterFour"},
	}
	clusterInProjectsFn := func(name string) ([]string, error) {
		clusters, ok := projectsContent[name]
		if !ok {
			t.Errorf("received project name = %v; not defined in a test", name)
		}
		return clusters, nil
	}
	pa := PolicyAutomationApp{
		out:       outputs.NewSilentOutput(),
		discovery: DiscoveryClientMock{GetClustersInProjectFn: clusterInProjectsFn},
		config:    &cfg.Config{ClusterDiscovery: cfg.ClusterDiscovery{Enabled: true, Projects: projects}},
	}
	results, err := pa.discoverClusters()
	if err != nil {
		t.Fatalf("err is not nil; want nil; err = %s", err)
	}
	allProjectsContent := make([]string, 0)
	allProjectsContent = append(allProjectsContent, projectsContent[projects[0]]...)
	allProjectsContent = append(allProjectsContent, projectsContent[projects[1]]...)
	if !reflect.DeepEqual(results, allProjectsContent) {
		t.Fatalf("results are %v; want %v", results, allProjectsContent)
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
	config := &cfg.Config{}
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
	validateFnMock := func(config cfg.Config) error {
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

func TestLoadCliConfig_defaults(t *testing.T) {
	cliConfig := &CliConfig{
		CredentialsFile: "./test-fixtures/test_credentials.json",
		ClusterName:     "test",
		ClusterLocation: "europe-central2",
		ProjectName:     "my-project",
	}
	pa := PolicyAutomationApp{ctx: context.Background()}
	err := pa.LoadCliConfig(cliConfig, nil)
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

func TestGetClusterName(t *testing.T) {
	input := []cfg.ConfigCluster{
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
	input := cfg.ConfigCluster{Name: "test", Location: "europe-east2"}
	_, err := getClusterName(input)
	if err == nil {
		t.Errorf("error is nil; want error")
	}
}

func TestClusterReviewWithNoPolicies(t *testing.T) {

	pa := PolicyAutomationApp{
		out: outputs.NewSilentOutput(),
		config: &cfg.Config{
			Policies: []cfg.ConfigPolicy{},
		},
	}

	err := pa.Check()

	if err != errNoPolicies {
		t.Fatalf("need noPoliciesError but err = %s", err)
	}
}

func TestPolicyAutomationAppClose_negative(t *testing.T) {
	closeErr := fmt.Errorf("close error")
	pa := PolicyAutomationApp{
		discovery: DiscoveryClientMock{
			CloseFn: func() error {
				return closeErr
			},
		},
	}
	err := pa.Close()
	if err == nil {
		t.Fatalf("error is nil; want error")
	}
	if err != closeErr {
		t.Errorf("error is %v; want %v", err, closeErr)
	}
}

func TestAddDatetimePrefix(t *testing.T) {

	testDate := time.Date(1994, 7, 20, 5, 20, 0, 0, time.UTC)
	expectedResult := "19940720_0520_value"

	result := addDatetimePrefix("value", testDate)

	if result != expectedResult {
		t.Errorf("%s should be %s", result, expectedResult)
	}
}
