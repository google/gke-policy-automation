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
	serviceMetricName = "service"
	serviceMetricHelp = "Number of services in a cluster split by namespace and type"
)

type serviceMetric struct {
	metric *prometheus.GaugeVec
	cache  map[string]int
	mutex  sync.Mutex
}

func NewServiceMetric() Metric {
	return &serviceMetric{
		metric: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: serviceMetricName,
				Help: serviceMetricHelp,
			},
			[]string{"namespace", "type"},
		),
		cache: map[string]int{},
	}
}

func (m *serviceMetric) GetName() string {
	return podMetricName
}

func (m *serviceMetric) OnAdd(obj interface{}) {
	m.updateMetric(obj, opAdd)
}

func (m *serviceMetric) OnUpdate(oldObj, newObj interface{}) {
}

func (m *serviceMetric) OnDelete(obj interface{}) {
	m.updateMetric(obj, opDelete)
}

func (m *serviceMetric) updateMetric(obj interface{}, op int) {
	service, ok := obj.(*v1.Service)
	if !ok {
		log.Warnf("got object of type %s, want *v1.Service", reflect.TypeOf(obj).Name())
		return
	}
	key := getServiceCacheKey(service)
	m.mutex.Lock()
	var newValue int
	if op == opAdd {
		newValue = m.cache[key] + 1
	} else {
		newValue = m.cache[key] - 1
	}
	log.Debugf("setting service metric for key %s to %d", key, newValue)
	m.cache[key] = newValue
	m.mutex.Unlock()
	m.metric.With(prometheus.Labels{"namespace": service.Namespace, "type": string(service.Spec.Type)}).Set(float64(newValue))
}

func getServiceCacheKey(service *v1.Service) string {
	return fmt.Sprintf("%s%c%s", service.Namespace, cacheDelimiter, string(service.Spec.Type))
}
