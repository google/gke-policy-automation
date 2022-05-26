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
	b64 "encoding/base64"
	"fmt"

	container "cloud.google.com/go/container/apiv1"
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

	if k8sCheck {
		k8cli, err = NewKubernetesClient(ctx, getKubeConfig())
		if err != nil {
			return nil, err
		}
	}

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

func getKubeConfig() *clientcmdapi.Config {
	cert, _ := b64.StdEncoding.DecodeString(`LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUVMVENDQXBXZ0F3SUJBZ0lSQU9jQ0t4dVY1bE43R3hjV0FjeWh2Rm93RFFZSktvWklodmNOQVFFTEJRQXcKTHpFdE1Dc0dBMVVFQXhNa04yRXdPVFkxWmpndE5URTROaTAwWVRVMkxUaGpaREF0T1dNNVpXTTNNR0prWW1FdwpNQ0FYRFRJeU1ETXhNekU1TlRFd01sb1lEekl3TlRJd016QTFNakExTVRBeVdqQXZNUzB3S3dZRFZRUURFeVEzCllUQTVOalZtT0MwMU1UZzJMVFJoTlRZdE9HTmtNQzA1WXpsbFl6Y3dZbVJpWVRBd2dnR2lNQTBHQ1NxR1NJYjMKRFFFQkFRVUFBNElCandBd2dnR0tBb0lCZ1FEUFQyNlpDZkZsQ3JWWFV4c2xBdm9DMSs1a3FoeVcrc3ZodGt3MApHSWJBTUNaNDcvRXNYREJneDRsMVNhZnNabkFWWlZiYkZhTGRkSXRDRlFOZXo1WGJDdjNCSU9uRTZxUVB3enZEClZhVHR2YlRpOHhjWlg4Y25iQUhFaHYzZkdnaE5BazNPWTRpcTNwVCtHY3ppa0JNSkFOclVpdENRdDBuQ0NjVm4KZTBIVEtub1RoNS9tVkJybVZkRGlyZ0w2dk0wYkhlcTVsTTRhYUwybGErOUEyeDRrR2c2Q0h2bWtsYVJVS3B1TgpwdC9aTlpaSDNDeVBITS9wdW9wUTJmc1JuN1A0Q245WGhGK084QStTbXB3QlV6RGpyVUVoUDRURXZYS21wcGpHClN1dklZMHhNcnE2Q0g5TTJaaFRVWGx2bENpTVlPbStWaXduMFQwb0lHZTNDZkxWL1NPRTFsMmlaNnBrSzNDOEgKMUtiMFRPdzFhM2I4OTlMWjI0bjM0N0haVXh0Wm9MaTFIbi96N1NzYy8zTlZxMEh2RTdLemkzZ3VoMGZ0c2FoVApZT0kybFJjbXFwWnJCK0hDamhSYWNkQ1RkZUNJWEFNZFJiZThTQkZvMW5MSkpoa0VtYWcyeVZzZGoxU0lCdDhECmprWFJrNUxtdkUxdS9PaDgzYVpiVTNnNHZVVUNBd0VBQWFOQ01FQXdEZ1lEVlIwUEFRSC9CQVFEQWdJRU1BOEcKQTFVZEV3RUIvd1FGTUFNQkFmOHdIUVlEVlIwT0JCWUVGSVZFMERvTjVVaDBBeVFDQldlQW81dmRaZlpkTUEwRwpDU3FHU0liM0RRRUJDd1VBQTRJQmdRREpVeG5ITmFYOVlsRnFNL3NCSXZkZ0M5VXZDMTR4c3hsWDJ0clBkOVhsCm9NeHhWZXg0Q0x4R001VER0dnd1Wmp3VnNTdkQ1Q0pWdmJYR3RzMTlmU3EzQXFuV3JRaWt2UmlaYVJhLzNNdFQKajE1RStKbXQwMGZVRDNXcHBHZmZhcGVjbUJHK1Yzd3dUNGpIeTVxTnoyMWxaN1pnV3A2ZTFjVldTbUUvaUtVRwpscDY3KzByMU0xNldUQ1hneGp5SmpmaUg3TlBEYm8yNVQxbjFVaHMzZ1BYd3Z5VGtsODZteDNZS3JTNXNTdkRDCm9MV1FaME1zVDNQalI4cHV4WXN6NkRWenBsMWxLaHlqOWxDSmFtVVpxQ0hNQWc0QzZLQWpydm9mWksrL1NlU2oKSDdQeW1VbWIxUm1ZZ0JkSCtyTE9CV2hwVU1DSE02VForSW10OG5rTDlxVGk1bDV5SG9zYmxFTldBTVNmODNJUQpvem50NlNvTW0rZmVOWVJiWVE0UDJJMXozOUNXSU1SVUl1aUJJS3ZuR1F1b094QTF5Y0ZHeStvVVV5Q2Z2OUh3CmlyWnVXVWVqVzV5U0pLZFZjUVI0OXlXV20vbE5sQzk0R3E2amM4c3hDeWwvWHVRUlF5RzlSZUExbmZ1TE9UdzEKNHBHZjd6UWhuRnpmYnNYWUtWVHFpUFk9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0=`)

	return &clientcmdapi.Config{
		APIVersion: "v1",
		Kind:       "Config",
		Clusters: map[string]*clientcmdapi.Cluster{
			"gke_project_europe-central2-a_cluster1": {
				CertificateAuthorityData: cert,
				Server:                   `https://1.1.1.1`},
		},
		AuthInfos: map[string]*clientcmdapi.AuthInfo{
			"gke_project_europe-central2-a_cluster1": {Token: "ya29.a0ARrdaM_kNe8CX9_DgCNw1rFPH_ZbtL3niYHw6xjgzjH6xzLIaKdwTxDb6YZaLIOAZRg1CwTmK2RJrSUb4MxcGrrn4FUFk3qaLar-oSlVt1uVVx87xNUqaPinrvpeg38sjj_WBvd1kl6buxoo8tTZKS-tkLAfLwudSkdpXXX"},
		},
		Contexts: map[string]*clientcmdapi.Context{
			"gke_project_europe-central2-a_cluster1": {
				Cluster:  "gke_project_europe-central2-a_cluster1",
				AuthInfo: "gke_project_europe-central2-a_cluster1",
			},
		},
		CurrentContext: "gke_project_europe-central2-a_cluster1",
	}
}
