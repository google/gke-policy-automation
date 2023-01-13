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

	container "cloud.google.com/go/container/apiv1"
	"cloud.google.com/go/container/apiv1/containerpb"
	"github.com/google/gke-policy-automation/internal/log"
	"github.com/google/gke-policy-automation/internal/version"
	gax "github.com/googleapis/gax-go/v2"
	"google.golang.org/api/option"
)

const (
	gkeAPIInputID          = "gkeAPI"
	gkeDataSourceName      = "gke"
	gkeAPIInputDescription = "GKE cluster data from GCP API"
)

type gkeAPIInput struct {
	ctx    context.Context
	client clusterManagerClient
}

type clusterManagerClient interface {
	GetCluster(ctx context.Context, req *containerpb.GetClusterRequest, opts ...gax.CallOption) (*containerpb.Cluster, error)
	Close() error
}

func NewGKEApiInput(ctx context.Context) (Input, error) {
	return newGKEApiInput(ctx, nil)
}

func NewGKEApiInputWithCredentials(ctx context.Context, credentialsFile string) (Input, error) {
	opts := []option.ClientOption{option.WithCredentialsFile(credentialsFile)}
	return newGKEApiInput(ctx, opts)
}

func newGKEApiInput(ctx context.Context, opts []option.ClientOption) (Input, error) {
	opts = append(opts, option.WithUserAgent(version.UserAgent))
	cli, err := container.NewClusterManagerClient(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return &gkeAPIInput{
		ctx:    ctx,
		client: cli,
	}, nil
}

func (i *gkeAPIInput) GetID() string {
	return gkeAPIInputID
}

func (i *gkeAPIInput) GetDescription() string {
	return gkeAPIInputDescription
}

func (i *gkeAPIInput) GetDataSourceName() string {
	return gkeDataSourceName
}

func (i *gkeAPIInput) GetData(clusterID string) (interface{}, error) {
	req := &containerpb.GetClusterRequest{
		Name: clusterID}
	log.Debugf("Fetching cluster data with request %v", req)
	cluster, err := i.client.GetCluster(i.ctx, req)
	if err != nil {
		return nil, err
	}
	return cluster, nil
}

func (i *gkeAPIInput) Close() error {
	if i.client != nil {
		return i.client.Close()
	}
	return nil
}

func GetClusterName(project string, location string, name string) string {
	return fmt.Sprintf("projects/%s/locations/%s/clusters/%s", project, location, name)
}
