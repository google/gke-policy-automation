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

package gke

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type KubernetesClient interface {
	GetNamespaces() ([]string, error)
}

type kubernetesClient struct {
	ctx       context.Context
	client    dynamic.Interface
	discovery *discovery.DiscoveryClient
}

func NewKubernetesClient(ctx context.Context, kubeConfig *clientcmdapi.Config) (KubernetesClient, error) {
	config, err := clientcmd.BuildConfigFromKubeconfigGetter("", func() (*clientcmdapi.Config, error) {
		return kubeConfig, nil
	})
	if err != nil {
		return nil, err
	}
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	discovery, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}
	return &kubernetesClient{
		ctx:       ctx,
		client:    client,
		discovery: discovery,
	}, nil
}

func (c *kubernetesClient) GetNamespaces() ([]string, error) {
	namespaceRes := schema.GroupVersionResource{
		Version:  "v1",
		Resource: "namespaces",
	}
	list, err := c.client.Resource(namespaceRes).List(c.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	namespaces := make([]string, len(list.Items))
	for i := range list.Items {
		namespaces[i] = list.Items[i].GetName()
	}

	return namespaces, nil
}
