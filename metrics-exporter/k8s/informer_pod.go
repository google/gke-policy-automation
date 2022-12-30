package k8s

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
)

type PodInformer struct {
	informer cache.SharedInformer
}

func (i *PodInformer) Run(stopCh <-chan struct{}) {
	i.informer.Run(stopCh)
}

func (i *PodInformer) GetPods() []*v1.Pod {
	pods := []*v1.Pod{}
	for _, p := range i.informer.GetStore().List() {
		p := p.(*v1.Pod)
		pods = append(pods, p)
	}
	return pods
}
