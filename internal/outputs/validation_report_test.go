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
	"errors"
	"testing"

	"github.com/google/gke-policy-automation/internal/policy"
	"github.com/stretchr/testify/assert"
)

type validationReportMapperMock struct {
	addResultFn     func(result *policy.PolicyEvaluationResult)
	addResultsFn    func(results []*policy.PolicyEvaluationResult)
	getReportFn     func() *ValidationReport
	getJsonReportFn func() ([]byte, error)
}

func (m validationReportMapperMock) AddResult(result *policy.PolicyEvaluationResult) {
	m.addResultFn(result)
}

func (m validationReportMapperMock) AddResults(results []*policy.PolicyEvaluationResult) {
	m.addResultsFn(results)
}

func (m validationReportMapperMock) GetReport() *ValidationReport {
	return m.getReportFn()
}

func (m validationReportMapperMock) GetJsonReport() ([]byte, error) {
	return m.getJsonReportFn()
}

func TestGetReport(t *testing.T) {
	clusterOneName := "cluster-one"
	clusterTwoName := "cluster-two"
	policies := []*policy.Policy{
		{
			Name:        "policy-one",
			Title:       "policy-one-title",
			Description: "policy-one-desc",
			Group:       "group",
		},
		{
			Name:        "policy-two",
			Title:       "policy-two-title",
			Description: "policy-two-desc",
			Group:       "group",
		},
	}
	expectedClusterEvaluations := [][]*ValidationReportClusterEvaluation{
		{
			{ClusterID: clusterOneName, Valid: true, ProcessingErrors: []string{}},
			{ClusterID: clusterTwoName, Violations: []string{"violation"}, ProcessingErrors: []string{}},
		},
		{
			{ClusterID: clusterOneName, Violations: []string{"violation"}, ProcessingErrors: []string{}},
			{ClusterID: clusterTwoName, Violations: []string{"violation"}, ProcessingErrors: []string{}},
		},
	}
	mapper := NewValidationReportMapper()
	mapper.AddResults([]*policy.PolicyEvaluationResult{
		{
			ClusterName: clusterOneName,
			Policies: []*policy.Policy{
				{
					Name:        policies[0].Name,
					Title:       policies[0].Title,
					Description: policies[0].Description,
					Group:       policies[0].Group,
					Valid:       true,
				}, {
					Name:        policies[1].Name,
					Title:       policies[1].Title,
					Description: policies[1].Description,
					Group:       policies[1].Group,
					Valid:       false,
					Violations:  []string{"violation"},
				},
			},
		},
		{
			ClusterName: clusterTwoName,
			Policies: []*policy.Policy{
				{
					Name:        policies[0].Name,
					Title:       policies[0].Title,
					Description: policies[0].Description,
					Group:       policies[0].Group,
					Valid:       false,
					Violations:  []string{"violation"},
				}, {
					Name:        policies[1].Name,
					Title:       policies[1].Title,
					Description: policies[1].Description,
					Group:       policies[1].Group,
					Valid:       false,
					Violations:  []string{"violation"},
				},
			},
		},
	})
	report := mapper.GetReport()
	if len(report.Policies) != 2 {
		t.Fatalf("number of policies in a report = %v; want %v", len(report.Policies), 2)
	}
	if len(report.ClusterStats) != 2 {
		t.Fatalf("number of clusterStats in a report = %v; want %v", len(report.ClusterStats), 2)
	}
	for i := range policies {
		assert.Contains(t, report.Policies, &ValidationReportPolicy{
			PolicyName:         policies[i].Name,
			PolicyGroup:        policies[i].Group,
			PolicyTitle:        policies[i].Title,
			PolicyDescription:  policies[i].Description,
			ClusterEvaluations: expectedClusterEvaluations[i],
		}, "report policies contains valid policy %v", policies[0].Name)
	}
	assert.Contains(t, report.ClusterStats, &ValidationReportClusterStats{
		ClusterID:             clusterOneName,
		ValidPoliciesCount:    1,
		ViolatedPoliciesCount: 1,
	}, "report cluster stats contains valid stats for cluster %v", clusterOneName)
	assert.Contains(t, report.ClusterStats, &ValidationReportClusterStats{
		ClusterID:             clusterTwoName,
		ViolatedPoliciesCount: 2,
	}, "report cluster stats contains valid stats for cluster %v", clusterTwoName)
}

func TestMapErrorSliceToStringSlice(t *testing.T) {
	errors := []error{errors.New("error-one"), errors.New("error-two"), errors.New("error-three")}
	expected := []string{"error-one", "error-two", "error-three"}
	result := mapErrorSliceToStringSlice(errors)
	assert.ElementsMatch(t, expected, result, "mapped slice of strings matches")
}
