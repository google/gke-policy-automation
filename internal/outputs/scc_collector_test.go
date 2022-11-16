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
	if sccCol.maxGoRoutines != defaultMaxCoroutines {
		t.Errorf("collector threadsNo = %v; want %v", sccCol.maxGoRoutines, defaultMaxCoroutines)
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
	c := &sccCollector{cli: mock, findings: findings, maxGoRoutines: 2}
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
			ClusterID: "testOne",
			Policies: []*policy.Policy{
				{Valid: true, Category: "categoryOne", Severity: "LOW", Description: "description"},
				{Violations: []string{"violation"}, Category: "categoryTwo", Severity: "LOW", Description: "description"},
			},
		},
		{
			ClusterID: "testTwo",
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
	}
	c := &sccCollector{cli: &mock, maxGoRoutines: 1}
	findingsChan := make(chan *scc.Finding, c.maxGoRoutines)
	errorsChan := make(chan error, c.maxGoRoutines)
	for _, finding := range findings {
		findingsChan <- finding
	}
	close(findingsChan)

	var wg sync.WaitGroup
	wg.Add(1)
	c.upsertFinding(0, &wg, findingsChan, sourceName, errorsChan)
	wg.Wait()
	close(errorsChan)
	assert.ElementsMatch(t, findings, upserted, "upserted findings match test findings")
	if len(errorsChan) != 0 {
		t.Errorf("number of errors in result channel = %v; want %v", len(errorsChan), 0)
	}
}

func TestGetSccSource_existing(t *testing.T) {
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
		Violations:     []string{"invalid"},
		Category:       "category",
		Severity:       "LOW",
		Description:    "description",
		Group:          "Security",
		File:           "file.rego",
		Name:           "gke.policy.test",
		CisVersion:     "1.2",
		CisID:          "6.2.3",
		ExternalURI:    "https://external-uri",
		Recommendation: "A good recommendation",
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
	if finding.SourcePolicyName != policy.Name {
		t.Errorf("finding sourcePolicyName = %v; want %v", finding.SourcePolicyName, policy.Name)
	}
	if finding.SourcePolicyGroup != policy.Group {
		t.Errorf("finding sourcePolicyGroup = %v; want %v", finding.SourcePolicyGroup, policy.Group)
	}
	if finding.SourcePolicyFile != policy.File {
		t.Errorf("finding sourcePolicyFile = %v; want %v", finding.SourcePolicyFile, policy.File)
	}
	if finding.CisID != policy.CisID {
		t.Errorf("finding cisID = %v; want %v", finding.CisID, policy.CisID)
	}
	if finding.CisVersion != policy.CisVersion {
		t.Errorf("finding cisVersion = %v; want %v", finding.CisVersion, policy.CisVersion)
	}
	if finding.State != scc.FINDING_STATE_STRING_ACTIVE {
		t.Errorf("finding state = %v; want %v", finding.State, scc.FINDING_STATE_STRING_ACTIVE)
	}
	if finding.Severity != scc.FINDING_SEVERITY_STRING_LOW {
		t.Errorf("finding severity = %v; want %v", finding.Severity, scc.FINDING_SEVERITY_STRING_LOW)
	}
	if finding.ExternalURI != policy.ExternalURI {
		t.Errorf("finding externalURI = %v; want %v", finding.ExternalURI, policy.ExternalURI)
	}
	if finding.Recommendation != policy.Recommendation {
		t.Errorf("finding recommendation = %v; want %v", finding.Recommendation, policy.Recommendation)
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
