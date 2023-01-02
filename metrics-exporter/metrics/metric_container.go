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
	containerMetricName     = "containers"
	containerMetricHelp     = "Number of containers in a cluster split by the node, namespace and pod"
	containerCacheDelimiter = '|'
)

type containerMetric struct {
	metric *prometheus.GaugeVec
	cache  map[string]int
	mutex  sync.Mutex
}

func NewContainerMetric() Metric {
	return &containerMetric{
		metric: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: containerMetricName,
				Help: containerMetricHelp,
			},
			[]string{"node", "namespace"},
		),
		cache: map[string]int{},
	}
}

func (m *containerMetric) GetName() string {
	return containerMetricName
}

func (m *containerMetric) OnAdd(obj interface{}) {
	m.updateMetric(obj, nil, opAdd)
}

func (m *containerMetric) OnUpdate(oldObj, newObj interface{}) {
	m.updateMetric(newObj, oldObj, opUpdate)
}

func (m *containerMetric) OnDelete(obj interface{}) {
	m.updateMetric(obj, nil, opDelete)
}

func (m *containerMetric) updateMetric(obj interface{}, oldObj interface{}, op int) {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		log.Warnf("got object of type %s, want *v1.Pod", reflect.TypeOf(obj).Name())
		return
	}
	key := getContainerCacheKey(pod)
	m.mutex.Lock()
	var newValue int
	switch op {
	case opAdd:
		newValue = m.cache[key] + len(pod.Spec.Containers)
	case opUpdate:
		oldPod, ok := oldObj.(*v1.Pod)
		if !ok {
			log.Warnf("got object of type %s, want *v1.Pod", reflect.TypeOf(obj).Name())
			break
		}
		delta := len(pod.Spec.Containers) - len(oldPod.Spec.Containers)
		newValue = m.cache[key] + delta
	case opDelete:
		newValue = m.cache[key] - len(pod.Spec.Containers)
	}
	log.Debugf("setting container metric for key %s to %d", key, newValue)
	m.cache[key] = newValue
	m.mutex.Unlock()
	m.metric.With(prometheus.Labels{"node": pod.Spec.NodeName, "namespace": pod.Namespace}).Set(float64(newValue))
}

func getContainerCacheKey(pod *v1.Pod) string {
	return fmt.Sprintf("%s%c%s", pod.Spec.NodeName, containerCacheDelimiter, pod.Namespace)
}
