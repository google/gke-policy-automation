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
	"regexp"
	"strings"

	container "cloud.google.com/go/container/apiv1"
	"github.com/google/gke-policy-automation/internal/log"
	"github.com/google/gke-policy-automation/internal/version"
	gax "github.com/googleapis/gax-go/v2"
	"google.golang.org/api/option"
	containerpb "google.golang.org/genproto/googleapis/container/v1"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type GKEClient interface {
	GetCluster(name string) (*Cluster, error)
	Close() error
}

type ClusterManagerClient interface {
	GetCluster(ctx context.Context, req *containerpb.GetClusterRequest, opts ...gax.CallOption) (*containerpb.Cluster, error)
	Close() error
}

type gkeApiClientBuilder struct {
	ctx             context.Context
	credentialsFile string
	k8sApiVersions  []string
	metrics         []MetricQuery
	k8sMaxQPS       int
}

func NewGKEApiClientBuilder(ctx context.Context) *gkeApiClientBuilder {
	return &gkeApiClientBuilder{ctx: ctx}
}

func (b *gkeApiClientBuilder) WithCredentialsFile(credentialsFile string) *gkeApiClientBuilder {
	b.credentialsFile = credentialsFile
	return b
}

func (b *gkeApiClientBuilder) WithK8SClient(apiVersions []string, maxQPS int) *gkeApiClientBuilder {
	b.k8sApiVersions = apiVersions
	b.k8sMaxQPS = maxQPS
	return b
}

func (b *gkeApiClientBuilder) WithMetricsClient(metricQueries []MetricQuery) *gkeApiClientBuilder {
	b.metrics = metricQueries
	return b
}

func (b *gkeApiClientBuilder) Build() (GKEClient, error) {
	opts := []option.ClientOption{option.WithUserAgent(version.UserAgent)}
	if b.credentialsFile != "" {
		opts = append(opts, option.WithCredentialsFile(b.credentialsFile))
	}

	cli, err := container.NewClusterManagerClient(b.ctx, opts...)
	if err != nil {
		return nil, err
	}

	var metricQueries []MetricQuery
	if len(b.metrics) > 0 {
		metricQueries = b.metrics
	}

	return &GKEApiClient{
		ctx:              b.ctx,
		client:           cli,
		authTokenFunc:    getClusterToken,
		k8sClientFunc:    NewKubernetesClient,
		k8sApiVersions:   b.k8sApiVersions,
		k8sMaxQPS:        b.k8sMaxQPS,
		metricClientFunc: NewMetricClient,
		metricQueries:    metricQueries,
	}, nil
}

type authTokenFunc func(ctx context.Context) (string, error)
type k8sClientFunc func(ctx context.Context, kubeConfig *clientcmdapi.Config, maxQPS int) (KubernetesClient, error)
type metricClientFunc func(ctx context.Context, projectId string, authToken string) (MetricsClient, error)

type GKEApiClient struct {
	ctx              context.Context
	client           ClusterManagerClient
	k8sClientFunc    k8sClientFunc
	authTokenFunc    authTokenFunc
	k8sApiVersions   []string
	k8sMaxQPS        int
	metricClientFunc metricClientFunc
	metricQueries    []MetricQuery
}

type Cluster struct {
	*containerpb.Cluster
	Resources []*Resource
	Metrics   map[string]Metric
}

func (c Cluster) ReadableId() string {
	r := regexp.MustCompile(`.+/(projects/.+/(locations|zones)/.+/clusters/.+)`)
	if !r.MatchString(c.SelfLink) {
		log.Warnf("cluster selfLink %s does not match readable identifier regex", c.SelfLink)
		return c.Id
	}
	matches := r.FindStringSubmatch(c.SelfLink)
	if len(matches) != 3 {
		log.Warnf("cluster selfLink %s has invalid number of readable identifier regex matches", c.SelfLink)
		return c.Id
	}
	return matches[1]
}

// GetCluster returns a Cluster object with all the information regarding the cluster,
// externally through the Containers API and Internally with the K8s APIs
func (c *GKEApiClient) GetCluster(name string) (*Cluster, error) {
	req := &containerpb.GetClusterRequest{
		Name: name}
	cluster, err := c.client.GetCluster(c.ctx, req)
	if err != nil {
		return nil, err
	}

	var resources []*Resource = nil
	metricMap := make(map[string]Metric)

	if len(c.k8sApiVersions) > 0 {
		clusterToken, err := c.authTokenFunc(c.ctx)
		if err != nil {
			log.Debugf("unable to get cluster token: %s", err)
			return nil, err
		}
		kubeConfig, err := getKubeConfig(cluster, clusterToken)
		if err != nil {
			log.Debugf("unable to get kubeconfig: %s", err)
			return nil, err
		}
		k8cli, err := c.k8sClientFunc(c.ctx, kubeConfig, c.k8sMaxQPS)
		if err != nil {
			return nil, err
		}
		resources, err = getResources(k8cli, c.k8sApiVersions)
		if err != nil {
			return nil, err
		}
	}

	if len(c.metricQueries) > 0 {
		clusterToken, err := c.authTokenFunc(c.ctx)
		if err != nil {
			log.Debugf("unable to get cluster token: %s", err)
			return nil, err
		}
		metricsClient, err := c.metricClientFunc(c.ctx, getProjectIdFromSelfLink(cluster.SelfLink), clusterToken)
		if err != nil {
			log.Debugf("unable to create metrics client: %s", err)
			return nil, err
		}

		res, err := metricsClient.GetMetricsForCluster(c.metricQueries, cluster.Name)
		if err != nil {
			log.Debugf("unable to get metric: %s", err)
			return nil, err
		}

		metricMap = res
	}
	return &Cluster{cluster, resources, metricMap}, err
}

// Close closes the client connection
func (c *GKEApiClient) Close() error {
	return c.client.Close()
}

// getResources returns an array of k8s resources that the tool has been able to fetch after the auth
func getResources(client KubernetesClient, apiVersions []string) ([]*Resource, error) {
	namespaces, err := client.GetNamespaces()
	if err != nil {
		return nil, err
	}

	resourceTypes, err := client.GetFetchableResourceTypes()
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

	return client.GetResources(toBeFetched, namespaces)
}

// GetClusterName returns the cluster's self-link in gcp
func GetClusterName(project string, location string, name string) string {
	return fmt.Sprintf("projects/%s/locations/%s/clusters/%s", project, location, name)
}

func buildApiVersionString(version string, group string) string {
	if group != "" {
		return group + "/" + version
	}
	return version
}

// getKubeConfig create a kubeconfig configuration file from a given clusterData and a gcp auth token
func getKubeConfig(clusterData *containerpb.Cluster, clusterToken string) (*clientcmdapi.Config, error) {
	clusterMasterAuth := clusterData.MasterAuth.ClusterCaCertificate
	clusterEndpoint := clusterData.Endpoint
	clusterName := clusterData.Name
	clusterLocation := clusterData.GetLocation()
	clusterProject := strings.Split(clusterData.GetSelfLink(), "/")[5]
	clusterContext := fmt.Sprintf("gke_%v_%v_%v", clusterProject, clusterLocation, clusterName)
	config := clientcmdapi.NewConfig()

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
	log.Info("Local kubernetes cluster configuration created")
	return config, nil
}

func getProjectIdFromSelfLink(selfLink string) string {

	r := regexp.MustCompile(`/projects/.+/(locations|zones)/.+`)
	if !r.MatchString(selfLink) {
		log.Errorf("cluster selfLink %s does not match selflink format", selfLink)
		return ""
	}
	matches := r.FindStringSubmatch(selfLink)

	if len(matches) < 2 {
		log.Errorf("cluster selfLink %s does not match selflink format", selfLink)
		return ""
	}
	match := matches[0]
	cuttingBySlash := strings.FieldsFunc(match, func(r rune) bool {
		if r == '/' {
			return true
		}
		return false
	})

	if len(cuttingBySlash) < 2 {
		log.Error("Error getting project id from selflink: " + selfLink)
	}
	return cuttingBySlash[1]
}
