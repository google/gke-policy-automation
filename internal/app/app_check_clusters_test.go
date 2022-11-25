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
	"reflect"
	"testing"

	cfg "github.com/google/gke-policy-automation/internal/config"
	"github.com/google/gke-policy-automation/internal/gke"
	"github.com/google/gke-policy-automation/internal/outputs"
)

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
