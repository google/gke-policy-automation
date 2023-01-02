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

package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/gke-policy-automation/metrics-exporter/k8s"
	"github.com/google/gke-policy-automation/metrics-exporter/log"
	"github.com/google/gke-policy-automation/metrics-exporter/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/client-go/kubernetes"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("\nError: %s\n", err)
			os.Exit(1)
		}
	}()

	log.Info("GKE Policy Automation metrics exporter starting")
	ctx, cancel := context.WithCancel(context.Background())
	config := NewConfigFromEnv()
	var kClient *kubernetes.Clientset
	var err error
	if config.KubeConfigGKE != "" {
		log.Debugf("Creating kube client: GKE client for cluster %s", config.KubeConfigGKE)
		kClient, err = k8s.NewClientFromGKE(ctx, config.KubeConfigGKE)
	} else {
		log.Debug("Creating kube client: in cluster client")
		kClient, err = k8s.NewInClusterClient()
	}
	if err != nil {
		log.Fatalf("Could not create kubernetes client: %s", err)
		os.Exit(1)
	}

	informers := k8s.NewInformerFactory(kClient)
	podInformer := informers.GetPodInformer()

	go k8s.NewClusterWatcher(ctx).
		WithInformer(podInformer).Start()

	go metrics.NewScheduler(ctx, time.Duration(1*time.Minute)).
		WithMetric(metrics.NewPodMetric(podInformer)).
		Start()

	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(":8080", nil)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	cancel()
	log.Infof("GKE Policy Automation metrics exporter exiting")
}
