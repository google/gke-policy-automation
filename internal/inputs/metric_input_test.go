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
	"reflect"
	"testing"

	"github.com/google/gke-policy-automation/internal/inputs/clients"
)

type metricsClientMock struct {
}

func (metricsClientMock) GetMetric(query clients.MetricQuery, clusterName string) (*clients.Metric, error) {
	return &clients.Metric{Name: "test", ScalarValue: float64(123)}, nil
}

func (metricsClientMock) GetMetricsForCluster(queries []clients.MetricQuery, clusterName string) (map[string]clients.Metric, error) {
	m := make(map[string]clients.Metric)

	m["numberOfNodes"] = clients.Metric{
		Name:        "numberOfNodes",
		ScalarValue: float64(123),
	}
	m["numberOfPods"] = clients.Metric{
		Name:        "numberOfPods",
		ScalarValue: float64(321),
	}

	return m, nil
}

func TestMetricsInputBuilder(t *testing.T) {

	testCredsFile := "test-fixtures/test_credentials.json"
	queries := make([]clients.MetricQuery, 2)
	numberOfGoroutines := 5
	projectID := "testProject"

	queries[0] = clients.MetricQuery{
		Name:  "numberOfNodes",
		Query: "apiserver_storage_objects{resource=\"nodes\"}",
	}
	queries[1] = clients.MetricQuery{
		Name:  "numberOfPods",
		Query: "apiserver_storage_objects{resource=\"pods\"}",
	}

	b := NewMetricsInputBuilder(context.Background(), queries).
		WithCredentialsFile(testCredsFile).
		WithMaxGoroutines(numberOfGoroutines).
		WithProjectID(projectID)

	input, err := b.Build()
	if err != nil {
		t.Fatalf("err = %v; want nil", err)
	}
	metricsInput, ok := input.(*metricsInput)
	if b.credentialsFile != testCredsFile {
		t.Errorf("builder credentialsFile = %v; want %v", b.credentialsFile, testCredsFile)
	}
	if !ok {
		t.Fatalf("input is not *metricsInput")
	}
	if metricsInput.maxGoRoutines != numberOfGoroutines {
		t.Errorf("maxGoRoutines = %v; want %v", metricsInput.maxGoRoutines, numberOfGoroutines)
	}
	if !reflect.DeepEqual(metricsInput.queries, queries) {
		t.Errorf("queries = %v; want %v", metricsInput.queries, queries)
	}
	if metricsInput.projectID != projectID {
		t.Errorf("projectId = %v; want %v", metricsInput.projectID, projectID)
	}
}

func TestMetricsInputGetID(t *testing.T) {
	input := metricsInput{}
	if id := input.GetID(); id != metricsInputID {
		t.Fatalf("id = %v; want %v", id, metricsInputID)
	}
}

func TestMetricsInputGetDescription(t *testing.T) {
	input := metricsInput{}
	if id := input.GetDescription(); id != metricsInputDescription {
		t.Fatalf("id = %v; want %v", id, metricsInputDescription)
	}
}

func TestMetricsInputGetData(t *testing.T) {
	queries := make([]clients.MetricQuery, 2)
	queries[0] = clients.MetricQuery{
		Name:  "numberOfNodes",
		Query: "apiserver_storage_objects{resource=\"nodes\"}",
	}
	queries[1] = clients.MetricQuery{
		Name:  "numberOfPods",
		Query: "apiserver_storage_objects{resource=\"pods\"}",
	}

	projectID := "testProject"
	clusterID := "projects/testProject/locations/us/clusters/cluster1"

	input := metricsInput{
		ctx:           context.Background(),
		projectID:     projectID,
		queries:       queries,
		metricsClient: metricsClientMock{},
	}

	_, err := input.GetData(clusterID)
	if err != nil {
		t.Fatalf("err = %v; want nil", err)
	}
}
