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

	"github.com/google/gke-policy-automation/internal/gke"
	"github.com/google/gke-policy-automation/internal/inputs/clients"
	"github.com/google/gke-policy-automation/internal/log"
)

const (
	metricsInputID          = "metricsAPI"
	metricsDataSourceName   = "monitoring"
	metricsInputDescription = "Cluster metrics data from Prometheus API"
)

type createTokenSourceFn func(ctx context.Context, credentialsFile string) (clients.TokenSource, error)

type metricsInput struct {
	ctx                 context.Context
	metricsClient       clients.MetricsClient
	projectID           string
	credentialsFile     string
	address             string
	username            string
	password            string
	queries             []clients.MetricQuery
	maxGoRoutines       int
	timeoutSeconds      int
	createTokenSourceFn createTokenSourceFn
}

type metricsInputBuilder struct {
	ctx                 context.Context
	credentialsFile     string
	projectID           string
	address             string
	username            string
	password            string
	queries             []clients.MetricQuery
	maxGoRoutines       int
	timeoutSeconds      int
	createTokenSourceFn createTokenSourceFn
}

func NewMetricsInputBuilder(ctx context.Context, queries []clients.MetricQuery) *metricsInputBuilder {
	return &metricsInputBuilder{
		ctx:                 ctx,
		queries:             queries,
		createTokenSourceFn: createTokenSource,
	}
}

func (b *metricsInputBuilder) WithCredentialsFile(credentialsFile string) *metricsInputBuilder {
	b.credentialsFile = credentialsFile
	return b
}

func (b *metricsInputBuilder) WithMaxGoroutines(maxGoRoutines int) *metricsInputBuilder {
	b.maxGoRoutines = maxGoRoutines
	return b
}

func (b *metricsInputBuilder) WithClientTimeoutSeconds(timeoutSeconds int) *metricsInputBuilder {
	b.timeoutSeconds = timeoutSeconds
	return b
}

func (b *metricsInputBuilder) WithProjectID(projectID string) *metricsInputBuilder {
	b.projectID = projectID
	return b
}

func (b *metricsInputBuilder) WithAddress(address string) *metricsInputBuilder {
	b.address = address
	return b
}

func (b *metricsInputBuilder) WithUsernamePassword(username, password string) *metricsInputBuilder {
	b.username = username
	b.password = password
	return b
}

func (b *metricsInputBuilder) Build() (Input, error) {
	var metricsClient clients.MetricsClient
	var err error
	if b.projectID != "" || b.address != "" {
		log.Debugf("creating global metric client, project %q, address %q", b.projectID, b.address)
		if metricsClient, err = newMetricsClientFromBuilder(b.ctx,
			b.credentialsFile, b.address, b.projectID, b.username, b.password,
			b.maxGoRoutines, b.timeoutSeconds, b.createTokenSourceFn); err != nil {
			return nil, err
		}
	}
	return &metricsInput{
		ctx:                 b.ctx,
		metricsClient:       metricsClient,
		credentialsFile:     b.credentialsFile,
		projectID:           b.projectID,
		address:             b.address,
		username:            b.username,
		password:            b.password,
		queries:             b.queries,
		maxGoRoutines:       b.maxGoRoutines,
		timeoutSeconds:      b.timeoutSeconds,
		createTokenSourceFn: b.createTokenSourceFn,
	}, nil
}

func (i *metricsInput) GetID() string {
	return metricsInputID
}

func (i *metricsInput) GetDescription() string {
	return metricsInputDescription
}

func (i *metricsInput) GetDataSourceName() string {
	return metricsDataSourceName
}

func (i *metricsInput) GetData(clusterID string) (interface{}, error) {
	projectID, _, _, err := gke.SliceAndValidateClusterID(clusterID)
	if err != nil {
		log.Errorf("error parsing clusterID: %s", err)
		return nil, err
	}

	metricsClient := i.metricsClient
	if metricsClient == nil {
		log.Debugf("global metric client is nil, creating scoped client for cluster %s", clusterID)
		if metricsClient, err = newMetricsClientFromBuilder(
			i.ctx, i.credentialsFile, i.address, projectID, i.username, i.password,
			i.maxGoRoutines, i.timeoutSeconds, i.createTokenSourceFn); err != nil {
			return nil, err
		}
	}

	if err := validateKubeStateMetrics(metricsClient, clusterID); err != nil {
		return nil, fmt.Errorf("failed to get results from kube-state-metrics test query for cluster %q, is kube-state-metrics installed?", clusterID)
	}

	data, err := metricsClient.GetMetricsForCluster(i.queries, clusterID)
	if err != nil {
		log.Errorf("error fetching metric: %s", err)
		return nil, err
	}
	return data, nil
}

func (i *metricsInput) Close() error {
	log.Debugf("closing metrics input")
	return nil
}

func newMetricsClientFromBuilder(ctx context.Context,
	credentialsFile, address, projectID, username, password string,
	maxGoRoutines, timeoutSeconds int,
	createTokenSourceFn createTokenSourceFn) (clients.MetricsClient, error) {

	builder := clients.NewMetricsClientBuilder(ctx)
	if address != "" {
		builder.WithAddress(address)
		if username != "" && password != "" {
			builder.WithUsernamePassword(username, password)
		}
	} else {
		if ts, err := createTokenSourceFn(ctx, credentialsFile); err == nil {
			builder.WithGoogleCloudMonitoring(projectID, ts)
		} else {
			return nil, err
		}
	}
	return builder.WithMaxGoroutines(maxGoRoutines).
		WithTimeout(timeoutSeconds).Build()
}

func createTokenSource(ctx context.Context, credentialsFile string) (clients.TokenSource, error) {
	if credentialsFile != "" {
		return clients.NewGoogleTokenSourceWithCredentials(ctx, credentialsFile)
	}
	return clients.NewGoogleTokenSource(ctx)
}

func validateKubeStateMetrics(client clients.MetricsClient, clusterID string) error {
	query := clients.MetricQuery{
		Name:  "namespaces",
		Query: "kube_namespace_created{cluster=$CLUSTER_NAME,location=$CLUSTER_LOCATION,project_id=$CLUSTER_PROJECT}",
	}
	_, err := client.GetMetric(query, clusterID)
	return err
}
