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

import "context"

type Informer interface {
	Run(stopCh <-chan struct{})
}

type clusterWatcher struct {
	ctx       context.Context
	informers []Informer
}

func NewClusterWatcher(ctx context.Context) *clusterWatcher {
	return &clusterWatcher{
		ctx: ctx,
	}
}

func (w *clusterWatcher) WithInformer(i Informer) *clusterWatcher {
	w.informers = append(w.informers, i)
	return w
}

func (w *clusterWatcher) Start() {
	stop := make(chan struct{})
	for _, i := range w.informers {
		go i.Run(stop)
	}
	go func() {
		<-w.ctx.Done()
		close(stop)
	}()
}
