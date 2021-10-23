package gke

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"testing"

	container "cloud.google.com/go/container/apiv1"
	gax "github.com/googleapis/gax-go/v2"
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
	c, err := NewGKEClient(context.Background())
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
	cluster, err := client.GetCluster(projectID, clusterLocation, clusterName)
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
