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
