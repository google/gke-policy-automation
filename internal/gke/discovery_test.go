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
	"reflect"
	"testing"

	asset "cloud.google.com/go/asset/apiv1"
	gax "github.com/googleapis/gax-go/v2"
	"google.golang.org/api/iterator"
	assetpb "google.golang.org/genproto/googleapis/cloud/asset/v1"
)

type assetInventoryClientMock struct {
	searchAllResourcesFn func(ctx context.Context, req *assetpb.SearchAllResourcesRequest, opts ...gax.CallOption) *asset.ResourceSearchResultIterator
}

func (m assetInventoryClientMock) SearchAllResources(ctx context.Context, req *assetpb.SearchAllResourcesRequest, opts ...gax.CallOption) *asset.ResourceSearchResultIterator {
	return m.searchAllResourcesFn(ctx, req, opts...)
}

func (m assetInventoryClientMock) Close() error {
	return nil
}

type assetInventorySearchResultIteratorMock struct {
	nextFn func() (*assetpb.ResourceSearchResult, error)
}

func (m assetInventorySearchResultIteratorMock) Next() (*assetpb.ResourceSearchResult, error) {
	return m.nextFn()
}

func TestNewDiscoveryClient(t *testing.T) {
	client, err := NewDiscoveryClient(context.Background())
	if err != nil {
		t.Fatalf("err is not nil; want nil; err = %s", err)
	}
	assetInvDiscClient, ok := client.(*AssetInventoryDiscoveryClient)
	if !ok {
		t.Errorf("discovery client is not an AssetInventoryDiscoveryClient")
	}
	if _, ok := assetInvDiscClient.cli.(*asset.Client); !ok {
		t.Errorf("asset inventory client is not asset.Client")
	}
	if assetInvDiscClient.searchLimit != defaultSearchLimit {
		t.Errorf("searchLimit = %v; want %v", assetInvDiscClient.searchLimit, defaultSearchLimit)
	}
}

func TestGetClustersForScope(t *testing.T) {
	scope := "projects/myProject"
	searchFn := func(ctx context.Context, req *assetpb.SearchAllResourcesRequest, opts ...gax.CallOption) *asset.ResourceSearchResultIterator {
		if req.Scope != scope {
			t.Fatalf("scope in request = %v; want %v", req.Scope, scope)
		}
		if !reflect.DeepEqual(req.AssetTypes, []string{clusterAssetType}) {
			t.Fatalf("asset types in request = %v; want %v", req.AssetTypes, []string{clusterAssetType})
		}
		return &asset.ResourceSearchResultIterator{}
	}
	client := AssetInventoryDiscoveryClient{
		cli: assetInventoryClientMock{searchAllResourcesFn: searchFn},
	}
	if _, err := client.getClustersForScope(scope); err != nil {
		t.Fatalf("err is not nil; want nil; err = %s", err)
	}
}

func TestCollectResourceSearchResults(t *testing.T) {
	expected := []*assetpb.ResourceSearchResult{
		{Name: "testNameOne", AssetType: "testAssetType"},
		{Name: "testNameTwo", AssetType: "testAssetType"},
		nil,
	}
	errors := []error{
		nil,
		nil,
		iterator.Done,
	}
	i := 0
	nextFn := func() (res *assetpb.ResourceSearchResult, err error) {
		res, err = expected[i], errors[i]
		i++
		return
	}

	iteratorMock := assetInventorySearchResultIteratorMock{nextFn}
	client := AssetInventoryDiscoveryClient{searchLimit: defaultSearchLimit}
	results, err := client.collectResourceSearchResults(iteratorMock)
	if err != nil {
		t.Fatalf("err is not nil; want nil; err = %s", err)
	}
	if len(results) != len(expected)-1 {
		t.Fatalf("number of results = %v; want %v", len(results), len(expected)-1)
	}
	for i := range results {
		if !reflect.DeepEqual(results[i], expected[i]) {
			t.Errorf("result [%d] is %v; want %v", i, results[i], expected[i])
		}
	}
}

func TestCollectResourceSearchResults_limit(t *testing.T) {
	expected := []*assetpb.ResourceSearchResult{
		{Name: "testNameOne", AssetType: "testAssetType"},
		{Name: "testNameTwo", AssetType: "testAssetType"},
		nil,
	}
	errors := []error{
		nil,
		nil,
		iterator.Done,
	}
	searchLimit := 1
	i := 0
	nextFn := func() (res *assetpb.ResourceSearchResult, err error) {
		res, err = expected[i], errors[i]
		i++
		return
	}

	iteratorMock := assetInventorySearchResultIteratorMock{nextFn}
	client := AssetInventoryDiscoveryClient{searchLimit: searchLimit}
	results, err := client.collectResourceSearchResults(iteratorMock)
	if err != nil {
		t.Fatalf("err is not nil; want nil; err = %s", err)
	}
	if len(results) != searchLimit {
		t.Fatalf("number of results = %v; want %v", len(results), searchLimit)
	}
}

func TestCollectResourceSearchResults_negative(t *testing.T) {
	expected := []*assetpb.ResourceSearchResult{{}}
	errors := []error{fmt.Errorf("some strange error")}
	i := 0
	nextFn := func() (res *assetpb.ResourceSearchResult, err error) {
		res, err = expected[i], errors[i]
		i++
		return
	}
	iteratorMock := assetInventorySearchResultIteratorMock{nextFn}
	client := AssetInventoryDiscoveryClient{searchLimit: defaultSearchLimit}
	if _, err := client.collectResourceSearchResults(iteratorMock); err == nil {
		t.Errorf("error is nil; want error")
	}
}

func TestFilterMapSeachResults(t *testing.T) {
	id := "projects/my-project/locations/europe-west2/clusters/my-cluster"
	results := []*assetpb.ResourceSearchResult{
		{Name: fmt.Sprintf("//container.googleapis.com/%s", id), AssetType: clusterAssetType},
		{Name: "testName", AssetType: "testAssetType"},
		{Name: "invalidName", AssetType: clusterAssetType},
	}
	ids := filterMapSeachResults(results)
	if len(ids) != 1 {
		t.Fatalf("number of cluster identifiers = %v; want %v", len(ids), 1)
	}
	if ids[0] != id {
		t.Errorf("cluster identifier [0] = %v; want %v", ids[0], id)
	}
}

func TestGetIDFromName(t *testing.T) {
	id := "projects/my-project/locations/europe-west2/clusters/my-cluster"
	name := fmt.Sprintf("//container.googleapis.com/%s", id)
	result, err := getIDFromName(name)
	if err != nil {
		t.Fatalf("err is not nil; want nil; err = %s", err)
	}
	if result != id {
		t.Errorf("result is %s; want %s", result, id)
	}
}

func TestGetIDFromName_negative(t *testing.T) {
	inputs := []string{
		"projects/my-project/locations/europe-west2/clusters/my-cluster",
		"//container.googleapis.com/project/test/locations/europe/my-cluster",
	}
	for _, input := range inputs {
		if _, err := getIDFromName(input); err == nil {
			t.Errorf("input = %s; error is nil; want error", input)
		}
	}
}
