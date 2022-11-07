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
	"regexp"
	"strings"

	"github.com/google/gke-policy-automation/internal/inputs/clients"
	"github.com/google/gke-policy-automation/internal/log"
)

const (
	metricsInputID          = "metricsAPI"
	metricsInputDescription = "Cluster metrics data from Prometheus API"
)

type newMetricsClientFunc func(ctx context.Context, projectId string, authToken string) (clients.MetricsClient, error)

type metricsInput struct {
	ctx                  context.Context
	tokenSource          clients.TokenSource
	metricsInput         Input
	newMetricsClientFunc newMetricsClientFunc
	metricsClient        clients.MetricsClient
	projectId            string
	queries              []clients.MetricQuery
	maxGoRoutines        int
	timeoutSeconds       int
}

type metricsInputBuilder struct {
	ctx             context.Context
	credentialsFile string
	projectId       string
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

func (b *metricsInputBuilder) WithProjectId(projectId string) *metricsInputBuilder {
	b.projectId = projectId
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
		projectId:      b.projectId,
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

func (i *metricsInput) GetData(clusterID string) (interface{}, error) {
	if i.metricsClient == nil {
		log.Debugf("Empty client - creating one for %v", clusterID)
		if err := i.createMetricsClient(getProjectIdFromClusterId(clusterID)); err != nil {
			return nil, err
		}
	}

	data, err := i.metricsClient.GetMetricsForCluster(i.queries, getClusterNameFromClusterId(clusterID))
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

func (i *metricsInput) newMetricsClientFromBuilder(ctx context.Context, projectId string, authToken string) (clients.MetricsClient, error) {
	client, err := clients.NewMetricsClientBuilder(ctx, projectId, authToken).
		WithMaxGoroutines(i.maxGoRoutines).
		WithTimeout(i.timeoutSeconds).
		Build()
	return client, err
}

func (i *metricsInput) createMetricsClient(clusterProjectId string) error {
	token, err := i.tokenSource.GetAuthToken()
	if err != nil {
		return err
	}

	var projectId string
	if i.projectId != "" {
		projectId = i.projectId
	} else {
		projectId = clusterProjectId
	}

	i.metricsClient, err = i.newMetricsClientFunc(i.ctx, projectId, token)
	if err != nil {
		return err
	}
	return nil
}

func getProjectIdFromClusterId(clusterId string) string {

	cuttingBySlash := sliceAndValidateClusterId(clusterId)

	if cuttingBySlash == nil || len(cuttingBySlash) < 2 {
		log.Error("Error getting project id from clusterId: " + clusterId)
	}
	return cuttingBySlash[1]
}

func getClusterNameFromClusterId(clusterId string) string {

	cuttingBySlash := sliceAndValidateClusterId(clusterId)

	if cuttingBySlash == nil || len(cuttingBySlash) < 6 {
		log.Error("Error getting cluster name from clusterId: " + clusterId)
	}
	return cuttingBySlash[5]
}

func sliceAndValidateClusterId(clusterId string) []string {
	r := regexp.MustCompile(`projects/.+/locations/.+/clusters/.+`)
	if !r.MatchString(clusterId) {
		log.Errorf("cluster id %s does not match clusterId format", clusterId)
		return nil
	}
	matches := r.FindStringSubmatch(clusterId)

	if len(matches) < 1 {
		log.Errorf("cluster id %s does not match clusterId format", clusterId)
		return nil
	}
	match := matches[0]
	cuttingBySlash := strings.FieldsFunc(match, func(r rune) bool {
		if r == '/' {
			return true
		}
		return false
	})

	return cuttingBySlash
}
