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

// Package k8s implements kubernetes related functions like clients and informers
package k8s

import (
	"context"
	"encoding/base64"
	"fmt"

	"cloud.google.com/go/container/apiv1/containerpb"
	"github.com/google/gke-policy-automation/metrics-exporter/gke"
	"github.com/google/gke-policy-automation/metrics-exporter/log"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func NewInClusterClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

func NewClientFromGKE(ctx context.Context, clusterID string) (*kubernetes.Clientset, error) {
	ts, err := gke.NewGoogleTokenSource(ctx)
	if err != nil {
		return nil, err
	}
	token, err := ts.GetAuthToken()
	if err != nil {
		return nil, err
	}
	gkeClient, err := gke.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	defer gkeClient.Close()
	gkeData, err := gkeClient.GetData(clusterID)
	if err != nil {
		return nil, err
	}
	kubeConfig, err := createKubeConfig(gkeData, token)
	if err != nil {
		return nil, err
	}
	config, _ := clientcmd.BuildConfigFromKubeconfigGetter("", func() (*clientcmdapi.Config, error) {
		return kubeConfig, nil
	})
	return kubernetes.NewForConfig(config)
}

func createKubeConfig(clusterData *containerpb.Cluster, clusterToken string) (*clientcmdapi.Config, error) {
	k8sKubeConfigContextName := "gke"
	clusterMasterAuth := clusterData.MasterAuth.ClusterCaCertificate
	clusterEndpoint := clusterData.Endpoint
	config := clientcmdapi.NewConfig()

	caCert, err := base64.StdEncoding.DecodeString(clusterMasterAuth)
	if err != nil {
		log.Debugf("Unable to retrieve clusterMasterAuth %s:", err)
		return nil, err
	}
	config.APIVersion = "v1"
	config.Kind = "Config"
	config.Clusters = map[string]*clientcmdapi.Cluster{
		k8sKubeConfigContextName: {
			CertificateAuthorityData: caCert,
			Server:                   fmt.Sprintf("https://%v", clusterEndpoint),
		},
	}
	config.AuthInfos = map[string]*clientcmdapi.AuthInfo{
		k8sKubeConfigContextName: {Token: clusterToken},
	}
	config.Contexts = map[string]*clientcmdapi.Context{
		k8sKubeConfigContextName: {
			Cluster:  k8sKubeConfigContextName,
			AuthInfo: k8sKubeConfigContextName,
		},
	}
	config.CurrentContext = k8sKubeConfigContextName
	log.Info("Local kubernetes cluster configuration created")
	return config, nil
}
