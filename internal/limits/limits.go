package main

import (
	"context"
	"fmt"

	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	ctrl "sigs.k8s.io/controller-runtime"
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

func main() {
	ctx := context.Background()
	config := ctrl.GetConfigOrDie()
	clientset := kubernetes.NewForConfigOrDie(config)

	namespace := "kube-system"
	items, err := GetDeployments(clientset, ctx, namespace)
	if err != nil {
		fmt.Println(err)
	} else {
		for _, item := range items {
			fmt.Printf("%+v\n", item)
		}
	}
}

func GetDeployments(clientset *kubernetes.Clientset, ctx context.Context,
	namespace string) ([]v1.Deployment, error) {

	list, err := clientset.AppsV1().Deployments(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func k8sStartingConfig() (*clientcmdapi.Config, error) {
	po := clientcmd.NewDefaultPathOptions()
	return po.GetStartingConfig()
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
