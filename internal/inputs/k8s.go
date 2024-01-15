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
	"fmt"

	"cloud.google.com/go/container/apiv1/containerpb"
	"github.com/google/gke-policy-automation/internal/inputs/clients"
	"github.com/google/gke-policy-automation/internal/log"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

const (
	k8sAPIInputID            = "k8sAPI"
	k8sAPIInputDescription   = "Cluster resource data from Kubernetes API"
	k8sDataSourceName        = "k8s"
	k8sKubeConfigContextName = "gke"
)

type newK8SClientFunc func(ctx context.Context, kubeConfig *clientcmdapi.Config) (clients.KubernetesClient, error)

type k8sAPIInput struct {
	ctx              context.Context
	tokenSource      clients.TokenSource
	gkeInput         Input
	newK8SClientFunc newK8SClientFunc
	k8sClient        clients.KubernetesClient
	apiVersions      []string
	maxQPS           int
	maxGoRoutines    int
	timeoutSeconds   int
}

type k8sInputBuilder struct {
	ctx             context.Context
	credentialsFile string
	apiVersions     []string
	maxQPS          int
	maxGoRoutines   int
	timeoutSeconds  int
}

func NewK8sAPIInputBuilder(ctx context.Context, apiVersions []string) *k8sInputBuilder {
	return &k8sInputBuilder{
		ctx:         ctx,
		apiVersions: apiVersions,
	}
}

func (b *k8sInputBuilder) WithCredentialsFile(credentialsFile string) *k8sInputBuilder {
	b.credentialsFile = credentialsFile
	return b
}

func (b *k8sInputBuilder) WithMaxQPS(maxQPS int) *k8sInputBuilder {
	b.maxQPS = maxQPS
	return b
}

func (b *k8sInputBuilder) WithMaxGoroutines(maxGoRoutines int) *k8sInputBuilder {
	b.maxGoRoutines = maxGoRoutines
	return b
}

func (b *k8sInputBuilder) WithClientTimeoutSeconds(timeoutSeconds int) *k8sInputBuilder {
	b.timeoutSeconds = timeoutSeconds
	return b
}

func (b *k8sInputBuilder) Build() (Input, error) {
	var ts clients.TokenSource
	var gkeInput Input
	var err error

	if b.credentialsFile != "" {
		ts, err = clients.NewGoogleTokenSourceWithCredentials(b.ctx, b.credentialsFile)
		if err != nil {
			return nil, err
		}
		gkeInput, err = NewGKEApiInputWithCredentials(b.ctx, b.credentialsFile)
		if err != nil {
			return nil, err
		}
	} else {
		ts, err = clients.NewGoogleTokenSource(b.ctx)
		if err != nil {
			return nil, err
		}
		gkeInput, err = NewGKEApiInput(b.ctx)
		if err != nil {
			return nil, err
		}
	}

	input := &k8sAPIInput{
		ctx:            b.ctx,
		tokenSource:    ts,
		gkeInput:       gkeInput,
		apiVersions:    b.apiVersions,
		maxQPS:         b.maxQPS,
		maxGoRoutines:  b.maxGoRoutines,
		timeoutSeconds: b.timeoutSeconds,
	}
	input.newK8SClientFunc = input.newK8sClientFromBuilder
	return input, nil
}

func (i *k8sAPIInput) GetID() string {
	return k8sAPIInputID
}

func (i *k8sAPIInput) GetDescription() string {
	return k8sAPIInputDescription
}

func (i *k8sAPIInput) GetDataSourceName() string {
	return k8sDataSourceName
}

func (i *k8sAPIInput) GetData(clusterID string) (interface{}, error) {
	if i.k8sClient == nil {
		if err := i.createK8SClient(clusterID); err != nil {
			return nil, err
		}
	}
	namespaces, err := i.k8sClient.GetNamespaces()
	if err != nil {
		return nil, err
	}

	resourceTypes, err := i.k8sClient.GetFetchableResourceTypes()
	if err != nil {
		return nil, err
	}

	toBeFetched := []*clients.ResourceType{}
	for _, t := range resourceTypes {
		if clients.StringSliceContains(i.apiVersions, buildAPIVersionString(t.Version, t.Group)) && t.Namespaced {
			toBeFetched = append(toBeFetched, t)
		}
	}
	return i.k8sClient.GetResources(toBeFetched, namespaces)
}

func (i *k8sAPIInput) Close() error {
	if i.gkeInput != nil {
		return i.gkeInput.Close()
	}
	return nil
}

func (i *k8sAPIInput) createK8SClient(clusterID string) error {
	token, err := i.tokenSource.GetAuthToken()
	if err != nil {
		return err
	}
	data, err := i.gkeInput.GetData(clusterID)
	if err != nil {
		return err
	}
	cluster := data.(*containerpb.Cluster)
	kubeConfig, err := createKubeConfig(cluster, token)
	if err != nil {
		return err
	}
	i.k8sClient, err = i.newK8SClientFunc(i.ctx, kubeConfig)
	if err != nil {
		return err
	}
	return nil
}

func (i *k8sAPIInput) newK8sClientFromBuilder(ctx context.Context, kubeConfig *clientcmdapi.Config) (clients.KubernetesClient, error) {
	client, err := clients.NewKubernetesClientBuilder(ctx, kubeConfig).
		WithMaxQPS(i.maxQPS).
		WithMaxGoroutines(i.maxGoRoutines).
		WithTimeout(i.timeoutSeconds).
		Build()
	return client, err
}

func createKubeConfig(clusterData *containerpb.Cluster, clusterToken string) (*clientcmdapi.Config, error) {
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

func buildAPIVersionString(version string, group string) string {
	if group != "" {
		return group + "/" + version
	}
	return version
}
