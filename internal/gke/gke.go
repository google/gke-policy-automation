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

	container "cloud.google.com/go/container/apiv1"
	gax "github.com/googleapis/gax-go/v2"
	"google.golang.org/api/option"
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

func NewClient(ctx context.Context) (*GKEClient, error) {
	return newGKEClient(ctx)
}

func NewClientWithCredentialsFile(ctx context.Context, credentialsFile string) (*GKEClient, error) {
	return newGKEClient(ctx, option.WithCredentialsFile(credentialsFile))
}

func newGKEClient(ctx context.Context, opts ...option.ClientOption) (*GKEClient, error) {
	cli, err := container.NewClusterManagerClient(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return &GKEClient{
		ctx:    ctx,
		client: cli,
	}, nil
}

func (c *GKEClient) GetCluster(name string) (*containerpb.Cluster, error) {
	req := &containerpb.GetClusterRequest{
		Name: name}
	return c.client.GetCluster(c.ctx, req)
}

func (c *GKEClient) Close() error {
	return c.client.Close()
}

func GetClusterName(project string, location string, name string) string {
	return fmt.Sprintf("projects/%s/locations/%s/clusters/%s", project, location, name)
}
