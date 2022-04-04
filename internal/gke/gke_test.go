// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
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
	gax "github.com/googleapis/gax-go/v2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	containerpb "google.golang.org/genproto/googleapis/container/v1"
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
	}, nil
}

func (mockClusterManagerClient) Close() error {
	return fmt.Errorf("mocked error")
}

func TestNewGKEClient(t *testing.T) {
	ctx := context.Background()
	testCreds, err := google.CredentialsFromJSON(ctx, []byte("{\"type\": \"authorized_user\",\"client_id\": \"dummy\"}"), container.DefaultAuthScopes()...)
	if err != nil {
		t.Fatalf("error when creating JSON credentials: %v", err)
	}
	c, err := newGKEClient(context.Background(), option.WithCredentials(testCreds))
	if err != nil {
		t.Fatalf("error when creating client: %v", err)
	}
	typeA := reflect.TypeOf(c.client)
	typeB := reflect.TypeOf(&container.ClusterManagerClient{})
	if typeA != typeB {
		t.Errorf("ClusterManagerClient type = %s; want %s", typeA, typeB)
	}
}

func TestGetCluster(t *testing.T) {
	client := GKEClient{
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
	client := GKEClient{
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
