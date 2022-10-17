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

package inputs

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/google/gke-policy-automation/internal/inputs/clients"
	containerpb "google.golang.org/genproto/googleapis/container/v1"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type tsMock struct {
	getAuthTokenFn func() (string, error)
}

func (m *tsMock) GetAuthToken() (string, error) {
	return m.getAuthTokenFn()
}

type k8sClientMock struct {
}

func (k8sClientMock) GetNamespaces() ([]string, error) {
	return []string{"namespace-one", "namespace-two"}, nil
}

func (k8sClientMock) GetFetchableResourceTypes() ([]*clients.ResourceType, error) {
	return []*clients.ResourceType{
		{
			Group:      "autoscaling",
			Version:    "v1",
			Name:       "horizontalpodautoscalers",
			Namespaced: true,
		},
		{
			Group:      "",
			Version:    "v1",
			Name:       "replicationcontrollers",
			Namespaced: true,
		},
		{
			Group:      "",
			Version:    "v1",
			Name:       "componentstatuses",
			Namespaced: false,
		},
		{
			Group:      "authorization.k8s.io",
			Version:    "v1",
			Name:       "localsubjectaccessreviews",
			Namespaced: true,
		},
	}, nil
}

func (k8sClientMock) GetNamespacedResources(resourceType clients.ResourceType, namespace string) ([]*clients.Resource, error) {
	return []*clients.Resource{
		{
			Type: resourceType,
			Data: nil,
		},
	}, nil
}

func (k8sClientMock) GetResources(resourceType []*clients.ResourceType, namespace []string) ([]*clients.Resource, error) {
	return []*clients.Resource{
		{
			Type: *resourceType[0],
			Data: nil,
		},
	}, nil
}

func TestK8sApiInputBuilder(t *testing.T) {
	credFile := "test-fixtures/test_credentials.json"
	apiVersions := []string{"policy/v1", "networking.k8s.io/v1"}
	maxQPS := 69
	b := NewK8sApiInputBuilder(context.Background(), apiVersions).
		WithCredentialsFile(credFile).
		WithMaxQPS(maxQPS)

	input, err := b.Build()
	if err != nil {
		t.Fatalf("err = %v; want nil", err)
	}
	k8sInput, ok := input.(*k8sApiInput)
	if b.credentialsFile != credFile {
		t.Errorf("builder credentialsFile = %v; want %v", b.credentialsFile, credFile)
	}
	if !ok {
		t.Fatalf("input is not *ks8ApiInput")
	}
	if k8sInput.maxQPS != maxQPS {
		t.Errorf("maxQPS = %v; want %v", k8sInput.maxQPS, maxQPS)
	}
	if !reflect.DeepEqual(k8sInput.apiVersions, apiVersions) {
		t.Errorf("apiversions = %v; want %v", k8sInput.apiVersions, apiVersions)
	}
	if k8sInput.tokenSource == nil {
		t.Errorf("tokenSource is nil")
	}
	if k8sInput.gkeInput == nil {
		t.Errorf("gkeInput is nil")
	}
}

func TestK8SApiInputGetID(t *testing.T) {
	input := k8sApiInput{}
	if id := input.GetID(); id != k8sApiInputID {
		t.Fatalf("id = %v; want %v", id, k8sApiInputID)
	}
}

func TestK8SApiInputGetDescription(t *testing.T) {
	input := k8sApiInput{}
	if id := input.GetDescription(); id != k8sApiInputDescription {
		t.Fatalf("id = %v; want %v", id, k8sApiInputDescription)
	}
}

func TestK8SApiInputClose(t *testing.T) {
	input := k8sApiInput{
		gkeInput: &inputMock{
			closeFn: func() error { return errors.New("test error") },
		},
	}
	err := input.Close()
	if err == nil {
		t.Errorf("k8sApiInput close() error is nil; want mocked error")
	}
}

func TestK8sApiInputGetData(t *testing.T) {
	testClusterID := "projects/myproject/locations/europe-central2/clusters/cluster-one"
	testMaxQPS := 100
	i := k8sApiInput{
		ctx: context.Background(),
		tokenSource: &tsMock{
			getAuthTokenFn: func() (string, error) {
				return "token", nil
			},
		},
		gkeInput: &inputMock{
			getDataFn: func(clusterID string) (interface{}, error) {
				if clusterID != testClusterID {
					t.Errorf("clusterID = %v; want %v", clusterID, testClusterID)
				}
				return &containerpb.Cluster{
					MasterAuth: &containerpb.MasterAuth{
						ClusterCaCertificate: base64.StdEncoding.EncodeToString([]byte("test")),
					},
					Endpoint: "some.endpoint.test",
				}, nil
			},
		},
		newK8SClientFunc: func(ctx context.Context, kubeConfig *clientcmdapi.Config) (clients.KubernetesClient, error) {
			return &k8sClientMock{}, nil
		},
		apiVersions: []string{"autoscaling/v1"},
		maxQPS:      testMaxQPS,
	}
	_, err := i.GetData(testClusterID)
	if err != nil {
		t.Fatalf("err = %v; want nil", err)
	}
}

func TestCreateKubeConfig(t *testing.T) {
	token := "test-token"
	clusterEndpoint := "some.endpoint.test"
	clusterCert := []byte("cert-data-test")
	clusterCertEncoded := base64.StdEncoding.EncodeToString(clusterCert)
	data := &containerpb.Cluster{
		MasterAuth: &containerpb.MasterAuth{
			ClusterCaCertificate: clusterCertEncoded,
		},
		Endpoint: clusterEndpoint,
	}

	config, err := createKubeConfig(data, token)
	if err != nil {
		t.Fatalf("err = %v; want nil", err)
	}
	clusterConfig, ok := config.Clusters[k8sKubeConfigContextName]
	if !ok {
		t.Fatalf("config has no definition of a cluster %v", k8sKubeConfigContextName)
	}
	if clusterConfig.Server != fmt.Sprintf("https://%s", clusterEndpoint) {
		t.Errorf("clusterConfig server = %v; want %v", clusterConfig.Server, fmt.Sprintf("https://%s", clusterEndpoint))
	}
	if !reflect.DeepEqual(clusterConfig.CertificateAuthorityData, clusterCert) {
		t.Errorf("clusterConfig certificateAuthorityData = %v; want %v", clusterConfig.CertificateAuthorityData, clusterCert)
	}
	authConfig, ok := config.AuthInfos[k8sKubeConfigContextName]
	if !ok {
		t.Fatalf("config has no definition of a authInfo %v", k8sKubeConfigContextName)
	}
	if authConfig.Token != token {
		t.Errorf("authInfo token = %v; want %v", authConfig.Token, token)
	}
	contextConfig, ok := config.Contexts[k8sKubeConfigContextName]
	if !ok {
		t.Fatalf("config has no definition of a context %v", k8sKubeConfigContextName)
	}
	if contextConfig.Cluster != k8sKubeConfigContextName {
		t.Errorf("contextConfig cluster = %v; want %v", contextConfig.Cluster, k8sKubeConfigContextName)
	}
	if contextConfig.AuthInfo != k8sKubeConfigContextName {
		t.Errorf("contextConfig authInfo = %v; want %v", contextConfig.AuthInfo, k8sKubeConfigContextName)
	}
	if config.CurrentContext != k8sKubeConfigContextName {
		t.Errorf("currentContext = %v; want %v", config.CurrentContext, k8sKubeConfigContextName)
	}
}
