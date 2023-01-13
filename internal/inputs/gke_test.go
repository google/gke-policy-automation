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

package inputs

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"cloud.google.com/go/container/apiv1/containerpb"
	gax "github.com/googleapis/gax-go/v2"
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

func TestNewGKEApiInputWithCredentials(t *testing.T) {
	testCredsFile := "test-fixtures/test_credentials.json"
	input, err := NewGKEApiInputWithCredentials(context.Background(), testCredsFile)
	if err != nil {
		t.Fatalf("err = %v; want nil", err)
	}
	_, ok := input.(*gkeAPIInput)
	if !ok {
		t.Fatalf("input is not *gkeAPIInput")
	}
}

func TestGetID(t *testing.T) {
	input := gkeAPIInput{}
	if id := input.GetID(); id != gkeAPIInputID {
		t.Fatalf("id = %v; want %v", id, gkeAPIInputID)
	}
}

func TestGetDescription(t *testing.T) {
	input := gkeAPIInput{}
	if id := input.GetDescription(); id != gkeAPIInputDescription {
		t.Fatalf("id = %v; want %v", id, gkeAPIInputDescription)
	}
}

func TestGetCluster(t *testing.T) {
	input := gkeAPIInput{
		ctx:    context.Background(),
		client: &mockClusterManagerClient{},
	}
	projectID := "test-project"
	clusterLocation := "europe-central2"
	clusterName := "warsaw"
	data, err := input.GetData(GetClusterName(projectID, clusterLocation, clusterName))
	if err != nil {
		t.Fatalf("error when fetching cluster: %v", err)
	}
	cluster, ok := data.(*containerpb.Cluster)
	if !ok {
		t.Fatalf("data is not *containerpb.Cluster")
	}
	if cluster.Name != clusterName {
		t.Errorf("cluster.Name = %s; want %s", cluster.Name, clusterName)
	}
	if cluster.Location != clusterLocation {
		t.Errorf("cluster.Name = %s; want %s", cluster.Location, clusterLocation)
	}
}

func TestClose(t *testing.T) {
	input := gkeAPIInput{
		ctx:    nil,
		client: &mockClusterManagerClient{}}
	err := input.Close()
	if err == nil {
		t.Errorf("gkeAPIInput close() error is nil; want mocked error")
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
