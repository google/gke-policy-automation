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
	"log"
	"net/http"
	"time"

	//monitoring "cloud.google.com/go/monitoring/apiv3/v2"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/config"
)

type MetricQuery struct {
	Name  string
	Query string
}

type Metric struct {
	Name  string
	Value string //int?
}

type MetricsClient interface {
	GetMetric(query string, clusterName string) (string, error)
}

type metricsClient struct {
	ctx    context.Context
	client api.Client
}

func NewMetricClient(ctx context.Context, projectId string, authToken string) (MetricsClient, error) {

	// Creates a client.
	client, err := api.NewClient(api.Config{
		Address:      "https://monitoring.googleapis.com/v1/projects/" + projectId + "/location/global/prometheus/",
		RoundTripper: config.NewAuthorizationCredentialsRoundTripper("Bearer", config.Secret(authToken), api.DefaultRoundTripper),
	})

	if err != nil {
		log.Fatalf("Failed to create metrics client: %v", err)
		return nil, err
	}

	return &metricsClient{
		ctx:    ctx,
		client: client,
	}, nil
}

func (m *metricsClient) GetMetric(query string, clusterName string) (string, error) {

	v1api := v1.NewAPI(m.client)
	//ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	//defer cancel()
	result, warnings, err := v1api.Query(context.Background(), query, time.Now())
	if err != nil {
		log.Fatalf("Failed to query metrics client: %v", err)
		return "", err
	}
	if warnings != nil {
		log.Fatalf("Warning when querying metrics client: %v", warnings)
	}
	return result.String(), nil
}

type authRoundTripper struct {
	next http.RoundTripper
}

func (a authRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	//auth
	return a.next.RoundTrip(r)
}
