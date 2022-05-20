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
	"fmt"

	"github.com/google/gke-policy-automation/internal/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type KubernetesClient interface {
	GetNamespaces() ([]string, error)
	GetFetchableResourceTypes() ([]*ResourceType, error)
	GetNamespacedResources(resourceType ResourceType, namespace string) ([]*Resource, error)
}

type KubernetesDiscoveryClient interface {
	ServerGroupsAndResources() ([]*metav1.APIGroup, []*metav1.APIResourceList, error)
}

type kubernetesClient struct {
	ctx       context.Context
	client    dynamic.Interface
	discovery KubernetesDiscoveryClient
}

type ResourceType struct {
	Group      string
	Version    string
	Name       string
	Namespaced bool
}

func (r ResourceType) toGroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    r.Group,
		Version:  r.Version,
		Resource: r.Name,
	}
}

type Resource struct {
	Type ResourceType
	Data map[string]interface{}
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
	log.Infof("fetching namespaces")
	list, err := c.client.Resource(namespaceRes).List(c.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	namespaces := make([]string, len(list.Items))
	for i := range list.Items {
		namespaces[i] = list.Items[i].GetName()
	}
	log.Infof("fetched %d namespaces", len(namespaces))
	return namespaces, nil
}

func (c *kubernetesClient) GetFetchableResourceTypes() ([]*ResourceType, error) {
	log.Infof("discovering server resource types")
	_, resourceGroupList, err := c.discovery.ServerGroupsAndResources()
	if err != nil {
		return nil, err
	}
	resourceTypes := make([]*ResourceType, 0)
	for _, resourceGroup := range resourceGroupList {
		resGroupVersion, err := schema.ParseGroupVersion(resourceGroup.GroupVersion)
		if err != nil {
			return nil, err
		}
		for i := range resourceGroup.APIResources {
			if !stringSliceContains(resourceGroup.APIResources[i].Verbs, "get") {
				log.Debugf("skipping resource type %q with groupVersion %q as it has no \"get\" verb",
					resourceGroup.APIResources[i].Name, resourceGroup.GroupVersion)
				continue
			}
			resourceTypes = append(resourceTypes, &ResourceType{
				Group:      resGroupVersion.Group,
				Version:    resGroupVersion.Version,
				Name:       resourceGroup.APIResources[i].Name,
				Namespaced: resourceGroup.APIResources[i].Namespaced})
		}
	}
	log.Infof("discovered %d fetchable resource types", len(resourceTypes))
	return resourceTypes, nil
}

func (c *kubernetesClient) GetNamespacedResources(resourceType ResourceType, namespace string) ([]*Resource, error) {
	if !resourceType.Namespaced {
		return nil, fmt.Errorf("resource type is not namespaced")
	}
	resourceList, err := c.client.Resource(resourceType.toGroupVersionResource()).Namespace(namespace).List(c.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	results := make([]*Resource, len(resourceList.Items))
	for i := range resourceList.Items {
		results[i] = &Resource{
			Type: resourceType,
			Data: resourceList.Items[i].Object,
		}
	}
	return results, nil
}

func stringSliceContains(hay []string, needle string) bool {
	for _, v := range hay {
		if v == needle {
			return true
		}
	}
	return false
}
