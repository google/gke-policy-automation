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
	"reflect"
	"testing"

	b64 "encoding/base64"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type kubeDynamicClientMock struct {
	ResourceFn func(resource schema.GroupVersionResource) dynamic.NamespaceableResourceInterface
}

func (m *kubeDynamicClientMock) Resource(resource schema.GroupVersionResource) dynamic.NamespaceableResourceInterface {
	return m.ResourceFn(resource)
}

type kubeNamespacedResourceMock struct {
	NamespaceFn func(n string) dynamic.ResourceInterface
	ListFn      func(ctx context.Context, opts metav1.ListOptions) (*unstructured.UnstructuredList, error)
}

func (m *kubeNamespacedResourceMock) Namespace(n string) dynamic.ResourceInterface {
	return m.NamespaceFn(n)
}

func (m *kubeNamespacedResourceMock) Create(ctx context.Context, obj *unstructured.Unstructured, options metav1.CreateOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, nil
}

func (m *kubeNamespacedResourceMock) Update(ctx context.Context, obj *unstructured.Unstructured, options metav1.UpdateOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, nil
}

func (m *kubeNamespacedResourceMock) UpdateStatus(ctx context.Context, obj *unstructured.Unstructured, options metav1.UpdateOptions) (*unstructured.Unstructured, error) {
	return nil, nil
}

func (m *kubeNamespacedResourceMock) Delete(ctx context.Context, name string, options metav1.DeleteOptions, subresources ...string) error {
	return nil
}

func (m *kubeNamespacedResourceMock) DeleteCollection(ctx context.Context, options metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	return nil
}

func (m *kubeNamespacedResourceMock) Get(ctx context.Context, name string, options metav1.GetOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, nil
}

func (m *kubeNamespacedResourceMock) List(ctx context.Context, opts metav1.ListOptions) (*unstructured.UnstructuredList, error) {
	return m.ListFn(ctx, opts)
}

func (m *kubeNamespacedResourceMock) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return nil, nil
}

func (m *kubeNamespacedResourceMock) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, options metav1.PatchOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, nil
}

type dicoveryClientMock struct {
	ServerGroupsAndResourcesFn func() ([]*metav1.APIGroup, []*metav1.APIResourceList, error)
}

func (m dicoveryClientMock) ServerGroupsAndResources() ([]*metav1.APIGroup, []*metav1.APIResourceList, error) {
	return m.ServerGroupsAndResourcesFn()
}

func TestResourceTypeToGroupVersionResource(t *testing.T) {
	resType := ResourceType{
		Group:      "apps",
		Version:    "v1",
		Name:       "deployments",
		Namespaced: true,
	}

	grVerRes := resType.toGroupVersionResource()
	if grVerRes.Group != resType.Group {
		t.Errorf("group = %v; want %v", grVerRes.Group, resType.Group)
	}
	if grVerRes.Version != resType.Version {
		t.Errorf("version = %v; want %v", grVerRes.Version, resType.Version)
	}
	if grVerRes.Resource != resType.Name {
		t.Errorf("resource = %v; want %v", grVerRes.Resource, resType.Name)
	}
}

func TestNewKubernetesClient(t *testing.T) {
	cert, _ := b64.StdEncoding.DecodeString(`LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUVMRENDQXBTZ0F3SUJBZ0lRRThETy9sTllZaUM0SWNhNm04N1EyVEFOQmdrcWhraUc5dzBCQVFzRkFEQXYKTVMwd0t3WURWUVFERXlSbVpUQTFPREF4TkMweU1UUTFMVFF3TURjdFltVTFNUzB6TkRZMVptWTRNREExWm1JdwpJQmNOTWpJd05URTNNVEUwTmpReVdoZ1BNakExTWpBMU1Ea3hNalEyTkRKYU1DOHhMVEFyQmdOVkJBTVRKR1psCk1EVTRNREUwTFRJeE5EVXROREF3TnkxaVpUVXhMVE0wTmpWbVpqZ3dNRFZtWWpDQ0FhSXdEUVlKS29aSWh2Y04KQVFFQkJRQURnZ0dQQURDQ0FZb0NnZ0dCQUs3TGFyT1U5VXFmNkRvM2JMUDR3aHV1ZUs1Sjd5NCtzT3d1aW1KLwpBYk82cG1lRDk0OGJaOXJhUUUwb0trU2RsZFlROEFUY0FSN3p0RjNhQURpZkVUNGJsYzExQ3o5dEZkUE9iMU02CkxoeU92MkZnNjBuMVhneFdhU1NHMVhLdG11OTVhUWRZbmR6TGNnWFM2MXNXWkgrQk90ZktvRWU4bitGY21JczUKWWZha3Iwb3gvVmZKWXMvWEVEUEw1UUdmcEtVY21kd0FiamQycVJ6MEROUlZLVThyRUJSSVp3ODFVUEJiaVpYVgozNlRHMDZCVFBCNnM3cFJQODFBeG10T3Z1dGZIT1l0STBLNU50Z1Vhc2ZLVGV6aTVsYTB1dDFGWnJQcFZ3NmpUCm91RE4wY1hkMjlyTlhKZDg4L0RpdUN4eXNrYk5xU2JkajBVYTV0UUw1ejltbDU2V2ZCMEszc2dzN0l2N1VOUGQKRGFSN2lUZGxoZWFxUmhQRkV3eW56Szg5NmlmSkltTnowTDBKNTFHekR1TFZmRUtyMFl2VytqdW9MUDl1OWhxdwpXSngxaEpsblhVb2NmY25Va1JMb3Y5VThOem5HTjZIZjhKTFBiZlhMdEtoWWFRMjFCZC9XcFV4TUVnRUJRNWdTCk5TOFRaNTVQRy94Vm14YjZpYzE4QjZNYXdRSURBUUFCbzBJd1FEQU9CZ05WSFE4QkFmOEVCQU1DQWdRd0R3WUQKVlIwVEFRSC9CQVV3QXdFQi96QWRCZ05WSFE0RUZnUVVyRWJaQTNkOGJxdHc5SmdxeFU3cU50YnF1N1F3RFFZSgpLb1pJaHZjTkFRRUxCUUFEZ2dHQkFER3BxRzJFR2t5Z3V3bDhZSnZpS1pBN01uMjl5QjJjV2JGZFIxMU9LdE5ZCnRxSm82ZTA3NlJQMjFyQzNVTVNWeHNadXZ4a3hFQkRwL05SZzlpYjVleEcrU3l3dENZcUZKRWpCUlQwck5YWHMKUWE1NVlFak5WRTJYVFN2NzludUVGSVR6aC9PYVV0S2h2SVdoaWJPN2ZQL1lDUEo4SDFlN0NFYW5UbENId3ZqRwpnK241V0ZwVWJWUk1naElFa0pDM0g3MjlZdmplUWs4Z2pxZDZDdlBDb3p6YTBDRkp3ZVBLQXRGRXMxVmJyKzJiClE4MVdlaS9JWWUwTWZROExWSHl2cGVDbnVWR3Z0OERVRmRFdXVXV0pHOGVBYnZkWGRtWDZqbmJNUFJUUkJoUUEKeUQ4aDZkS2x5a2FVckcrM0Y3N00zek8yQk00ZXhqQ0NmTzNXMWZCaEVjUVR6dU5SRFhLY01LY3F5Z1QyNWhHcgpCM0lVa1U2T3crOG43c3hFWnd0cDF4THppWjJzUmdlNkI4MGloZGxCalc2aHYxMVkvRU9LNkI4M2RYSVo4d0lrCkFqV1BBOFlyZWhkSVpBOHhiOWJoQWNuZ2R2NnJqR0loY2h3N25tK3hQa25RM1pOcXZacjY1Y1RPQmVYU2lLYnoKN2hTT25iTGpNTjhZd05VREUvbUpjZz09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K`)
	testConfig := &clientcmdapi.Config{
		APIVersion: "v1",
		Kind:       "Config",
		Clusters: map[string]*clientcmdapi.Cluster{
			"cluster": {
				CertificateAuthorityData: cert,
				Server:                   `https://1.1.1.1`},
		},
		AuthInfos: map[string]*clientcmdapi.AuthInfo{"user": {Token: "test-token"}},
		Contexts: map[string]*clientcmdapi.Context{
			"context": {
				Cluster:  "cluster",
				AuthInfo: "user",
			},
		},
		CurrentContext: "context",
	}

	ctx := context.TODO()
	cli, err := NewKubernetesClient(ctx, testConfig)
	if err != nil {
		t.Fatalf("err is not nil; want nil; err = %s", err)
	}
	realCli, ok := cli.(*kubernetesClient)
	if !ok {
		t.Fatalf("cli is not *gke.kubernetesClient")
	}
	if realCli.ctx != ctx {
		t.Errorf("context is %v; want %v", realCli.ctx, ctx)
	}
	if realCli.client == nil {
		t.Errorf("client is nil; want dynamic.Interface")
	}
	if _, ok := realCli.discovery.(*discovery.DiscoveryClient); !ok {
		t.Errorf("discovery is not *discovery.DiscoveryClient")
	}
}

func TestGetNamespaces(t *testing.T) {
	ns1Name := "namespace-one"
	ns1 := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name": ns1Name,
		},
	}
	ns2Name := "namespace-two"
	ns2 := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name": ns2Name,
		},
	}

	dynCliMock := &kubeDynamicClientMock{
		ResourceFn: func(resource schema.GroupVersionResource) dynamic.NamespaceableResourceInterface {
			if resource.Resource != "namespaces" || resource.Version != "v1" {
				t.Fatal("resrouce function received resource that is not v1/namespaces")
			}
			return &kubeNamespacedResourceMock{
				ListFn: func(ctx context.Context, opts metav1.ListOptions) (*unstructured.UnstructuredList, error) {
					return &unstructured.UnstructuredList{
						Items: []unstructured.Unstructured{{Object: ns1}, {Object: ns2}},
					}, nil
				},
			}
		},
	}

	client := &kubernetesClient{ctx: context.TODO(), client: dynCliMock}
	results, err := client.GetNamespaces()
	if err != nil {
		t.Fatalf("err is not nil; want nil; err = %s", err)
	}
	if len(results) != 2 {
		t.Fatalf("number of results is %v; want %v", len(results), 2)
	}
	if results[0] != ns1Name {
		t.Errorf("result[0] is %v; want %v", results[0], ns1Name)
	}
	if results[1] != ns2Name {
		t.Errorf("result[1] is %v; want %v", results[1], ns2Name)
	}
}

func TestGetFetchableResourceTypes(t *testing.T) {
	list := []*metav1.APIResourceList{
		{
			GroupVersion: "apps/v1",
			APIResources: []metav1.APIResource{
				{Name: "deployments", Namespaced: true, Verbs: []string{"get", "delete"}},
				{Name: "statefulsets", Namespaced: true, Verbs: []string{"get", "delete"}},
			},
		},
		{
			GroupVersion: "mikotest/v2",
			APIResources: []metav1.APIResource{
				{Name: "sweets", Namespaced: false, Verbs: []string{"get", "eat"}},
				{Name: "cars", Namespaced: false, Verbs: []string{"ride", "repair"}},
			},
		},
	}
	expected := []*ResourceType{
		{Group: "apps", Version: "v1", Name: "deployments", Namespaced: true},
		{Group: "apps", Version: "v1", Name: "statefulsets", Namespaced: true},
		{Group: "mikotest", Version: "v2", Name: "sweets", Namespaced: false},
	}
	client := kubernetesClient{
		ctx: context.TODO(),
		discovery: dicoveryClientMock{
			ServerGroupsAndResourcesFn: func() ([]*metav1.APIGroup, []*metav1.APIResourceList, error) {
				return []*metav1.APIGroup{}, list, nil
			},
		},
	}

	results, err := client.GetFetchableResourceTypes()
	if err != nil {
		t.Fatalf("err is not nil; want nil; err = %s", err)
	}
	if len(results) != len(expected) {
		t.Fatalf("number of results is %d; want %d", len(results), len(expected))
	}
	for i := range results {
		result := *results[i]
		expected := *expected[i]
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("result[%d] is %v; want %v", i, result, expected)
		}
	}
}

func TestGetResources(t *testing.T) {
	resourceOne := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name": "some-object-one",
		},
	}
	resourceTwo := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name": "some-object-two",
		},
	}
	resourceType1 := ResourceType{Group: "apps", Version: "v1", Name: "deployments", Namespaced: true}
	namespace1 := "my-first-namespace"

	resourceType2 := ResourceType{Group: "apps", Version: "v1", Name: "pods", Namespaced: true}
	namespace2 := "my-second-namespace"

	dynCliMock := &kubeDynamicClientMock{
		ResourceFn: func(resource schema.GroupVersionResource) dynamic.NamespaceableResourceInterface {
			if resource.Group == resourceType1.Group && resource.Version == resourceType1.Version && resource.Resource == resourceType1.Name {
				return &kubeNamespacedResourceMock{
					NamespaceFn: func(n string) dynamic.ResourceInterface {
						if n == namespace1 {
							return &kubeNamespacedResourceMock{
								ListFn: func(ctx context.Context, opts metav1.ListOptions) (*unstructured.UnstructuredList, error) {
									return &unstructured.UnstructuredList{
										Items: []unstructured.Unstructured{{Object: resourceOne}},
									}, nil
								},
							}
						} else {
							return &kubeNamespacedResourceMock{
								ListFn: func(ctx context.Context, opts metav1.ListOptions) (*unstructured.UnstructuredList, error) {
									return &unstructured.UnstructuredList{
										Items: []unstructured.Unstructured{},
									}, nil
								},
							}
						}
					},
				}
			}
			if resource.Group == resourceType2.Group && resource.Version == resourceType2.Version && resource.Resource == resourceType2.Name {
				return &kubeNamespacedResourceMock{
					NamespaceFn: func(n string) dynamic.ResourceInterface {
						if n == namespace2 {
							return &kubeNamespacedResourceMock{
								ListFn: func(ctx context.Context, opts metav1.ListOptions) (*unstructured.UnstructuredList, error) {
									return &unstructured.UnstructuredList{
										Items: []unstructured.Unstructured{{Object: resourceTwo}},
									}, nil
								},
							}
						} else {
							return &kubeNamespacedResourceMock{
								ListFn: func(ctx context.Context, opts metav1.ListOptions) (*unstructured.UnstructuredList, error) {
									return &unstructured.UnstructuredList{
										Items: []unstructured.Unstructured{},
									}, nil
								},
							}
						}
					},
				}
			}
			return &kubeNamespacedResourceMock{
				ListFn: func(ctx context.Context, opts metav1.ListOptions) (*unstructured.UnstructuredList, error) {
					return &unstructured.UnstructuredList{
						Items: []unstructured.Unstructured{},
					}, nil
				},
			}
		},
	}
	expected := []*Resource{{Type: resourceType1, Data: resourceOne}, {Type: resourceType2, Data: resourceTwo}}

	client := &kubernetesClient{ctx: context.TODO(), client: dynCliMock}

	resourceTypes := []*ResourceType{&resourceType1, &resourceType2}
	namespaces := []string{namespace1, namespace2}
	results, err := client.GetResources(resourceTypes, namespaces)

	if err != nil {
		t.Fatalf("err is not nil; want nil; err = %s", err)
	}

	t.Logf("results:")
	for _, r := range results {
		t.Logf("%v", r)
	}
	t.Logf("expected:")
	for _, e := range expected {
		t.Logf("%v", e)
	}

	if len(results) != len(expected) {
		t.Fatalf("number of results is %d; want %d", len(results), len(expected))
	}
}

func TestGetNamespacedResources(t *testing.T) {
	resourceOne := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name": "some-object-one",
		},
	}
	resourceTwo := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name": "some-object-two",
		},
	}
	resourceType := ResourceType{Group: "apps", Version: "v1", Name: "deployments", Namespaced: true}
	namespace := "my-test-namespace"

	dynCliMock := &kubeDynamicClientMock{
		ResourceFn: func(resource schema.GroupVersionResource) dynamic.NamespaceableResourceInterface {
			if resource.Group != resourceType.Group {
				t.Fatalf("resource group is %v; want %v", resource.Group, resourceType.Group)
			}
			if resource.Version != resourceType.Version {
				t.Fatalf("resource version is %v; want %v", resource.Version, resourceType.Version)
			}
			if resource.Resource != resourceType.Name {
				t.Fatalf("resource name is %v; want %v", resource.Resource, resourceType.Name)
			}
			return &kubeNamespacedResourceMock{
				NamespaceFn: func(n string) dynamic.ResourceInterface {
					if n != namespace {
						t.Fatalf("namespace is %v; want %v", n, namespace)
					}
					return &kubeNamespacedResourceMock{
						ListFn: func(ctx context.Context, opts metav1.ListOptions) (*unstructured.UnstructuredList, error) {
							return &unstructured.UnstructuredList{
								Items: []unstructured.Unstructured{{Object: resourceOne}, {Object: resourceTwo}},
							}, nil
						},
					}
				},
			}
		},
	}
	expected := []*Resource{{Type: resourceType, Data: resourceOne}, {Type: resourceType, Data: resourceTwo}}

	client := &kubernetesClient{ctx: context.TODO(), client: dynCliMock}
	results, err := client.GetNamespacedResources(resourceType, namespace)
	if err != nil {
		t.Fatalf("err is not nil; want nil; err = %s", err)
	}
	if len(results) != len(expected) {
		t.Fatalf("number of results is %d; want %d", len(results), len(expected))
	}
	for i := range results {
		result := *results[i]
		expected := *expected[i]
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("result[%d] is %v; want %v", i, result, expected)
		}
	}
}

func TestGetNamespacedResources_notNamespaced(t *testing.T) {
	resourceType := ResourceType{Group: "apps", Version: "v1", Name: "deployments", Namespaced: false}
	client := &kubernetesClient{}
	_, err := client.GetNamespacedResources(resourceType, "my-test-namespace")
	if err == nil {
		t.Fatalf("err is not nil; want err")
	}
}

func TestStringSliceContains(t *testing.T) {
	hay := []string{"elem-one", "elem-two", "elem-three", "elem-two"}
	needle := "elem-two"
	if !stringSliceContains(hay, needle) {
		t.Errorf("result of stringSliceContains is false; want true")
	}
}

func TestStringSliceContains_negative(t *testing.T) {
	hay := []string{"elem-one", "elem-two", "elem-three", "elem-two"}
	needle := "elem-six"
	if stringSliceContains(hay, needle) {
		t.Errorf("result of stringSliceContains is true; want false")
	}
}
