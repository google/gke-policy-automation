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
