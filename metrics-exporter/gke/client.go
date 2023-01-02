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

// Package gke implements Google Kubernetes Engine related functions, mostly for creating
// kube config with a data acquired from GKE API
package gke

import (
	"context"

	container "cloud.google.com/go/container/apiv1"
	"cloud.google.com/go/container/apiv1/containerpb"
	"github.com/google/gke-policy-automation/metrics-exporter/log"
	"github.com/google/gke-policy-automation/metrics-exporter/version"
	"github.com/googleapis/gax-go"
	"google.golang.org/api/option"
)

type clusterManagerClient interface {
	GetCluster(ctx context.Context, req *containerpb.GetClusterRequest, opts ...gax.CallOption) (*containerpb.Cluster, error)
	Close() error
}

type gkeClient struct {
	ctx    context.Context
	client clusterManagerClient
}

func NewClient(ctx context.Context) (*gkeClient, error) {
	opts := []option.ClientOption{option.WithUserAgent(version.UserAgent)}
	cli, err := container.NewClusterManagerClient(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return &gkeClient{
		ctx:    ctx,
		client: cli,
	}, nil
}

func (c *gkeClient) GetData(clusterID string) (*containerpb.Cluster, error) {
	req := &containerpb.GetClusterRequest{
		Name: clusterID}
	log.Debugf("Fetching cluster data with request %v", req)
	return c.client.GetCluster(c.ctx, req)
}

func (c *gkeClient) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}
