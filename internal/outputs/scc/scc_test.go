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

package scc

import (
	"context"
	"regexp"
	"testing"
	"time"

	scc "cloud.google.com/go/securitycenter/apiv1"
	gax "github.com/googleapis/gax-go/v2"
	"github.com/stretchr/testify/assert"
	"google.golang.org/api/iterator"
	sccpb "google.golang.org/genproto/googleapis/cloud/securitycenter/v1"
)

type sccApiClientMock struct {
	ListSourcesFn   func(ctx context.Context, req *sccpb.ListSourcesRequest, opts ...gax.CallOption) *scc.SourceIterator
	CreateSourceFn  func(ctx context.Context, req *sccpb.CreateSourceRequest, opts ...gax.CallOption) (*sccpb.Source, error)
	ListFindingsFn  func(ctx context.Context, req *sccpb.ListFindingsRequest, opts ...gax.CallOption) *scc.ListFindingsResponse_ListFindingsResultIterator
	UpdateFindingFn func(ctx context.Context, req *sccpb.UpdateFindingRequest, opts ...gax.CallOption) (*sccpb.Finding, error)
}

func (m *sccApiClientMock) ListSources(ctx context.Context, req *sccpb.ListSourcesRequest, opts ...gax.CallOption) *scc.SourceIterator {
	return m.ListSourcesFn(ctx, req, opts...)
}

func (m *sccApiClientMock) CreateSource(ctx context.Context, req *sccpb.CreateSourceRequest, opts ...gax.CallOption) (*sccpb.Source, error) {
	return m.CreateSourceFn(ctx, req, opts...)
}

func (m *sccApiClientMock) ListFindings(ctx context.Context, req *sccpb.ListFindingsRequest, opts ...gax.CallOption) *scc.ListFindingsResponse_ListFindingsResultIterator {
	return m.ListFindingsFn(ctx, req, opts...)
}

func (m *sccApiClientMock) UpdateFinding(ctx context.Context, req *sccpb.UpdateFindingRequest, opts ...gax.CallOption) (*sccpb.Finding, error) {
	return m.UpdateFindingFn(ctx, req, opts...)
}

type sccSourceIteratorMock struct {
	NextFn func() (*sccpb.Source, error)
}

func (m *sccSourceIteratorMock) Next() (*sccpb.Source, error) {
	return m.NextFn()
}

func TestGetActiveFindings(t *testing.T) {
	source := "source"
	resource := "resource"
	category := "category"
	mock := &sccApiClientMock{
		ListFindingsFn: func(ctx context.Context, req *sccpb.ListFindingsRequest, opts ...gax.CallOption) *scc.ListFindingsResponse_ListFindingsResultIterator {
			if req.Parent != source {
				t.Errorf("parent = %v; want %v", req.Parent, source)
			}
			r := regexp.MustCompile("resourceName=\"(.+)\" AND category=\"(.+)\" AND state=\"(.+)\"")
			if !r.MatchString(req.Filter) {
				t.Fatalf("filter does not match regexp")
			}
			matches := r.FindStringSubmatch(req.Filter)
			if matches[1] != resource {
				t.Errorf("resourceName in filter = %v; want %v", matches[1], resource)
			}
			if matches[2] != category {
				t.Errorf("category in filter = %v; want %v", matches[2], category)
			}
			if matches[3] != FINDING_STATE_STRING_ACTIVE {
				t.Errorf("state in filter = %v; want %v", matches[3], FINDING_STATE_STRING_ACTIVE)
			}
			return &scc.ListFindingsResponse_ListFindingsResultIterator{}
		},
	}
	cli := securityCommandCenterClientImpl{ctx: context.TODO(), client: mock, sourcesSearchLimit: 0}
	cli.getActiveFindings(source, resource, category)
}

func TestCreateFinding(t *testing.T) {
	sourceName := "sourceName"
	finding := &Finding{
		Time:         time.Now(),
		ResourceName: "cluster-resource",
		Category:     "category",
		Description:  "description",
		State:        FINDING_STATE_STRING_ACTIVE,
		Severity:     FINDING_SEVERITY_STRING_HIGH,
	}
	mock := &sccApiClientMock{
		UpdateFindingFn: func(ctx context.Context, req *sccpb.UpdateFindingRequest, opts ...gax.CallOption) (*sccpb.Finding, error) {
			findingTime := req.Finding.EventTime.AsTime()
			if findingTime != finding.Time.UTC() {
				t.Errorf("finding time = %v; want %v", findingTime, finding.Time.UTC())
			}
			if req.Finding.ResourceName != finding.ResourceName {
				t.Errorf("finding resource name = %v; want %v", req.Finding.ResourceName, finding.ResourceName)
			}
			if req.Finding.Category != finding.Category {
				t.Errorf("finding category = %v; want %v", req.Finding.Category, finding.Category)
			}
			if req.Finding.Description != finding.Description {
				t.Errorf("finding description = %v; want %v", req.Finding.Description, finding.Description)
			}
			if req.Finding.State.String() != finding.State {
				t.Errorf("finding state = %v; want %v", req.Finding.State.String(), finding.State)
			}
			if req.Finding.Severity.String() != finding.Severity {
				t.Errorf("finding severity = %v; want %v", req.Finding.Severity.String(), finding.Severity)
			}
			if req.UpdateMask != nil {
				t.Errorf("update mask = %v; want nil", req.UpdateMask)
			}
			return req.Finding, nil
		},
	}
	cli := securityCommandCenterClientImpl{ctx: context.TODO(), client: mock}
	result, err := cli.createFinding(sourceName, finding)
	if err != nil {
		t.Fatalf("err = %v; want nil", err)
	}
	r := regexp.MustCompile(".+/findings/(.+)$")
	if !r.MatchString(result) {
		t.Fatalf("finding name does not match regexp")
	}
	matches := r.FindStringSubmatch(result)
	if len(matches[1]) != 32 {
		t.Fatalf("length of generated finding ID = %v; want %v", len(matches[0]), 32)
	}
}

func TestUpdateFindingEventTimeFinding(t *testing.T) {
	findingName := "test"
	time := time.Now()
	mock := &sccApiClientMock{
		UpdateFindingFn: func(ctx context.Context, req *sccpb.UpdateFindingRequest, opts ...gax.CallOption) (*sccpb.Finding, error) {
			if req.Finding.Name != findingName {
				t.Fatalf("finding name = %v; want %v", req.Finding.Name, findingName)
			}
			eventTime := req.Finding.EventTime.AsTime()
			if eventTime != time.UTC() {
				t.Fatalf("finding eventTime = %v; want %v", eventTime, time.UTC())
			}
			assert.ElementsMatch(t, req.UpdateMask.Paths, []string{"event_time"}, "request update mask paths matches")
			return req.Finding, nil
		},
	}
	cli := securityCommandCenterClientImpl{ctx: context.TODO(), client: mock}
	err := cli.updateFindingEventTime(findingName, time)
	if err != nil {
		t.Fatalf("err = %v; want nil", err)
	}
}

func TestDeactivateFinding(t *testing.T) {
	findingName := "test"
	time := time.Now()
	mock := &sccApiClientMock{
		UpdateFindingFn: func(ctx context.Context, req *sccpb.UpdateFindingRequest, opts ...gax.CallOption) (*sccpb.Finding, error) {
			if req.Finding.Name != findingName {
				t.Fatalf("finding name = %v; want %v", req.Finding.Name, findingName)
			}
			if req.Finding.State != sccpb.Finding_INACTIVE {
				t.Fatalf("finding state = %v; want %v", req.Finding.State, sccpb.Finding_INACTIVE)
			}
			eventTime := req.Finding.EventTime.AsTime()
			if eventTime != time.UTC() {
				t.Fatalf("finding eventTime = %v; want %v", eventTime, time.UTC())
			}
			assert.ElementsMatch(t, req.UpdateMask.Paths, []string{"state", "event_time"}, "request update mask paths matches")
			return req.Finding, nil
		},
	}
	cli := securityCommandCenterClientImpl{ctx: context.TODO(), client: mock}
	err := cli.deactivateFinding(findingName, time)
	if err != nil {
		t.Fatalf("err = %v; want nil", err)
	}
}

func TestUpsertFinding(t *testing.T) {
	finding := &sccpb.Finding{}
	paths := []string{"path1", "path2"}
	mock := &sccApiClientMock{
		UpdateFindingFn: func(ctx context.Context, req *sccpb.UpdateFindingRequest, opts ...gax.CallOption) (*sccpb.Finding, error) {
			if req.Finding != finding {
				t.Fatalf("finding pointer = %v; want %v", req.Finding, finding)
			}
			if req.UpdateMask == nil {
				t.Fatalf("finding pointer = %v; want %v", req.Finding, finding)
			}
			if len(req.UpdateMask.Paths) != len(paths) {
				t.Fatalf("number of paths in update mask = %v; want %v", len(req.UpdateMask.Paths), len(paths))
			}
			return finding, nil
		},
	}
	cli := securityCommandCenterClientImpl{ctx: context.TODO(), client: mock}
	result, err := cli.upsertFinding(finding, paths)
	if err != nil {
		t.Fatalf("err = %v; want nil", err)
	}
	if result != finding {
		t.Fatalf("result pointer = %v; want %v", result, finding)
	}
}

func TestResourceIteratorToSlice(t *testing.T) {
	expected := []*sccpb.Source{
		{Name: "sourceOne", Description: "sourceOneDesc"},
		{Name: "sourceTwo", Description: "sourceTwoDesc"},
		nil,
	}
	errors := []error{
		nil,
		nil,
		iterator.Done,
	}
	i := 0
	nextFn := func() (res *sccpb.Source, err error) {
		res, err = expected[i], errors[i]
		i++
		return
	}
	itMock := &sccSourceIteratorMock{NextFn: nextFn}
	limit := 10
	results, err := resourceIteratorToSlice[*sccpb.Source](itMock, limit, func(s *sccpb.Source) bool { return true })
	if err != nil {
		t.Fatalf("err = %v; want nil", err)
	}
	if len(results) != 2 {
		t.Fatalf("number of results = %v; want %v", len(results), 2)
	}
	for i := range results {
		if results[i].Name != expected[i].Name {
			t.Errorf("result [%d] name = %v; want %v", i, results[i].Name, expected[i].Name)
		}
		if results[i].Description != expected[i].Description {
			t.Errorf("result [%d] description = %v; want %v", i, results[i].Description, expected[i].Description)
		}
	}
}

func TestMapFindingSeverityString(t *testing.T) {
	data := map[string]sccpb.Finding_Severity{
		FINDING_SEVERITY_STRING_CRITICAL: sccpb.Finding_CRITICAL,
		FINDING_SEVERITY_STRING_HIGH:     sccpb.Finding_HIGH,
		FINDING_SEVERITY_STRING_MEDIUM:   sccpb.Finding_MEDIUM,
		FINDING_SEVERITY_STRING_LOW:      sccpb.Finding_LOW,
		"bogus":                          sccpb.Finding_SEVERITY_UNSPECIFIED,
		"":                               sccpb.Finding_SEVERITY_UNSPECIFIED,
	}
	for k, v := range data {
		r := mapFindingSeverityString(k)
		if r != v {
			t.Errorf("severity of %v = %v; want %v", k, r, v)
		}
	}
}
