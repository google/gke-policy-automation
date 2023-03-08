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
	"net/http"
	"reflect"
	"regexp"
	"sort"
	"sync"
	"time"

	"github.com/google/gke-policy-automation/internal/gke"
	"github.com/google/gke-policy-automation/internal/log"
	"github.com/google/gke-policy-automation/internal/version"

	b64 "encoding/base64"

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

type emptyResultError struct {
	msg string
}

func (e emptyResultError) Error() string {
	return e.msg
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

func newMetricsClient(ctx context.Context, address string, roundTripper http.RoundTripper, maxGoroutines int) (MetricsClient, error) {
	// Creates a client.
	client, err := api.NewClient(api.Config{
		Address:      address,
		RoundTripper: roundTripper,
	})

	if err != nil {
		log.Fatalf("failed to create metrics client: %v", err)
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
	tokenSource   TokenSource
	maxGoroutines int
	timeout       int
	address       string
	username      string
	password      string
}

func NewMetricsClientBuilder(ctx context.Context) *metricsClientBuilder {
	return &metricsClientBuilder{
		ctx: ctx,
	}
}

func (b *metricsClientBuilder) WithGoogleCloudMonitoring(projectID string, tokenSource TokenSource) *metricsClientBuilder {
	b.projectID = projectID
	b.tokenSource = tokenSource
	return b
}

func (b *metricsClientBuilder) WithAddress(address string) *metricsClientBuilder {
	b.address = address
	return b
}

func (b *metricsClientBuilder) WithUsernamePassword(username string, password string) *metricsClientBuilder {
	b.username = username
	b.password = password
	return b
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
	maxGoRoutines := b.maxGoroutines
	if b.maxGoroutines == 0 {
		maxGoRoutines = defaultMaxGoroutines
	}

	address := b.address
	if address == "" {
		address = fmt.Sprintf("https://monitoring.googleapis.com/v1/projects/%s/location/global/prometheus/", b.projectID)
	}

	roundTripper, err := getRoundTripper(b.tokenSource, b.username, b.password)
	if err != nil {
		return nil, err
	}
	return newMetricsClient(b.ctx, address, roundTripper, maxGoRoutines)
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
		return nil, &emptyResultError{
			msg: fmt.Sprintf("metric query %q returned no results", query),
		}
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
		if len(metricValues) < 1 {
			return nil, fmt.Errorf("metric query result has no labels")
		}
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

	for err := range errorChannel {
		switch err.(type) {
		case *emptyResultError:
			log.Warnf("metric fetch error: %s", err)
		default:
			return nil, err
		}
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
	if len(labels) < 1 {
		return
	}
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

func getRoundTripper(ts TokenSource, username, password string) (http.RoundTripper, error) {
	if ts != nil {
		authToken, err := ts.GetAuthToken()
		if err != nil {
			return nil, err
		}
		return config.NewAuthorizationCredentialsRoundTripper("Bearer", config.Secret(authToken), getDefaultRoundTripper()), nil
	}
	if username != "" && password != "" {
		secret := b64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password)))
		return config.NewAuthorizationCredentialsRoundTripper("Basic", config.Secret(secret), getDefaultRoundTripper()), nil
	}
	return getDefaultRoundTripper(), nil
}

type metricsRoundTripper struct {
	rt http.RoundTripper
}

func (r metricsRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", version.UserAgent)
	return r.rt.RoundTrip(req)
}

func getDefaultRoundTripper() http.RoundTripper {
	return &metricsRoundTripper{rt: api.DefaultRoundTripper}
}
