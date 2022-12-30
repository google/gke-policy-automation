package metrics

import (
	"context"
	"time"

	"github.com/google/gke-policy-automation/internal/log"
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
