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
	"testing"
	"time"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	pmodel "github.com/prometheus/common/model"
)

type metricsAPIClientMock struct {
	QueryFn func(ctx context.Context, query string, ts time.Time, opts ...v1.Option) (pmodel.Value, v1.Warnings, error)
}

func (m *metricsAPIClientMock) Query(ctx context.Context, query string, ts time.Time, opts ...v1.Option) (pmodel.Value, v1.Warnings, error) {
	return m.QueryFn(ctx, query, ts, opts...)
}

func (m *metricsAPIClientMock) Alerts(ctx context.Context) (v1.AlertsResult, error) {
	return v1.AlertsResult{}, nil
}

func (m *metricsAPIClientMock) AlertManagers(ctx context.Context) (v1.AlertManagersResult, error) {
	return v1.AlertManagersResult{}, nil
}

func (m *metricsAPIClientMock) CleanTombstones(ctx context.Context) error {
	return nil
}

func (m *metricsAPIClientMock) Config(ctx context.Context) (v1.ConfigResult, error) {
	return v1.ConfigResult{}, nil
}

func (m *metricsAPIClientMock) DeleteSeries(ctx context.Context, matches []string, startTime, endTime time.Time) error {
	return nil
}

func (m *metricsAPIClientMock) Flags(ctx context.Context) (v1.FlagsResult, error) {
	return nil, nil
}

func (m *metricsAPIClientMock) LabelNames(ctx context.Context, matches []string, startTime, endTime time.Time) ([]string, v1.Warnings, error) {
	return nil, nil, nil
}

func (m *metricsAPIClientMock) LabelValues(ctx context.Context, label string, matches []string, startTime, endTime time.Time) (pmodel.LabelValues, v1.Warnings, error) {
	return nil, nil, nil
}

func (m *metricsAPIClientMock) QueryRange(ctx context.Context, query string, r v1.Range, opts ...v1.Option) (pmodel.Value, v1.Warnings, error) {
	return nil, nil, nil
}

func (m *metricsAPIClientMock) QueryExemplars(ctx context.Context, query string, startTime, endTime time.Time) ([]v1.ExemplarQueryResult, error) {
	return nil, nil
}

func (m *metricsAPIClientMock) Buildinfo(ctx context.Context) (v1.BuildinfoResult, error) {
	return v1.BuildinfoResult{}, nil
}

func (m *metricsAPIClientMock) Runtimeinfo(ctx context.Context) (v1.RuntimeinfoResult, error) {
	return v1.RuntimeinfoResult{}, nil
}

func (m *metricsAPIClientMock) Series(ctx context.Context, matches []string, startTime, endTime time.Time) ([]pmodel.LabelSet, v1.Warnings, error) {
	return nil, nil, nil
}

func (m *metricsAPIClientMock) Snapshot(ctx context.Context, skipHead bool) (v1.SnapshotResult, error) {
	return v1.SnapshotResult{}, nil
}

func (m *metricsAPIClientMock) Rules(ctx context.Context) (v1.RulesResult, error) {
	return v1.RulesResult{}, nil
}

func (m *metricsAPIClientMock) Targets(ctx context.Context) (v1.TargetsResult, error) {
	return v1.TargetsResult{}, nil
}

func (m *metricsAPIClientMock) TargetsMetadata(ctx context.Context, matchTarget, metric, limit string) ([]v1.MetricMetadata, error) {
	return nil, nil
}

func (m *metricsAPIClientMock) Metadata(ctx context.Context, metric, limit string) (map[string][]v1.Metadata, error) {
	return nil, nil
}

func (m *metricsAPIClientMock) TSDB(ctx context.Context) (v1.TSDBResult, error) {
	return v1.TSDBResult{}, nil
}

func (m *metricsAPIClientMock) WalReplay(ctx context.Context) (v1.WalReplayStatus, error) {
	return v1.WalReplayStatus{}, nil
}

func TestNewMetricClient(t *testing.T) {
	ctx := context.TODO()
	cli, err := newMetricsClient(ctx, "test-project", "fake-token", 20)
	if err != nil {
		t.Fatalf("err is not nil; want nil; err = %s", err)
	}
	realCli, ok := cli.(*metricsClient)
	if !ok {
		t.Fatalf("cli is not *gke.metricsClient")
	}
	if realCli.ctx != ctx {
		t.Errorf("context is %v; want %v", realCli.ctx, ctx)
	}
	if realCli.client == nil {
		t.Errorf("client is nil; want api.Client")
	}
}

func TestGetMetric(t *testing.T) {

	metricName := "test-metric"
	metricValue := 22
	testQuery := "test-query"

	v1ApiMock := &metricsAPIClientMock{
		QueryFn: func(ctx context.Context, query string, ts time.Time, opts ...v1.Option) (pmodel.Value, v1.Warnings, error) {

			if query != testQuery {
				t.Errorf("query is %v; want %v", query, testQuery)
			}

			return pmodel.Vector{
				&pmodel.Sample{
					Metric:    nil,
					Value:     pmodel.SampleValue(metricValue),
					Timestamp: pmodel.Now(),
				},
			}, nil, nil
		},
	}

	client := &metricsClient{ctx: context.TODO(), client: nil, api: v1ApiMock, maxGoRoutines: defaultMaxGoroutines}

	result, err := client.GetMetric(MetricQuery{Query: testQuery, Name: metricName}, "sample-cluster")

	if err != nil {
		t.Fatalf("err is not nil; want nil; err = %s", err)
	}
	if result != fmt.Sprint(metricValue) {
		t.Errorf("result is %v; want %v", result, fmt.Sprint(metricValue))
	}
}

func TestGetMetricsForCluster(t *testing.T) {

	metric1Name := "test-metric"
	metricValue := 22

	metric2Name := "test-metric2"

	v1ApiMock := &metricsAPIClientMock{
		QueryFn: func(ctx context.Context, query string, ts time.Time, opts ...v1.Option) (pmodel.Value, v1.Warnings, error) {

			return pmodel.Vector{
				&pmodel.Sample{
					Metric:    nil,
					Value:     pmodel.SampleValue(metricValue),
					Timestamp: pmodel.Now(),
				},
			}, nil, nil
		},
	}

	client := &metricsClient{ctx: context.TODO(), client: nil, api: v1ApiMock, maxGoRoutines: defaultMaxGoroutines}

	result, err := client.GetMetricsForCluster([]MetricQuery{{Query: "test-query", Name: metric1Name}, {Query: "test-query", Name: metric2Name}}, "sample-cluster")

	if err != nil {
		t.Fatalf("err is not nil; want nil; err = %s", err)
	}
	if len(result) != 2 {
		t.Errorf("result len is %v; want %v", len(result), 2)
	}
}

func TestReplaceWildcard(t *testing.T) {
	query := "apiserver_storage_objects{resource=\"pods\", cluster=CLUSTER_NAME}"
	clusterName := "test_cluster"

	expected := "apiserver_storage_objects{resource=\"pods\", cluster=\"test_cluster\"}"

	result := replaceWildcard("CLUSTER_NAME", clusterName, query)

	if result != expected {
		t.Errorf("result query is %v; want %v", result, expected)
	}

}
