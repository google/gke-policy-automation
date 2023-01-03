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

package k8s

import (
	"github.com/google/gke-policy-automation/metrics-exporter/log"
	"github.com/google/gke-policy-automation/metrics-exporter/metrics"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type Informer interface {
	WithMetric(m metrics.Metric) Informer
	Run(stopCh <-chan struct{})
}

type genericInformer struct {
	informer cache.SharedInformer
}

type informerFactory struct {
	factory informers.SharedInformerFactory
}

func NewInformerFactory(kubeClient kubernetes.Interface) *informerFactory {
	return &informerFactory{
		factory: informers.NewSharedInformerFactory(kubeClient, 0),
	}
}

func (f *informerFactory) GetPodInformer() Informer {
	informer := f.factory.Core().V1().Pods().Informer()
	return &genericInformer{
		informer: informer,
	}
}

func (f *informerFactory) GetNodeInformer() Informer {
	informer := f.factory.Core().V1().Nodes().Informer()
	return &genericInformer{
		informer: informer,
	}
}

func (f *informerFactory) GetServiceInformer() Informer {
	informer := f.factory.Core().V1().Services().Informer()
	return &genericInformer{
		informer: informer,
	}
}

func (i *genericInformer) WithMetric(m metrics.Metric) Informer {
	_, err := i.informer.AddEventHandler(m)
	if err != nil {
		log.Errorf("failed to add new event handler for metric %s: %s", m.GetName(), err)
	}
	return i
}

func (i *genericInformer) Run(stopCh <-chan struct{}) {
	i.informer.Run(stopCh)
}
