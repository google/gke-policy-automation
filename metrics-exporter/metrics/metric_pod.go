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
	"fmt"
	"reflect"
	"sync"

	"github.com/google/gke-policy-automation/metrics-exporter/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	v1 "k8s.io/api/core/v1"
)

const (
	podMetricName = "pods"
	podMetricHelp = "Number of PODs in a cluster split by node and namespace"
)

type podMetric struct {
	metric *prometheus.GaugeVec
	cache  map[string]int
	mutex  sync.Mutex
}

func NewPodMetric() Metric {
	return &podMetric{
		metric: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: podMetricName,
				Help: podMetricHelp,
			},
			[]string{"node", "namespace"},
		),
		cache: map[string]int{},
	}
}

func (m *podMetric) GetName() string {
	return podMetricName
}

func (m *podMetric) OnAdd(obj interface{}) {
	m.updateMetric(obj, opAdd)
}

func (m *podMetric) OnUpdate(oldObj, newObj interface{}) {
}

func (m *podMetric) OnDelete(obj interface{}) {
	m.updateMetric(obj, opDelete)
}

func (m *podMetric) updateMetric(obj interface{}, op int) {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		log.Warnf("got object of type %s, want *v1.Pod", reflect.TypeOf(obj).Name())
		return
	}
	key := getPodCacheKey(pod)
	m.mutex.Lock()
	var newValue int
	if op == opAdd {
		newValue = m.cache[key] + 1
	} else {
		newValue = m.cache[key] - 1
	}
	log.Debugf("setting pod metric for key %s to %d", key, newValue)
	m.cache[key] = newValue
	m.mutex.Unlock()
	m.metric.With(prometheus.Labels{"node": pod.Spec.NodeName, "namespace": pod.Namespace}).Set(float64(newValue))
}

func getPodCacheKey(pod *v1.Pod) string {
	return fmt.Sprintf("%s%c%s", pod.Spec.NodeName, cacheDelimiter, pod.Namespace)
}
