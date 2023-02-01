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
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"sync"
	"time"

	"github.com/google/gke-policy-automation/internal/gke"
	"github.com/google/gke-policy-automation/internal/log"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/config"
	pmodel "github.com/prometheus/common/model"
)

const (
	MetricQueryWildcardClusterID       = `\$CLUSTER_ID`
	MetricQueryWildcardClusterName     = `\$CLUSTER_NAME`
	MetricQueryWildcardClusterLocation = `\$CLUSTER_LOCATION`
	MetricQueryWildcardClusterProject  = `\$CLUSTER_PROJECT`
)

type MetricQuery struct {
	Name  string
	Query string
}

type Metric struct {
	Name        string                 `json:"name"`
	ScalarValue float64                `json:"scalar"`
	VectorValue map[string]interface{} `json:"vector"`
}

type MetricsClient interface {
	GetMetric(query MetricQuery, clusterID string) (*Metric, error)
	GetMetricsForCluster(queries []MetricQuery, clusterID string) (map[string]Metric, error)
}

type metricsClient struct {
	ctx           context.Context
	client        api.Client
	api           v1.API
	maxGoRoutines int
}

func newMetricsClient(ctx context.Context, projectID string, authToken string, maxGoroutines int) (MetricsClient, error) {
	// Creates a client.
	client, err := api.NewClient(api.Config{
		Address:      "https://monitoring.googleapis.com/v1/projects/" + projectID + "/location/global/prometheus/",
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
	projectID     string
	authToken     string
	maxGoroutines int
	timeout       int
}

func NewMetricsClientBuilder(ctx context.Context, projectID string, authToken string) *metricsClientBuilder {
	return &metricsClientBuilder{
		ctx:       ctx,
		projectID: projectID,
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

	metricsClient, err := newMetricsClient(b.ctx, b.projectID, b.authToken, maxGoRoutines)

	if err != nil {
		log.Fatalf("failed to create metrics client: %v", err)
		return nil, err
	}
	return metricsClient, nil
}

func (m *metricsClient) GetMetric(metricQuery MetricQuery, clusterID string) (*Metric, error) {
	query := replaceAllWildcards(clusterID, metricQuery.Query)
	log.Debugf("querying metric client with a query: %s", query)

	result, warnings, err := m.api.Query(m.ctx, query, time.Now())
	if err != nil {
		log.Fatalf("failed to query metrics client: %v", err)
		return nil, err
	}
	if warnings != nil {
		log.Warnf("warning when querying metrics client: %v", warnings)
	}

	data, ok := result.(pmodel.Vector)
	if !ok {
		badType := reflect.TypeOf(result)
		log.Fatalf("unsupported result format: %s", badType)
		return nil, fmt.Errorf("unsupported result format: %s", badType)
	}

	if len(data) < 1 {
		return nil, fmt.Errorf("empty result vector")
	}

	if data.Len() == 1 {
		dataSample := data[0]
		return &Metric{
			Name:        metricQuery.Name,
			ScalarValue: float64(dataSample.Value),
		}, nil
	}
	vectorValue := make(map[string]interface{})
	for _, dataSample := range data {
		metricValues := valuesFromMetric(dataSample.Metric)
		populateVectorMap(vectorValue, metricValues, float64(dataSample.Value))
	}
	return &Metric{
		Name:        metricQuery.Name,
		VectorValue: vectorValue,
	}, nil
}

func (m *metricsClient) GetMetricsForCluster(queries []MetricQuery, clusterID string) (map[string]Metric, error) {

	metricsResult := make(map[string]Metric)

	queryChannel := make(chan MetricQuery, m.maxGoRoutines)

	go func() {
		for _, q := range queries {
			queryChannel <- q
		}
		close(queryChannel)
	}()

	resultsChannel := make(chan *Metric, m.maxGoRoutines)
	errorChannel := make(chan error, m.maxGoRoutines)

	go func() {
		wg := new(sync.WaitGroup)
		wg.Add(m.maxGoRoutines)

		for gr := 0; gr < m.maxGoRoutines; gr++ {
			log.Debugf("Starting getMetrics goroutine")
			go func() {
				for q := range queryChannel {
					log.Debugf("getMetric for cluster %s, query %q", clusterID, q)
					metric, err := m.GetMetric(q, clusterID)
					if err != nil {
						log.Debugf("unable to get metric for cluster: %s, query: %s, reason: %s", clusterID, q, err)
						errorChannel <- err
					} else {
						resultsChannel <- metric
					}
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
		metricsResult[result.Name] = *result
	}
	return metricsResult, nil
}

func replaceWildcard(wildcard string, value string, query string) string {
	clusterNameExp := regexp.MustCompile(wildcard)

	return clusterNameExp.ReplaceAllString(query, "\""+value+"\"")
}

func replaceAllWildcards(clusterID string, query string) string {
	result := replaceWildcard(MetricQueryWildcardClusterID, clusterID, query)
	if clusterProjectID, clusterLocation, clusterName, err := gke.SliceAndValidateClusterID(clusterID); err == nil {
		result = replaceWildcard(MetricQueryWildcardClusterProject, clusterProjectID, result)
		result = replaceWildcard(MetricQueryWildcardClusterLocation, clusterLocation, result)
		result = replaceWildcard(MetricQueryWildcardClusterName, clusterName, result)
	} else {
		log.Warnf("failed to replace some wildcards due to project identifier validation: %s", err)
	}
	return result
}

func valuesFromMetric(metric pmodel.Metric) []string {
	result := make([]string, 0, len(metric))
	keys := make([]string, 0, len(metric))
	for key := range metric {
		keys = append(keys, string(key))
	}
	sort.Strings(keys)
	for _, key := range keys {
		result = append(result, string(metric[pmodel.LabelName(key)]))
	}
	return result
}

func populateVectorMap(m map[string]interface{}, labels []string, value float64) {
	if len(labels) == 1 {
		m[labels[0]] = value
		return
	}
	label := labels[0]
	mValue, ok := m[label]
	if !ok {
		mValue = make(map[string]interface{})
		m[label] = mValue
	}
	populateVectorMap(mValue.(map[string]interface{}), labels[1:], value)
}
