package gke

import (
	"context"
	"fmt"

	container "cloud.google.com/go/container/apiv1"
	gax "github.com/googleapis/gax-go/v2"
	containerpb "google.golang.org/genproto/googleapis/container/v1"
)

type ClusterManagerClient interface {
	GetCluster(ctx context.Context, req *containerpb.GetClusterRequest, opts ...gax.CallOption) (*containerpb.Cluster, error)
	Close() error
}

type GKEClient struct {
	ctx    context.Context
	client ClusterManagerClient
}

func NewGKEClient(ctx context.Context) (*GKEClient, error) {
	cli, err := container.NewClusterManagerClient(ctx)
	if err != nil {
		return nil, err
	}
	return &GKEClient{
		ctx:    ctx,
		client: cli,
	}, nil
}

func (c *GKEClient) GetCluster(project string, location string, name string) (*containerpb.Cluster, error) {
	req := &containerpb.GetClusterRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/clusters/%s", project, location, name)}
	return c.client.GetCluster(c.ctx, req)
}

func (c *GKEClient) Close() error {
	return c.client.Close()
}
