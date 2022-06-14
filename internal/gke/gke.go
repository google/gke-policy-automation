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
	"encoding/base64"
	"fmt"
	"strings"

	container "cloud.google.com/go/container/apiv1"
	"github.com/google/gke-policy-automation/internal/log"
	"github.com/google/gke-policy-automation/internal/version"
	gax "github.com/googleapis/gax-go/v2"
	"google.golang.org/api/option"
	containerpb "google.golang.org/genproto/googleapis/container/v1"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type ClusterManagerClient interface {
	GetCluster(ctx context.Context, req *containerpb.GetClusterRequest, opts ...gax.CallOption) (*containerpb.Cluster, error)
	Close() error
}

type GKEClient struct {
	ctx      context.Context
	client   ClusterManagerClient
	k8client KubernetesClient
}

type Cluster struct {
	*containerpb.Cluster
	Resources []*Resource
}

func NewClient(ctx context.Context, k8sCheck bool) (*GKEClient, error) {
	return newGKEClient(ctx, k8sCheck)
}

func NewClientWithCredentialsFile(ctx context.Context, k8sCheck bool, credentialsFile string) (*GKEClient, error) {
	return newGKEClient(ctx, k8sCheck, option.WithCredentialsFile(credentialsFile))
}

func newGKEClient(ctx context.Context, k8sCheck bool, opts ...option.ClientOption) (*GKEClient, error) {
	opts = append(opts, option.WithUserAgent(version.UserAgent))
	cli, err := container.NewClusterManagerClient(ctx, opts...)
	if err != nil {
		return nil, err
	}

	var k8cli KubernetesClient = nil

	return &GKEClient{
		ctx:      ctx,
		client:   cli,
		k8client: k8cli,
	}, nil
}

func (c *GKEClient) GetCluster(name string, k8sCheck bool, apiVersions []string) (*Cluster, error) {
	req := &containerpb.GetClusterRequest{
		Name: name}
	cluster, err := c.client.GetCluster(c.ctx, req)
	if err != nil {
		return nil, err
	}

	var resources []*Resource = nil

	if k8sCheck {
		if c.k8client == nil {
			clusterToken, err := getClusterToken()
			if err != nil {
				log.Debugf("unable to get cluster token: %s", err)
				return nil, err
			}
			kubeConfig, err := getKubeConfig(cluster, clusterToken)
			if err != nil {
				log.Debugf("unable to get kubeconfig: %s", err)
				return nil, err
			}
			k8cli, err := NewKubernetesClient(c.ctx, kubeConfig)
			if err != nil {
				log.Debugf("unable to create kube client: %s", err)
				return nil, err
			}
			c.k8client = k8cli
		}
		resources, err = c.getResources(c.ctx, apiVersions)
		if err != nil {
			return nil, err
		}
	}

	return &Cluster{cluster, resources}, err
}

func (c *GKEClient) getResources(ctx context.Context, apiVersions []string) ([]*Resource, error) {

	var resources []*Resource
	namespaces, err := c.k8client.GetNamespaces()
	if err != nil {
		return nil, err
	}

	resourceTypes, err := c.k8client.GetFetchableResourceTypes()
	if err != nil {
		return nil, err
	}

	resourceGroupsToBeFetched := apiVersions

	toBeFetched := []*ResourceType{}
	for _, i := range resourceTypes {
		if stringSliceContains(resourceGroupsToBeFetched, buildApiVersionString(i.Version, i.Group)) && i.Namespaced {
			toBeFetched = append(toBeFetched, i)
		}
	}

	for ns := range namespaces {
		for rt := range toBeFetched {
			res, err := c.k8client.GetNamespacedResources(*toBeFetched[rt], namespaces[ns])
			resources = append(resources, res...)
			if err != nil {
				return nil, err
			}
		}
	}
	return resources, nil
}

func (c *GKEClient) Close() error {
	return c.client.Close()
}

func GetClusterName(project string, location string, name string) string {
	return fmt.Sprintf("projects/%s/locations/%s/clusters/%s", project, location, name)
}

func buildApiVersionString(version string, group string) string {

	if group != "" {
		return group + "/" + version
	}
	return version
}

func getKubeConfig(clusterData *containerpb.Cluster, clusterToken string) (*clientcmdapi.Config, error) {
	clusterMasterAuth := clusterData.MasterAuth.ClusterCaCertificate
	clusterEndpoint := clusterData.Endpoint
	clusterName := clusterData.Name
	clusterLocation := clusterData.GetLocation()
	clusterProject := strings.Split(clusterData.GetSelfLink(), "/")[5]
	clusterContext := fmt.Sprintf("gke_%v_%v_%v", clusterProject, clusterLocation, clusterName)
	config := clientcmdapi.NewConfig()
	cluster := clientcmdapi.NewCluster()

	caCert, err := base64.StdEncoding.DecodeString(clusterMasterAuth)
	if err != nil {
		log.Debugf("Unable to retrieve clusterMasterAuth %s:", err)
		return nil, err
	}
	log.Info("Cluster Master Auth retrieved")

	config.APIVersion = "v1"
	config.Kind = "Config"
	config.Clusters = map[string]*clientcmdapi.Cluster{
		clusterContext: {
			CertificateAuthorityData: caCert,
			Server:                   fmt.Sprintf("https://%v", clusterEndpoint),
		},
	}
	config.AuthInfos = map[string]*clientcmdapi.AuthInfo{
		clusterContext: {Token: clusterToken},
	}
	config.Contexts = map[string]*clientcmdapi.Context{
		clusterContext: {
			Cluster:  clusterContext,
			AuthInfo: clusterContext,
		},
	}

	config.CurrentContext = clusterContext
	cluster.CertificateAuthorityData = []byte(caCert)
	cluster.Server = fmt.Sprintf("https://%v", clusterEndpoint)
	log.Info("Local kubernetes cluster configuration created")
	return config, nil
}
