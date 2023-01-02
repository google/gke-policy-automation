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

package metrics

import (
	"github.com/google/gke-policy-automation/metrics-exporter/k8s"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	podMetricName = "pods"
	podMetricHelp = "Number of PODs in a cluster split by node and namespace"
)

type podMetric struct {
	informer *k8s.PodInformer
	metric   *prometheus.GaugeVec
}

func NewPodMetric(i *k8s.PodInformer) *podMetric {
	return &podMetric{
		informer: i,
		metric: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: podMetricName,
				Help: podMetricHelp,
			},
			[]string{"node", "namespace"},
		),
	}
}

func (m *podMetric) GetName() string {
	return podMetricName
}

func (m *podMetric) Update() {
	count := make(map[string]map[string]int)
	for _, p := range m.informer.GetPods() {
		nsCount, ok := count[p.Namespace]
		if !ok {
			nsCount = make(map[string]int)
			count[p.Namespace] = nsCount
		}
		nsCount[p.Spec.NodeName] = nsCount[p.Spec.NodeName] + 1
	}
	for ns := range count {
		for node, val := range count[ns] {
			m.metric.With(prometheus.Labels{"node": node, "namespace": ns}).Set(float64(val))
		}
	}
}
