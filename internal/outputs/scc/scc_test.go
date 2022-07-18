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
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"testing"
	"time"

	scc "cloud.google.com/go/securitycenter/apiv1"
	gax "github.com/googleapis/gax-go/v2"
	"github.com/stretchr/testify/assert"
	"google.golang.org/api/iterator"
	sccpb "google.golang.org/genproto/googleapis/cloud/securitycenter/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type sccApiClientMock struct {
	ListSourcesFn   func(ctx context.Context, req *sccpb.ListSourcesRequest, opts ...gax.CallOption) *scc.SourceIterator
	CreateSourceFn  func(ctx context.Context, req *sccpb.CreateSourceRequest, opts ...gax.CallOption) (*sccpb.Source, error)
	ListFindingsFn  func(ctx context.Context, req *sccpb.ListFindingsRequest, opts ...gax.CallOption) *scc.ListFindingsResponse_ListFindingsResultIterator
	UpdateFindingFn func(ctx context.Context, req *sccpb.UpdateFindingRequest, opts ...gax.CallOption) (*sccpb.Finding, error)
	CloseFn         func() error
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

func (m *sccApiClientMock) Close() error {
	return m.CloseFn()
}

type sccSourceIteratorMock struct {
	NextFn func() (*sccpb.Source, error)
}

func (m *sccSourceIteratorMock) Next() (*sccpb.Source, error) {
	return m.NextFn()
}

func TestCreateSource(t *testing.T) {
	orgNumber := "123456789"
	srcName := "testSourceName"
	mock := &sccApiClientMock{
		CreateSourceFn: func(ctx context.Context, req *sccpb.CreateSourceRequest, opts ...gax.CallOption) (*sccpb.Source, error) {
			if req.Source.DisplayName != sourceDisplayName {
				t.Errorf("req source display name = %v; want %v", req.Source.DisplayName, sourceDisplayName)
			}
			if req.Source.Description != sourceDescription {
				t.Errorf("req source description = %v; want %v", req.Source.Description, sourceDescription)
			}
			return &sccpb.Source{Name: srcName}, nil
		},
	}
	cli := securityCommandCenterClientImpl{ctx: context.TODO(), client: mock}
	cli.organizationNumber = orgNumber
	result, err := cli.CreateSource()
	if err != nil {
		t.Fatalf("err = %v; want nil", err)
	}
	if result != srcName {
		t.Errorf("result source name = %v; want %v", result, srcName)
	}
}

func TestUpsertFinding_emptySearch_active(t *testing.T) {
	source := "source"
	finding := &Finding{
		ResourceName: "resource",
		Category:     "category",
		State:        FINDING_STATE_STRING_ACTIVE,
	}
	mock := &sccApiClientMock{
		ListFindingsFn: func(ctx context.Context, req *sccpb.ListFindingsRequest, opts ...gax.CallOption) *scc.ListFindingsResponse_ListFindingsResultIterator {
			return &scc.ListFindingsResponse_ListFindingsResultIterator{}
		},
		UpdateFindingFn: func(ctx context.Context, req *sccpb.UpdateFindingRequest, opts ...gax.CallOption) (*sccpb.Finding, error) {
			if req.Finding.Parent != source {
				t.Fatalf("new finding parent = %v; want %v", req.Finding.Parent, source)
			}
			if req.Finding.ResourceName != finding.ResourceName {
				t.Fatalf("new finding resourceName = %v; want %v", req.Finding.ResourceName, finding.ResourceName)
			}
			if req.Finding.Category != finding.Category {
				t.Fatalf("new finding category = %v; want %v", req.Finding.Category, finding.Category)
			}
			if req.Finding.State.String() != finding.State {
				t.Fatalf("new finding state  = %v; want %v", req.Finding.State.String(), finding.State)
			}
			return &sccpb.Finding{}, nil
		},
	}
	c := securityCommandCenterClientImpl{client: mock, sourcesSearchLimit: 0}
	err := c.UpsertFinding(source, finding)
	if err != nil {
		t.Fatalf("err = %v; want nil", err)
	}
}

func TestUpsertFinding_emptySearch_inactive(t *testing.T) {
	source := "source"
	finding := &Finding{
		ResourceName: "resource",
		Category:     "category",
		State:        FINDING_STATE_STRING_INACTIVE,
	}
	mock := &sccApiClientMock{
		ListFindingsFn: func(ctx context.Context, req *sccpb.ListFindingsRequest, opts ...gax.CallOption) *scc.ListFindingsResponse_ListFindingsResultIterator {
			return &scc.ListFindingsResponse_ListFindingsResultIterator{}
		},
		UpdateFindingFn: func(ctx context.Context, req *sccpb.UpdateFindingRequest, opts ...gax.CallOption) (*sccpb.Finding, error) {
			t.Fatal("update finding was called on non-existing, inactive finding")
			return nil, nil

		},
	}
	c := securityCommandCenterClientImpl{client: mock, sourcesSearchLimit: 0}
	err := c.UpsertFinding(source, finding)
	if err != nil {
		t.Fatalf("err = %v; want nil", err)
	}
}

func TestClose(t *testing.T) {
	err := errors.New("test error")
	mock := &sccApiClientMock{
		CloseFn: func() error {
			return err
		},
	}
	c := securityCommandCenterClientImpl{client: mock}
	result := c.Close()
	if result != err {
		t.Errorf("result = %v; want %v", result, err)
	}
}

func TestFindSource(t *testing.T) {
	orgNumber := "123456789"
	mock := &sccApiClientMock{
		ListSourcesFn: func(ctx context.Context, req *sccpb.ListSourcesRequest, opts ...gax.CallOption) *scc.SourceIterator {
			expectedParent := fmt.Sprintf("organizations/%s", orgNumber)
			if req.Parent != expectedParent {
				t.Fatalf("request parent = %v; want %v", req.Parent, expectedParent)
			}
			return &scc.SourceIterator{}
		},
	}
	cli := securityCommandCenterClientImpl{ctx: context.TODO(), client: mock, sourcesSearchLimit: 0}
	cli.organizationNumber = orgNumber
	_, err := cli.FindSource()
	if err != nil {
		t.Fatalf("err = %v; want nil", err)
	}
}

func TestFindSourceNameByDisplayName(t *testing.T) {
	name := "sourceOne"
	displayName := "sourceOneDisplayName"
	expected := []*sccpb.Source{
		{Name: name, DisplayName: displayName},
		{Name: "sourceTwo", DisplayName: "sourceTwoDisplayName"},
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
	cli := securityCommandCenterClientImpl{ctx: context.TODO(), sourcesSearchLimit: 3}
	result, err := cli.findSourceNameByDisplayName(displayName, itMock)
	if err != nil {
		t.Fatalf("err = %v; want nil", err)
	}
	if result == nil {
		t.Fatalf("result is nil; want %s", name)
	}
	if *result != name {
		t.Errorf("result is %s; want %s", *result, name)
	}
}

func TestGetFindings(t *testing.T) {
	source := "source"
	resource := "resource"
	category := "category"
	mock := &sccApiClientMock{
		ListFindingsFn: func(ctx context.Context, req *sccpb.ListFindingsRequest, opts ...gax.CallOption) *scc.ListFindingsResponse_ListFindingsResultIterator {
			if req.Parent != source {
				t.Errorf("parent = %v; want %v", req.Parent, source)
			}
			r := regexp.MustCompile("resourceName=\"(.+)\" AND category=\"(.+)\"")
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
			return &scc.ListFindingsResponse_ListFindingsResultIterator{}
		},
	}
	cli := securityCommandCenterClientImpl{ctx: context.TODO(), client: mock, sourcesSearchLimit: 0}
	cli.getFindings(source, resource, category)
}

func TestCreateFinding(t *testing.T) {
	sourceName := "sourceName"
	finding := &Finding{
		Time:              time.Now(),
		ResourceName:      "cluster-resource",
		Category:          "category",
		Description:       "description",
		State:             FINDING_STATE_STRING_ACTIVE,
		Severity:          FINDING_SEVERITY_STRING_HIGH,
		SourcePolicyName:  "gke.policy.some_policy",
		SourcePolicyGroup: "Security",
		SourcePolicyFile:  "name.rego",
		CisVersion:        "1.0",
		CisID:             "1.2.3",
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
			expectedSrcProperties := mapFindingSourceProperties(finding)
			if !reflect.DeepEqual(req.Finding.SourceProperties, expectedSrcProperties) {
				t.Errorf("finding sourceProperties = %v; want %v", req.Finding.SourceProperties, expectedSrcProperties)
			}
			expectedCompliances := mapFindingCompliances(finding)
			if !reflect.DeepEqual(req.Finding.Compliances, expectedCompliances) {
				t.Errorf("finding compliances = %v; want %v", req.Finding.Compliances, expectedCompliances)
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

func TestUpdateFinding(t *testing.T) {
	finding := &Finding{
		Time:              time.Now(),
		SourcePolicyName:  "gke.policy.test",
		SourcePolicyFile:  "file.rego",
		SourcePolicyGroup: "group",
		CisVersion:        "1.0",
		CisID:             "1.4.5",
	}
	resultFinding := &sccpb.Finding{
		Name:      "finding",
		EventTime: timestamppb.New(finding.Time),
		State:     sccpb.Finding_ACTIVE,
	}
	mock := &sccApiClientMock{
		UpdateFindingFn: func(ctx context.Context, req *sccpb.UpdateFindingRequest, opts ...gax.CallOption) (*sccpb.Finding, error) {
			if req.Finding.Name != resultFinding.Name {
				t.Fatalf("finding name = %v; want %v", req.Finding.Name, resultFinding.Name)
			}
			if req.Finding.EventTime != resultFinding.EventTime {
				t.Fatalf("finding eventTime = %v; want %v", req.Finding.EventTime, resultFinding.EventTime)
			}
			if req.Finding.State != resultFinding.State {
				t.Fatalf("finding state = %v; want %v", req.Finding.State, resultFinding.State)
			}
			expectedSrcProperties := mapFindingSourceProperties(finding)
			if !reflect.DeepEqual(req.Finding.SourceProperties, expectedSrcProperties) {
				t.Errorf("finding sourceProperties = %v; want %v", req.Finding.SourceProperties, expectedSrcProperties)
			}
			assert.ElementsMatch(t, req.UpdateMask.Paths, []string{"state", "event_time", "source_properties"}, "request update mask paths matches")
			return req.Finding, nil
		},
	}
	cli := securityCommandCenterClientImpl{ctx: context.TODO(), client: mock}
	err := cli.updateFinding(resultFinding, finding)
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

func TestMapFindingSourceProperties(t *testing.T) {
	finding := &Finding{
		SourcePolicyName:  "name",
		SourcePolicyFile:  "file",
		SourcePolicyGroup: "group",
		CisVersion:        "1.2",
		CisID:             "6.9.1",
	}
	result := mapFindingSourceProperties(finding)
	expectedPolicyName := fmt.Sprintf("string_value:%q", finding.SourcePolicyName)
	if result["PolicyName"].String() != expectedPolicyName {
		t.Errorf("PolicyName = %v; want %v", result["PolicyName"].String(), expectedPolicyName)
	}
	expectedPolicyFile := fmt.Sprintf("string_value:%q", finding.SourcePolicyFile)
	if result["PolicyFile"].String() != expectedPolicyFile {
		t.Errorf("PolicyFile = %v; want %v", result["PolicyFile"].String(), expectedPolicyFile)
	}
	expectedPolicyGroup := fmt.Sprintf("string_value:%q", finding.SourcePolicyGroup)
	if result["PolicyGroup"].String() != expectedPolicyGroup {
		t.Errorf("PolicyGroup = %v; want %v", result["PolicyGroup"].String(), expectedPolicyGroup)
	}
	complianceStruct := result["compliance_standards"].GetStructValue()
	if complianceStruct == nil {
		t.Fatalf("result compliance_standards struct is nil")
	}
	cisList := complianceStruct.Fields["cis"].GetListValue()
	if cisList == nil {
		t.Fatalf("result compliance_standards struct is nil")
	}
	if len(cisList.Values) < 1 {
		t.Fatalf("result compliance_standards has empty or nil cis value")
	}
	cisElementStruct := cisList.Values[0].GetStructValue()
	if cisElementStruct == nil {
		t.Fatalf("result compliance_standards cis element 0 is nil")
	}
	version := cisElementStruct.Fields["version"].GetStringValue()
	if version != finding.CisVersion {
		t.Errorf("result compliance_standards cis element 0 version = %v; want %v", version, finding.CisVersion)
	}
	idList := cisElementStruct.Fields["ids"].GetListValue()
	if idList == nil {
		t.Fatalf("result compliance_standards cis element 0 ids is nil")
	}
	id := idList.Values[0].GetStringValue()
	if id != finding.CisID {
		t.Errorf("result compliance_standards cis element 0 ids element 0 = %v; want %v", id, finding.CisID)
	}
}

func TestMapFindingCompliances_positive(t *testing.T) {
	finding := &Finding{CisVersion: "1.0", CisID: "6.1.2"}
	expected := []*sccpb.Compliance{{Standard: "cis", Version: "1.0", Ids: []string{"6.1.2"}}}

	result := mapFindingCompliances(finding)
	if len(result) != 1 {
		t.Errorf("result has %v elements; want %v", len(result), 1)
	}
	if !reflect.DeepEqual(*result[0], *expected[0]) {
		t.Errorf("result element =  %v; want %v", *result[0], *expected[0])
	}
}

func TestMapFindingCompliances_negative(t *testing.T) {
	findings := []*Finding{
		{CisVersion: "1.0"},
		{CisID: "2.2.1"},
		{},
	}
	for i := range findings {
		result := mapFindingCompliances(findings[i])
		if result != nil {
			t.Errorf("result = %v; want nil", result[i])
		}
	}
}
