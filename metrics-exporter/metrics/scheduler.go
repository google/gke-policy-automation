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

// Package metrics implements custom metrics created from kubernetes cluster data
package metrics

import (
	"context"
	"time"

	"github.com/google/gke-policy-automation/metrics-exporter/log"
)

type Metric interface {
	GetName() string
	Update()
}

type scheduler struct {
	ctx      context.Context
	interval time.Duration
	metrics  []Metric
}

func NewScheduler(ctx context.Context, interval time.Duration) *scheduler {
	return &scheduler{
		ctx:      ctx,
		interval: interval,
	}
}

func (s *scheduler) WithMetric(m Metric) *scheduler {
	s.metrics = append(s.metrics, m)
	return s
}

func (s *scheduler) Start() {
	t := time.NewTicker(s.interval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			log.Infof("updating metrics")
			for _, m := range s.metrics {
				m.Update()
			}
		case <-s.ctx.Done():
			return
		}
	}
}
