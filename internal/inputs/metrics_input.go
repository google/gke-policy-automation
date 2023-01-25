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

	"github.com/google/gke-policy-automation/internal/gke"
	"github.com/google/gke-policy-automation/internal/inputs/clients"
	"github.com/google/gke-policy-automation/internal/log"
)

const (
	metricsInputID          = "metricsAPI"
	metricsDataSourceName   = "monitoring"
	metricsInputDescription = "Cluster metrics data from Prometheus API"
)

type newMetricsClientFunc func(ctx context.Context, projectId string, authToken string) (clients.MetricsClient, error)

type metricsInput struct {
	ctx                  context.Context
	tokenSource          clients.TokenSource
	newMetricsClientFunc newMetricsClientFunc
	metricsClient        clients.MetricsClient
	projectID            string
	queries              []clients.MetricQuery
	maxGoRoutines        int
	timeoutSeconds       int
}

type metricsInputBuilder struct {
	ctx             context.Context
	credentialsFile string
	projectID       string
	queries         []clients.MetricQuery
	maxGoRoutines   int
	timeoutSeconds  int
}

func NewMetricsInputBuilder(ctx context.Context, queries []clients.MetricQuery) *metricsInputBuilder {
	return &metricsInputBuilder{
		ctx:     ctx,
		queries: queries,
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

func (b *metricsInputBuilder) Build() (Input, error) {
	var ts clients.TokenSource
	var err error

	if b.credentialsFile != "" {
		ts, err = clients.NewGoogleTokenSourceWithCredentials(b.ctx, b.credentialsFile)
		if err != nil {
			return nil, err
		}
	} else {
		ts, err = clients.NewGoogleTokenSource(b.ctx)
		if err != nil {
			return nil, err
		}
	}

	input := &metricsInput{
		ctx:            b.ctx,
		tokenSource:    ts,
		projectID:      b.projectID,
		queries:        b.queries,
		maxGoRoutines:  b.maxGoRoutines,
		timeoutSeconds: b.timeoutSeconds,
	}
	input.newMetricsClientFunc = input.newMetricsClientFromBuilder

	return input, nil
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
		log.Error("Error parsing clusterId: " + err.Error())
		return nil, err
	}

	if i.metricsClient == nil {
		log.Debugf("Empty client - creating one for %v", clusterID)
		if err := i.createMetricsClient(projectID); err != nil {
			return nil, err
		}
	}

	data, err := i.metricsClient.GetMetricsForCluster(i.queries, clusterID)
	if err != nil {
		log.Errorf("Error fetching metric: %s", err)
		return nil, err
	}

	return data, nil
}

func (i *metricsInput) Close() error {
	log.Debugf("closing metrics input")
	return nil
}

func (i *metricsInput) newMetricsClientFromBuilder(ctx context.Context, projectID string, authToken string) (clients.MetricsClient, error) {
	client, err := clients.NewMetricsClientBuilder(ctx, projectID, authToken).
		WithMaxGoroutines(i.maxGoRoutines).
		WithTimeout(i.timeoutSeconds).
		Build()
	return client, err
}

func (i *metricsInput) createMetricsClient(clusterProjectID string) error {
	token, err := i.tokenSource.GetAuthToken()
	if err != nil {
		return err
	}

	var projectID string
	if i.projectID != "" {
		projectID = i.projectID
	} else {
		projectID = clusterProjectID
	}

	i.metricsClient, err = i.newMetricsClientFunc(i.ctx, projectID, token)
	if err != nil {
		return err
	}
	return nil
}
