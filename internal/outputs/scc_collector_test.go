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
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/gke-policy-automation/internal/outputs/scc"
	"github.com/google/gke-policy-automation/internal/policy"
	"github.com/stretchr/testify/assert"
)

type sccClientMock struct {
	CreateSourceFn  func() (string, error)
	FindSourceFn    func() (*string, error)
	UpsertFindingFn func(sourceName string, finding *scc.Finding) error
	CloseFn         func() error
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

func (m *sccClientMock) Close() error {
	return m.CloseFn()
}

func TestNewSccCollector(t *testing.T) {
	ctx := context.TODO()
	createSource := true
	mock := &sccClientMock{}
	col := newSccCollector(ctx, createSource, mock)

	sccCol, ok := col.(*sccCollector)
	if !ok {
		t.Fatalf("created collector is not *sccCollector")
	}
	if sccCol.createSource != createSource {
		t.Errorf("collector createSource = %v; want %v", sccCol.createSource, createSource)
	}
	if defaultNoThreads < 2 {
		t.Errorf("defaultNoThreads = %v; want at least 2", defaultNoThreads)
	}
	if sccCol.threadsNo != defaultNoThreads {
		t.Errorf("collector threadsNo = %v; want %v", sccCol.threadsNo, defaultNoThreads)
	}
}

func TestGetName(t *testing.T) {
	c := sccCollector{}
	result := c.Name()
	if result != collectorName {
		t.Errorf("name = %v; want %v", result, collectorName)
	}
}

func TestProcessFindings(t *testing.T) {
	findings := []*scc.Finding{
		{Category: "category1"},
		{Category: "category2"},
		{Category: "category3"},
	}
	findingsToErr := map[string]error{
		"category1": errors.New("error1"),
		"category2": errors.New("error2"),
		"category3": errors.New("error3"),
	}
	source := "src"
	mock := &sccClientMock{
		UpsertFindingFn: func(sourceName string, finding *scc.Finding) error {
			if sourceName != source {
				t.Fatalf("upsert fn source = %v; want %v", sourceName, source)
			}
			return findingsToErr[finding.Category]
		},
	}
	c := &sccCollector{cli: mock, findings: findings, threadsNo: 2}
	result := c.processFindings(source)
	if len(result) != len(findingsToErr) {
		t.Fatalf("number of results = %v; want %v", len(result), len(findingsToErr))
	}
	for _, v := range findingsToErr {
		assert.Containsf(t, result, v, "result contains error %v", v)
	}
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

func TestUpsertFindings(t *testing.T) {
	sourceName := "testSource"
	var upserted []*scc.Finding
	mock := sccClientMock{
		UpsertFindingFn: func(sourceName string, finding *scc.Finding) error {
			upserted = append(upserted, finding)
			return nil
		},
	}
	findings := []*scc.Finding{
		{Category: "test-one"},
		{Category: "test-two"},
		{Category: "test-three"},
	}
	var wg sync.WaitGroup
	wg.Add(1)
	ch := make(chan error)
	c := &sccCollector{cli: &mock}
	c.upsertFindings(0, &wg, findings, sourceName, ch)
	wg.Wait()
	close(ch)
	assert.ElementsMatch(t, findings, upserted, "upserted findings match test findings")
	if len(ch) != 0 {
		t.Errorf("number of errors in result channel = %v; want %v", len(ch), 0)
	}
}

func TestDivideFindings(t *testing.T) {
	chunksNo := 7
	var findings []*scc.Finding
	for i := 0; i < 24; i++ {
		findings = append(findings, &scc.Finding{
			Category: fmt.Sprintf("test-%d", i),
		})
	}
	result := divideFindings(chunksNo, findings)
	var mergedBack []*scc.Finding
	for i := range result {
		for j := range result[i] {
			mergedBack = append(mergedBack, result[i][j])
		}
	}
	if len(mergedBack) != len(findings) {
		t.Errorf("sum len of slices in result = %v; want %v", len(mergedBack), len(findings))
	}
	for i := range findings {
		assert.Containsf(t, mergedBack, findings[i], "result contains finding %+v", findings[i])
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
