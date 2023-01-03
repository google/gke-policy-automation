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
	nodeMetricName    = "nodes"
	nodeMetricHelp    = "Number of nodes in a cluster split by zone, node-pool"
	nodeLabelZone     = "topology.kubernetes.io/zone"
	nodeLabelNodePool = "cloud.google.com/gke-nodepool"
)

type nodeMetric struct {
	metric *prometheus.GaugeVec
	cache  map[string]int
	mutex  sync.Mutex
}

func NewNodeMetric() Metric {
	return &nodeMetric{
		metric: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: nodeMetricName,
				Help: nodeMetricHelp,
			},
			[]string{"zone", "nodepool"},
		),
		cache: map[string]int{},
	}
}

func (m *nodeMetric) GetName() string {
	return podMetricName
}

func (m *nodeMetric) OnAdd(obj interface{}) {
	m.updateMetric(obj, opAdd)
}

func (m *nodeMetric) OnUpdate(oldObj, newObj interface{}) {
}

func (m *nodeMetric) OnDelete(obj interface{}) {
	m.updateMetric(obj, opDelete)
}

func (m *nodeMetric) updateMetric(obj interface{}, op int) {
	node, ok := obj.(*v1.Node)
	if !ok {
		log.Warnf("got object of type %s, want *v1.Node", reflect.TypeOf(obj).Name())
		return
	}
	if err := validNodeLabels(node); err != nil {
		log.Warnf("invalid node spec: %s", err)
		return
	}
	key := getNodeCacheKey(node)
	m.mutex.Lock()
	var newValue int
	if op == opAdd {
		newValue = m.cache[key] + 1
	} else {
		newValue = m.cache[key] - 1
	}
	log.Debugf("setting node metric for key %s to %d", key, newValue)
	m.cache[key] = newValue
	m.mutex.Unlock()
	m.metric.With(prometheus.Labels{"zone": node.Labels[nodeLabelZone], "nodepool": node.Labels[nodeLabelNodePool]}).Set(float64(newValue))
}

func validNodeLabels(node *v1.Node) error {
	if _, ok := node.Labels[nodeLabelZone]; !ok {
		return fmt.Errorf("missing zone label %s", nodeLabelZone)
	}
	if _, ok := node.Labels[nodeLabelNodePool]; !ok {
		return fmt.Errorf("missing nodepool label %s", nodeLabelNodePool)
	}
	return nil
}

func getNodeCacheKey(node *v1.Node) string {
	return fmt.Sprintf("%s%c%s", node.Labels[nodeLabelZone], cacheDelimiter, node.Labels[nodeLabelNodePool])
}
