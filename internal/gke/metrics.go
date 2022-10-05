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

package gke

import (
	"context"
	"regexp"
	"sync"
	"time"

	"github.com/google/gke-policy-automation/internal/log"

	//monitoring "cloud.google.com/go/monitoring/apiv3/v2"

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
	Value string //? which type - json
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

func NewMetricClient(ctx context.Context, projectId string, authToken string) (MetricsClient, error) {

	// Creates a client.
	client, err := api.NewClient(api.Config{
		Address:      "https://monitoring.googleapis.com/v1/projects/" + projectId + "/location/global/prometheus/",
		RoundTripper: config.NewAuthorizationCredentialsRoundTripper("Bearer", config.Secret(authToken), api.DefaultRoundTripper),
	}) //plik - credential file

	if err != nil {
		log.Fatalf("Failed to create metrics client: %v", err)
		return nil, err
	}

	api := v1.NewAPI(client)

	return &metricsClient{
		ctx:           ctx,
		client:        client,
		api:           api,
		maxGoRoutines: defaultMaxGoroutines,
	}, nil
}

func (m *metricsClient) GetMetric(metricQuery MetricQuery, clusterName string) (string, error) {

	query := metricQuery.Query
	clusterNameExp := regexp.MustCompile("CLUSTER_NAME")

	query = clusterNameExp.ReplaceAllString(query, "\""+clusterName+"\"")

	log.Debugf("Querying metric client with query: " + query)

	result, warnings, err := m.api.Query(context.Background(), query, time.Now())
	if err != nil {
		log.Fatalf("Failed to query metrics client: %v", err)
		return "", err
	}
	if warnings != nil {
		log.Warnf("Warning when querying metrics client: %v", warnings)
	}

	queryResults := make([]string, 0)

	data, ok := result.(pmodel.Vector)
	if !ok {
		log.Fatalf("Unsupported result format: %s", result.Type().String())
	}
	for _, v := range data {
		queryResults = append(queryResults, v.Value.String())
	}

	ret := ""
	if len(queryResults) > 0 {
		ret = queryResults[0]
	} else {
		log.Debugf("query %s returned no value found for cluster %s", query, clusterName)
	}

	return ret, nil //get first value only
}

func (m *metricsClient) GetMetricsForCluster(queries []MetricQuery, clusterName string) (map[string]Metric, error) {

	metricsResult := make(map[string]Metric)

	queryChannel := make(chan MetricQuery, m.maxGoRoutines)
	wg := new(sync.WaitGroup)
	wg.Add(m.maxGoRoutines)

	go func() {
		for _, q := range queries {
			queryChannel <- q
		}
		close(queryChannel)
	}()

	resultsChannel := make(chan Metric, m.maxGoRoutines)
	errorChannel := make(chan error, m.maxGoRoutines)

	for gr := 0; gr < m.maxGoRoutines; gr++ {
		log.Debugf("Starting getMetrics goroutine")
		go func() {
			for q := range queryChannel {
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
	log.Debugf("waiting for getMetrics goroutines to finish")
	wg.Wait()
	log.Debugf("all getMetrics goroutines finished")

	close(resultsChannel)
	close(errorChannel)
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
