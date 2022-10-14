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
	"regexp"

	asset "cloud.google.com/go/asset/apiv1"
	"github.com/google/gke-policy-automation/internal/log"
	"github.com/google/gke-policy-automation/internal/version"
	gax "github.com/googleapis/gax-go/v2"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	assetpb "google.golang.org/genproto/googleapis/cloud/asset/v1"
)

const (
	clusterAssetType   = "container.googleapis.com/Cluster"
	defaultSearchLimit = 10000
)

type AssetInventoryClient interface {
	SearchAllResources(ctx context.Context, req *assetpb.SearchAllResourcesRequest, opts ...gax.CallOption) *asset.ResourceSearchResultIterator
	Close() error
}

type AssetInventorySearchResultIterator interface {
	Next() (*assetpb.ResourceSearchResult, error)
}

type DiscoveryClient interface {
	GetClustersInProject(name string) ([]string, error)
	GetClustersInFolder(number string) ([]string, error)
	GetClustersInOrg(number string) ([]string, error)
	Close() error
}

type AssetInventoryDiscoveryClient struct {
	cli         AssetInventoryClient
	ctx         context.Context
	searchLimit int
}

func NewDiscoveryClient(ctx context.Context) (DiscoveryClient, error) {
	return newAssetInventoryDiscoveryClient(ctx)
}

func NewDiscoveryClientWithCredentialsFile(ctx context.Context, credentialsFile string) (DiscoveryClient, error) {
	return newAssetInventoryDiscoveryClient(ctx, option.WithCredentialsFile(credentialsFile))
}

func newAssetInventoryDiscoveryClient(ctx context.Context, opts ...option.ClientOption) (*AssetInventoryDiscoveryClient, error) {
	opts = append(opts, option.WithUserAgent(version.UserAgent))
	client, err := asset.NewClient(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return &AssetInventoryDiscoveryClient{ctx: ctx, cli: client, searchLimit: defaultSearchLimit}, nil
}

// GetClustersInProject finds GKE clusters in a given GCP project (identified by name)
// and returns slice with their identifiers.
func (c *AssetInventoryDiscoveryClient) GetClustersInProject(name string) ([]string, error) {
	scope := fmt.Sprintf("projects/%s", name)
	return c.getClustersForScope(scope)
}

// GetClustersInFolder finds GKE clusters in a given GCP folder (identified by number)
// and returns slice with their identifiers.
func (c *AssetInventoryDiscoveryClient) GetClustersInFolder(number string) ([]string, error) {
	scope := fmt.Sprintf("folders/%s", number)
	return c.getClustersForScope(scope)
}

// GetClustersInFolder finds GKE clusters in a given GCP organization (identified by number)
// and returns slice with their identifiers.
func (c *AssetInventoryDiscoveryClient) GetClustersInOrg(number string) ([]string, error) {
	scope := fmt.Sprintf("organizations/%s", number)
	return c.getClustersForScope(scope)
}

// Close closes the client and underlying connections to other services.
func (c *AssetInventoryDiscoveryClient) Close() error {
	return c.cli.Close()
}

// getClustersForScope searches for a GKE clusters in a given Asset Inventory scope
// and returns slice with cluster identifiers.
func (c *AssetInventoryDiscoveryClient) getClustersForScope(scope string) ([]string, error) {
	req := &assetpb.SearchAllResourcesRequest{
		Scope:      scope,
		AssetTypes: []string{clusterAssetType}}

	results, err := c.clusterSearch(req)
	if err != nil {
		return nil, err
	}
	return filterMapSeachResults(results), nil
}

// clusterSearch runs asset inventory searchAllResults with a given request and iterates
// through the results, returning them as a slice.
func (c *AssetInventoryDiscoveryClient) clusterSearch(req *assetpb.SearchAllResourcesRequest) ([]*assetpb.ResourceSearchResult, error) {
	log.Debugf("cluster search with request: %s", req)
	return c.collectResourceSearchResults(c.cli.SearchAllResources(c.ctx, req))
}

// collectResourceSearchResults collects ResourceSearchResult with a given iterator.
func (c *AssetInventoryDiscoveryClient) collectResourceSearchResults(it AssetInventorySearchResultIterator) ([]*assetpb.ResourceSearchResult, error) {
	results := make([]*assetpb.ResourceSearchResult, 0)
	i := 0
	for ; i < c.searchLimit; i++ {
		result, err := it.Next()
		if err == iterator.Done {
			log.Debugf("search iterator done")
			break
		}
		if err != nil {
			return nil, err
		}
		log.Debugf("search iterator result: %s", result)
		results = append(results, result)
	}
	if i == c.searchLimit {
		log.Warnf("search limit of %d was reached", c.searchLimit)
	}
	return results, nil
}

// filterMapSeachResults filters search results to GKE clusters, maps to
// the cluster identifiers and returns as a slice.
func filterMapSeachResults(results []*assetpb.ResourceSearchResult) []string {
	identifiers := make([]string, 0, len(results))
	for _, result := range results {
		if result.AssetType != clusterAssetType {
			log.Debugf("skipping search result as it is not a cluster asset type of %q", clusterAssetType)
			continue
		}
		id, err := getIDFromName(result.Name)
		if err != nil {
			log.Warnf("skipping cluster asset search result due to invalid name: %s", err)
			continue
		}
		identifiers = append(identifiers, id)
	}
	return identifiers
}

// getIDFromName returns cluster identifier from full cluster asset name.
func getIDFromName(name string) (string, error) {
	r := regexp.MustCompile(`//container\.googleapis\.com/(projects/.+/(locations|zones)/.+/clusters/.+)`)
	if !r.MatchString(name) {
		return "", fmt.Errorf("given name %q does not match GKE cluster name pattern", name)
	}
	matches := r.FindStringSubmatch(name)
	if len(matches) != 3 {
		return "", fmt.Errorf("invalid number of regexp matches")
	}
	return matches[1], nil
}
