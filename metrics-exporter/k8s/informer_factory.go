package k8s

import (
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
)

type informerFactory struct {
	factory informers.SharedInformerFactory
}

func NewInformerFactory(kubeClient kubernetes.Interface) *informerFactory {
	return &informerFactory{
		factory: informers.NewSharedInformerFactory(kubeClient, 0),
	}
}

func (f *informerFactory) GetPodInformer() *PodInformer {
	informer := f.factory.Core().V1().Pods().Informer()
	return &PodInformer{
		informer: informer,
	}
}
