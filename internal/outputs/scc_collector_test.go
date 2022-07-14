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

package outputs

import (
	"testing"
	"time"

	"github.com/google/gke-policy-automation/internal/outputs/scc"
	"github.com/google/gke-policy-automation/internal/policy"
)

type sccClientMock struct {
	CreateSourceFn  func() (string, error)
	FindSourceFn    func() (*string, error)
	UpsertFindingFn func(sourceName string, finding *scc.Finding) error
}

func (m *sccClientMock) CreateSource() (string, error) {
	return m.CreateSourceFn()
}

func (m *sccClientMock) FindSource() (*string, error) {
	return m.FindSourceFn()
}

func (m *sccClientMock) UpsertFinding(sourceName string, finding *scc.Finding) error {
	return m.UpsertFindingFn(sourceName, finding)
}

func TestSccCollectorRegisterResult(t *testing.T) {
	results := []*policy.PolicyEvaluationResult{
		{
			ClusterName: "testOne",
			Policies: []*policy.Policy{
				{Valid: true, Category: "categoryOne", Severity: "LOW", Description: "description"},
				{Violations: []string{"violation"}, Category: "categoryTwo", Severity: "LOW", Description: "description"},
			},
		},
		{
			ClusterName: "testTwo",
			Policies: []*policy.Policy{
				{Valid: true, Category: "categoryOne", Severity: "LOW", Description: "description"},
				{Violations: []string{"violation"}, Category: "categoryTwo", Severity: "LOW", Description: "description"},
			},
		},
	}
	c := &sccCollector{}
	err := c.RegisterResult(results)
	if err != nil {
		t.Fatalf("err = %v; want nil", err)
	}
	if len(c.findings) != 4 {
		t.Fatalf("number of findings = %v; want %v", len(c.findings), 4)
	}
}

func TestGetSccSource_exisitng(t *testing.T) {
	sourceName := "testSourceName"
	mock := &sccClientMock{
		FindSourceFn: func() (*string, error) {
			return &sourceName, nil
		},
	}
	col := &sccCollector{cli: mock, createSource: true}
	result, err := col.getSccSource()
	if err != nil {
		t.Fatalf("err = %v; want nil", err)
	}
	if result != sourceName {
		t.Fatalf("result = %v; want %v", result, sourceName)
	}
}

func TestGetSccSource_missing(t *testing.T) {
	sourceName := "testSourceName"
	mock := &sccClientMock{
		FindSourceFn: func() (*string, error) {
			return nil, nil
		},
		CreateSourceFn: func() (string, error) {
			return sourceName, nil
		},
	}
	col := &sccCollector{cli: mock, createSource: true}
	result, err := col.getSccSource()
	if err != nil {
		t.Fatalf("err = %v; want nil", err)
	}
	if result != sourceName {
		t.Fatalf("result = %v; want %v", result, sourceName)
	}
}

func TestMapPolicyToFinding(t *testing.T) {
	resourceName := "projects/tst/locations/europe-west3/clusters/tst"
	time := time.Now()
	policy := &policy.Policy{
		Violations:  []string{"invalid"},
		Category:    "category",
		Severity:    "LOW",
		Description: "description",
	}
	finding := mapPolicyToFinding(resourceName, time, policy)
	if finding.Time != time {
		t.Errorf("finding time = %v; want %v", finding.Time, time)
	}
	expectedResName := "//container.googleapis.com/" + resourceName
	if finding.ResourceName != expectedResName {
		t.Errorf("finding resource name = %v; want %v", expectedResName, resourceName)
	}
	if finding.Category != policy.Category {
		t.Errorf("finding category = %v; want %v", finding.Category, policy.Category)
	}
	if finding.Description != policy.Description {
		t.Errorf("finding description = %v; want %v", finding.Description, policy.Description)
	}
	if finding.State != scc.FINDING_STATE_STRING_ACTIVE {
		t.Errorf("finding state = %v; want %v", finding.State, scc.FINDING_STATE_STRING_ACTIVE)
	}
	if finding.Severity != scc.FINDING_SEVERITY_STRING_LOW {
		t.Errorf("finding severity = %v; want %v", finding.Severity, scc.FINDING_SEVERITY_STRING_LOW)
	}
}

func TestMapPolicyEvaluationToFindingState(t *testing.T) {
	policies := []*policy.Policy{
		{Valid: true},
		{Violations: []string{"violation"}},
		{},
	}
	results := []string{
		scc.FINDING_STATE_STRING_INACTIVE,
		scc.FINDING_STATE_STRING_ACTIVE,
		scc.FINDING_STATE_STRING_UNSPECIFIED,
	}
	for i := range policies {
		result := mapPolicyEvaluationToFindingState(policies[i])
		if result != results[i] {
			t.Errorf("policy [%d] result = %v; want %v", i, result, results[i])
		}
	}
}
