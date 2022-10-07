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

package gke

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"testing"

	container "cloud.google.com/go/container/apiv1"
	"github.com/google/gke-policy-automation/internal/config"
	gax "github.com/googleapis/gax-go/v2"
	containerpb "google.golang.org/genproto/googleapis/container/v1"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type mockClusterManagerClient struct {
}

func (mockClusterManagerClient) GetCluster(ctx context.Context, req *containerpb.GetClusterRequest, opts ...gax.CallOption) (*containerpb.Cluster, error) {
	re := regexp.MustCompile(`^projects/([^/]+)/locations/([^/]+)/clusters/([^/]+)$`)
	if !re.MatchString(req.Name) {
		return nil, fmt.Errorf("request name: %q, does not match regexp: %q", req.Name, re.String())
	}
	matches := re.FindStringSubmatch(req.Name)
	return &containerpb.Cluster{
		Name:     matches[3],
		Location: matches[2],
		MasterAuth: &containerpb.MasterAuth{
			ClusterCaCertificate: "dGVzdCBjZXJ0IGRhdGE=",
		},
		Endpoint: "1.1.1.1",
		SelfLink: fmt.Sprintf("https://container.googleapis.com/v1/projects/%s/locations/%s/clusters/%s", matches[1], matches[2], matches[3]),
	}, nil
}

func (mockClusterManagerClient) Close() error {
	return fmt.Errorf("mocked error")
}

func TestNewGKEClient(t *testing.T) {
	testCredsFile := "test-fixtures/test_credentials.json"
	c, err := NewGKEApiClientBuilder(context.Background()).WithCredentialsFile(testCredsFile).
		WithK8SClient([]string{"v1"}, config.DefaultK8SClientQPS).
		Build()
	if err != nil {
		t.Fatalf("error when creating client: %v", err)
	}
	apiClient, ok := c.(*GKEApiClient)
	if !ok {
		t.Fatalf("can't cast GKE client to GKEApiClient")
	}
	typeA := reflect.TypeOf(apiClient.client)
	typeB := reflect.TypeOf(&container.ClusterManagerClient{})
	if typeA != typeB {
		t.Errorf("ClusterManagerClient type = %s; want %s", typeA, typeB)
	}
}

func TestNewGKEClientWithMetrics(t *testing.T) {
	testCredsFile := "test-fixtures/test_credentials.json"
	metricQueries := []MetricQuery{{
		Name:  "xxx",
		Query: "apiserver_storage_objects{resource=\"pods\"}",
	}}

	c, err := NewGKEApiClientBuilder(context.Background()).WithCredentialsFile(testCredsFile).WithMetricsClient(metricQueries).Build()

	if err != nil {
		t.Fatalf("error when creating client: %v", err)
	}
	apiClient, ok := c.(*GKEApiClient)
	if !ok {
		t.Fatalf("can't cast GKE client to GKEApiClient")
	}

	if !reflect.DeepEqual(apiClient.metricQueries, metricQueries) {
		t.Errorf("apiClient metricQueries = %v; want %v", apiClient.metricQueries, metricQueries)
	}
}

type mockK8Client struct {
}

func (mockK8Client) GetNamespaces() ([]string, error) {
	return []string{"namespace-one", "namespace-two"}, nil
}

func (mockK8Client) GetFetchableResourceTypes() ([]*ResourceType, error) {
	return []*ResourceType{
		{
			Group:      "autoscaling",
			Version:    "v1",
			Name:       "horizontalpodautoscalers",
			Namespaced: true,
		},
		{
			Group:      "",
			Version:    "v1",
			Name:       "replicationcontrollers",
			Namespaced: true,
		},
		{
			Group:      "",
			Version:    "v1",
			Name:       "componentstatuses",
			Namespaced: false,
		},
		{
			Group:      "authorization.k8s.io",
			Version:    "v1",
			Name:       "localsubjectaccessreviews",
			Namespaced: true,
		},
	}, nil
}

func (mockK8Client) GetNamespacedResources(resourceType ResourceType, namespace string) ([]*Resource, error) {
	return []*Resource{
		{
			Type: resourceType,
			Data: nil,
		},
	}, nil
}

func (mockK8Client) GetResources(resourceType []*ResourceType, namespace []string) ([]*Resource, error) {
	return []*Resource{
		{
			Type: *resourceType[0],
			Data: nil,
		},
	}, nil
}

func TestGKEApiClientBuilder(t *testing.T) {
	credFile := "test-fixtures/test_credentials.json"
	apiVersions := []string{"policy/v1", "networking.k8s.io/v1"}
	maxQPS := 69
	b := NewGKEApiClientBuilder(context.TODO()).
		WithCredentialsFile(credFile).
		WithK8SClient(apiVersions, maxQPS)
	client, err := b.Build()
	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	apiClient, ok := client.(*GKEApiClient)
	if !ok {
		t.Fatalf("client is not *GKEApiClient")
	}
	if !reflect.DeepEqual(apiClient.k8sApiVersions, apiVersions) {
		t.Errorf("apiClient k8sApiVersions = %v; want %v", apiClient.k8sApiVersions, apiVersions)
	}
	if apiClient.k8sMaxQPS != maxQPS {
		t.Errorf("apiClient k8sMaxQPS = %v; want %v", apiClient.k8sMaxQPS, maxQPS)
	}
	if b.credentialsFile != credFile {
		t.Errorf("builder credentialsFile = %v; want %v", b.credentialsFile, credFile)
	}
}

func TestGetCluster(t *testing.T) {
	client := GKEApiClient{
		ctx:    context.Background(),
		client: &mockClusterManagerClient{},
		k8sClientFunc: func(ctx context.Context, kubeConfig *clientcmdapi.Config, maxQPS int) (KubernetesClient, error) {
			return &mockK8Client{}, nil

		},
		authTokenFunc: func(ctx context.Context) (string, error) {
			return "fake-token", nil
		},
	}
	projectID := "test-project"
	clusterLocation := "europe-central2"
	clusterName := "warsaw"
	cluster, err := client.GetCluster(GetClusterName(projectID, clusterLocation, clusterName))
	if err != nil {
		t.Fatalf("error when fetching cluster: %v", err)
	}
	if cluster.Name != clusterName {
		t.Errorf("cluster.Name = %s; want %s", cluster.Name, clusterName)
	}
	if cluster.Location != clusterLocation {
		t.Errorf("cluster.Name = %s; want %s", cluster.Location, clusterLocation)
	}
}

func TestGetClusterWithoutK8SApiCheckConfigured(t *testing.T) {
	client := GKEApiClient{
		ctx:    context.Background(),
		client: &mockClusterManagerClient{},
	}
	projectID := "test-project"
	clusterLocation := "europe-central2"
	clusterName := "warsaw"
	cluster, err := client.GetCluster(GetClusterName(projectID, clusterLocation, clusterName))
	if err != nil {
		t.Fatalf("error when fetching cluster: %v", err)
	}
	if cluster.Name != clusterName {
		t.Errorf("cluster.Name = %s; want %s", cluster.Name, clusterName)
	}
	if cluster.Location != clusterLocation {
		t.Errorf("cluster.Name = %s; want %s", cluster.Location, clusterLocation)
	}
}

func TestClose(t *testing.T) {
	client := GKEApiClient{
		ctx:    nil,
		client: &mockClusterManagerClient{}}
	err := client.Close()
	if err == nil {
		t.Errorf("GKEClient close() error is nil; want mocked error")
	}
}

func TestGetClusterName(t *testing.T) {
	projectID := "test-project"
	clusterLocation := "europe-central2"
	clusterName := "warsaw"
	name := GetClusterName(projectID, clusterLocation, clusterName)
	re := regexp.MustCompile(`^projects/([^/]+)/locations/([^/]+)/clusters/([^/]+)$`)
	if !re.MatchString(name) {
		t.Fatalf("name: %q, does not match regexp: %q", name, re.String())
	}
	matches := re.FindStringSubmatch(name)
	if matches[1] != projectID {
		t.Errorf("match[1] = %v; want %v", matches[1], projectID)
	}
	if matches[2] != clusterLocation {
		t.Errorf("match[2] = %v; want %v", matches[2], clusterLocation)
	}
	if matches[3] != clusterName {
		t.Errorf("match[3] = %v; want %v", matches[3], clusterName)
	}
}

func TestGetClusterResourcesForEmptyConfig(t *testing.T) {
	client := GKEApiClient{
		ctx:    context.Background(),
		client: &mockClusterManagerClient{},
		k8sClientFunc: func(ctx context.Context, kubeConfig *clientcmdapi.Config, maxQPS int) (KubernetesClient, error) {
			return &mockK8Client{}, nil

		},
		authTokenFunc: func(ctx context.Context) (string, error) {
			return "fake-token", nil
		},
		k8sApiVersions: []string{},
	}
	projectID := "test-project"
	clusterLocation := "europe-central2"
	clusterName := "warsaw"
	cluster, err := client.GetCluster(GetClusterName(projectID, clusterLocation, clusterName))
	if err != nil {
		t.Fatalf("error when fetching cluster: %v", err)
	}
	if len(cluster.Resources) > 0 {
		t.Errorf("should not return any resources for empty configuration. Returned %d; want 0", len(cluster.Resources))
	}
}

func TestGetClusterResourcesForNonEmptyConfig(t *testing.T) {
	client := GKEApiClient{
		ctx:    context.Background(),
		client: &mockClusterManagerClient{},
		k8sClientFunc: func(ctx context.Context, kubeConfig *clientcmdapi.Config, maxQPS int) (KubernetesClient, error) {
			return &mockK8Client{}, nil

		},
		authTokenFunc: func(ctx context.Context) (string, error) {
			return "fake-token", nil
		},
		k8sApiVersions: []string{"v1"},
	}
	projectID := "test-project"
	clusterLocation := "europe-central2"
	clusterName := "warsaw"
	cluster, err := client.GetCluster(GetClusterName(projectID, clusterLocation, clusterName))
	if err != nil {
		t.Fatalf("error when fetching cluster: %v", err)
	}
	if len(cluster.Resources) == 0 {
		t.Errorf("should return resources for v1 configuration. Returned %d; want 1", len(cluster.Resources))
	}
}

func TestReadableId(t *testing.T) {
	expected := "projects/test/zones/europe-north1-a/clusters/cluster-demo"
	cluster := &Cluster{
		Cluster: &containerpb.Cluster{
			SelfLink: fmt.Sprintf("%s/%s", "https://container.googleapis.com/v1", expected),
		},
	}
	readableId := cluster.ReadableId()
	if readableId != expected {
		t.Errorf("readable id = %v; want %v", readableId, expected)
	}
}

func TestGetProjectId(t *testing.T) {
	expected := "test-project"
	selfLink := "https://container.googleapis.com/v1/projects/test-project/zones/europe-central2-a/clusters/test-cluster/nodePools/default-pool"
	result := getProjectIdFromSelfLink(selfLink)
	if result != expected {
		t.Errorf("projectId = %v; want %v", result, expected)
	}
}
