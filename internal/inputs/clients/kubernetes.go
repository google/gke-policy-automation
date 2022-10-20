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

package clients

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/gke-policy-automation/internal/log"
	"github.com/google/gke-policy-automation/internal/version"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type KubernetesClient interface {
	GetNamespaces() ([]string, error)
	GetFetchableResourceTypes() ([]*ResourceType, error)
	GetResources(toBeFetched []*ResourceType, namespaces []string) ([]*Resource, error)
	GetNamespacedResources(resourceType ResourceType, namespace string) ([]*Resource, error)
}

type KubernetesDiscoveryClient interface {
	ServerGroupsAndResources() ([]*metav1.APIGroup, []*metav1.APIResourceList, error)
}

type kubernetesClient struct {
	ctx           context.Context
	client        dynamic.Interface
	discovery     KubernetesDiscoveryClient
	maxGoroutines int
	maxQPS        int
	timeout       int
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

const (
	defaultMaxGoroutines    = 10
	defaultClientQPS        = 20
	defaultClientTimeoutSec = 3
)

type kubernetesClientBuilder struct {
	ctx           context.Context
	kubeConfig    *clientcmdapi.Config
	maxGoroutines int
	maxQPS        int
	timeout       int
}

func NewKubernetesClientBuilder(ctx context.Context, kubeConfig *clientcmdapi.Config) *kubernetesClientBuilder {
	return &kubernetesClientBuilder{
		ctx:        ctx,
		kubeConfig: kubeConfig,
	}
}

func (b *kubernetesClientBuilder) WithMaxGoroutines(maxGoroutines int) *kubernetesClientBuilder {
	b.maxGoroutines = maxGoroutines
	return b
}

func (b *kubernetesClientBuilder) WithMaxQPS(maxQPS int) *kubernetesClientBuilder {
	b.maxQPS = maxQPS
	return b
}

func (b *kubernetesClientBuilder) WithTimeout(timeout int) *kubernetesClientBuilder {
	b.timeout = timeout
	return b
}

func (b *kubernetesClientBuilder) Build() (KubernetesClient, error) {
	kubernetesClient := &kubernetesClient{
		ctx:           b.ctx,
		maxQPS:        defaultClientQPS,
		timeout:       defaultClientTimeoutSec,
		maxGoroutines: defaultMaxGoroutines,
	}
	if b.maxGoroutines != 0 {
		kubernetesClient.maxGoroutines = b.maxGoroutines
	}
	if b.maxQPS != 0 {
		kubernetesClient.maxQPS = b.maxQPS
	}
	if b.timeout != 0 {
		kubernetesClient.timeout = b.timeout
	}
	config, err := kubernetesClient.getKubernetesRestClientConfig(b.kubeConfig)
	if err != nil {
		return nil, err
	}
	kubernetesClient.client, err = dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	kubernetesClient.discovery, err = discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}
	return kubernetesClient, nil
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
			if !StringSliceContains(resourceGroup.APIResources[i].Verbs, "get") {
				log.Debugf("skipping resource type %q with groupVersion %q as it has no \"get\" verb",
					resourceGroup.APIResources[i].Name, resourceGroup.GroupVersion)
				continue
			}
			if strings.Contains(resourceGroup.APIResources[i].Name, "/") {
				log.Debugf("skipping resource type %q with groupVersion %q",
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

func (c *kubernetesClient) GetResources(toBeFetched []*ResourceType, namespaces []string) ([]*Resource, error) {
	var resources []*Resource

	namespaceCounter := len(namespaces)
	namespaceChannel := make(chan string, namespaceCounter)

	for _, ns := range namespaces {
		namespaceChannel <- ns
	}
	close(namespaceChannel)
	wg := new(sync.WaitGroup)
	wg.Add(c.maxGoroutines)

	resultsChannel := make(chan []*Resource, namespaceCounter)
	errorsChannel := make(chan error, namespaceCounter)

	for gr := 0; gr < c.maxGoroutines; gr++ {
		log.Debugf("Starting fetchNamespace goroutine")
		go c.getNamespaceResourcesByResourceTypeAsync(wg, toBeFetched, namespaceChannel, resultsChannel, errorsChannel)
	}
	log.Debugf("waiting for fetchNamespace goroutines to finish")
	wg.Wait()
	log.Debugf("all fetchNamespace goroutines finished")

	close(resultsChannel)
	close(errorsChannel)
	if len(errorsChannel) > 0 {
		err := <-errorsChannel
		log.Errorf("unable to get resources: %s", err)
		return nil, err
	}
	for result := range resultsChannel {
		resources = append(resources, result...)
	}
	return resources, nil
}

func (c *kubernetesClient) getNamespaceResourcesByResourceTypeAsync(wg *sync.WaitGroup, toBeFetched []*ResourceType, namespaces <-chan string, results chan<- []*Resource, errors chan<- error) {
	for namespace := range namespaces {
		var namespaceResources []*Resource = make([]*Resource, 0)

		for rt := range toBeFetched {
			res, err := c.GetNamespacedResources(*toBeFetched[rt], namespace)
			namespaceResources = append(namespaceResources, res...)
			if err != nil {
				log.Errorf("unable to get namespace resources: %s", err)
				errors <- err
				wg.Done()
				return
			}
		}
		results <- namespaceResources
		log.Debugf("fetchNamespace goroutine for namespace: %s finished", namespace)
	}
	wg.Done()
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

func (c *kubernetesClient) getKubernetesRestClientConfig(kubeConfig *clientcmdapi.Config) (*restclient.Config, error) {
	config, err := clientcmd.BuildConfigFromKubeconfigGetter("", func() (*clientcmdapi.Config, error) {
		return kubeConfig, nil
	})
	if err != nil {
		return nil, err
	}
	config.QPS = float32(c.maxQPS)
	config.UserAgent = version.UserAgent
	config.Timeout = time.Duration(c.timeout) * time.Second
	return config, nil
}

func StringSliceContains(hay []string, needle string) bool {
	for _, v := range hay {
		if v == needle {
			return true
		}
	}
	return false
}
