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

package clients

import (
	"context"
	"regexp"
	"sync"
	"time"

	"github.com/google/gke-policy-automation/internal/log"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/config"
	pmodel "github.com/prometheus/common/model"
)

type MetricQuery struct {
	Name  string
	Query string
}

type Metric struct {
	Name  string
	Value string
}

type MetricsClient interface {
	GetMetric(query MetricQuery, clusterName string) (string, error)
	GetMetricsForCluster(queries []MetricQuery, clusterName string) (map[string]Metric, error)
}

type metricsClient struct {
	ctx           context.Context
	client        api.Client
	api           v1.API
	maxGoRoutines int
}

func newMetricsClient(ctx context.Context, projectId string, authToken string, maxGoroutines int) (MetricsClient, error) {

	// Creates a client.
	client, err := api.NewClient(api.Config{
		Address:      "https://monitoring.googleapis.com/v1/projects/" + projectId + "/location/global/prometheus/",
		RoundTripper: config.NewAuthorizationCredentialsRoundTripper("Bearer", config.Secret(authToken), api.DefaultRoundTripper),
	})

	if err != nil {
		log.Fatalf("Failed to create metrics client: %v", err)
		return nil, err
	}

	api := v1.NewAPI(client)

	return &metricsClient{
		ctx:           ctx,
		client:        client,
		api:           api,
		maxGoRoutines: maxGoroutines,
	}, nil
}

type metricsClientBuilder struct {
	ctx           context.Context
	projectId     string
	authToken     string
	maxGoroutines int
	timeout       int
}

func NewMetricsClientBuilder(ctx context.Context, projectId string, authToken string) *metricsClientBuilder {
	return &metricsClientBuilder{
		ctx:       ctx,
		projectId: projectId,
		authToken: authToken,
	}
}

func (b *metricsClientBuilder) WithMaxGoroutines(maxGoroutines int) *metricsClientBuilder {
	b.maxGoroutines = maxGoroutines
	return b
}

func (b *metricsClientBuilder) WithTimeout(timeout int) *metricsClientBuilder {
	b.timeout = timeout
	return b
}

func (b *metricsClientBuilder) Build() (MetricsClient, error) {

	var maxGoRoutines = defaultMaxGoroutines
	if b.maxGoroutines != 0 {
		maxGoRoutines = b.maxGoroutines
	}

	metricsClient, err := newMetricsClient(b.ctx, b.projectId, b.authToken, maxGoRoutines)

	if err != nil {
		log.Fatalf("Failed to create metrics client: %v", err)
		return nil, err
	}
	return metricsClient, nil
}

func (m *metricsClient) GetMetric(metricQuery MetricQuery, clusterName string) (string, error) {

	query := metricQuery.Query

	query = replaceWildcard("CLUSTER_NAME", clusterName, query)

	log.Debugf("Querying metric client with query: " + query)

	result, warnings, err := m.api.Query(m.ctx, query, time.Now())
	if err != nil {
		log.Fatalf("Failed to query metrics client: %v", err)
		return "", err
	}
	if warnings != nil {
		log.Warnf("Warning when querying metrics client: %v", warnings)
	}

	queryResults := make([]string, 0, 1)

	data, ok := result.(pmodel.Vector)
	if !ok {
		log.Fatalf("Unsupported result format: %s", result.Type().String())
		return "", err
	}
	for _, v := range data {
		queryResults = append(queryResults, v.Value.String())
	}

	ret := ""
	if len(queryResults) > 0 {
		ret = queryResults[0]
		if len(queryResults) > 1 {
			log.Warnf("query %s returned more than one result for cluster %s", query, clusterName)
		}
	} else {
		log.Warnf("query %s returned no value found for cluster %s", query, clusterName)
	}

	return ret, nil
}

func (m *metricsClient) GetMetricsForCluster(queries []MetricQuery, clusterName string) (map[string]Metric, error) {

	metricsResult := make(map[string]Metric)

	queryChannel := make(chan MetricQuery, m.maxGoRoutines)

	go func() {
		for _, q := range queries {
			queryChannel <- q
		}
		close(queryChannel)
	}()

	resultsChannel := make(chan Metric, m.maxGoRoutines)
	errorChannel := make(chan error, m.maxGoRoutines)

	go func() {
		wg := new(sync.WaitGroup)
		wg.Add(m.maxGoRoutines)

		for gr := 0; gr < m.maxGoRoutines; gr++ {
			log.Debugf("Starting getMetrics goroutine")
			go func() {
				for q := range queryChannel {
					log.Debugf("GetMetric for %s, cluster %s", q, clusterName)
					r, err := m.GetMetric(q, clusterName)
					if err != nil {
						log.Debugf("unable to get metric: %s", err)
						errorChannel <- err
						wg.Done()
					}
					metricResult := Metric{
						Name:  q.Name,
						Value: r,
					}
					resultsChannel <- metricResult
				}
				wg.Done()
			}()
		}
		wg.Wait()
		close(resultsChannel)
		close(errorChannel)

	}()

	if len(errorChannel) > 0 {
		err := <-errorChannel
		log.Errorf("unable to get metric: %s", err)
		return nil, err
	}
	for result := range resultsChannel {
		metricsResult[result.Name] = result
	}
	return metricsResult, nil
}

func replaceWildcard(wildcard string, value string, query string) string {
	clusterNameExp := regexp.MustCompile(wildcard)

	return clusterNameExp.ReplaceAllString(query, "\""+value+"\"")
}
