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
	"encoding/json"
	"sort"
	"strings"
	"time"

	"github.com/google/gke-policy-automation/internal/policy"
)

const (
	SeverityCritical = 4
	SeverityHigh     = 3
	SeverityMedium   = 2
	SeverityLow      = 1
	SeverityUnknown  = 0
)

type ValidationReport struct {
	ValidationTime time.Time                       `json:"validationDate"`
	Policies       []*ValidationReportPolicy       `json:"policies"`
	ClusterStats   []*ValidationReportClusterStats `json:"statistics"`
}

type ValidationReportPolicy struct {
	PolicyName         string                               `json:"name"`
	PolicyGroup        string                               `json:"group"`
	PolicyTitle        string                               `json:"title"`
	PolicyDescription  string                               `json:"description"`
	Recommendation     string                               `json:"recommendation,omitempty"`
	ExternalURI        string                               `json:"externalURI,omitempty"`
	Severity           string                               `json:"severity,omitempty"`
	SeverityNumber     int                                  `json:"-"`
	ClusterEvaluations []*ValidationReportClusterEvaluation `json:"clusters"`
}

type ValidationReportClusterEvaluation struct {
	ClusterID        string   `json:"cluster"`
	Valid            bool     `json:"isValid"`
	Errored          bool     `json:"isErrored"`
	Violations       []string `json:"violations,omitempty"`
	ProcessingErrors []string `json:"errors,omitempty"`
}

type ValidationReportClusterStats struct {
	ClusterID             string `json:"cluster"`
	ValidPoliciesCount    int    `json:"validPoliciesCount"`
	ViolatedPoliciesCount int    `json:"violatedPoliciesCount"`
	ErroredPoliciesCount  int    `json:"erroredPoliciesCount"`
	ViolatedCriticalCount int    `json:"violatedCriticalCount"`
	ViolatedHighCount     int    `json:"violatedHighCount"`
	ViolatedMediumCount   int    `json:"violatedMediumCount"`
	ViolatedLowCount      int    `json:"violatedLowCount"`
}

type ValidationReportMapper interface {
	AddResult(result *policy.PolicyEvaluationResult)
	AddResults(results []*policy.PolicyEvaluationResult)
	GetReport() *ValidationReport
	GetJSONReport() ([]byte, error)
}

type validationReportMapperImpl struct {
	validationTime  time.Time
	policies        map[string]*ValidationReportPolicy
	clusterStats    map[string]*ValidationReportClusterStats
	jsonMarshalFunc func(v any) ([]byte, error)
}

func NewValidationReportMapper() ValidationReportMapper {
	return &validationReportMapperImpl{
		policies:        make(map[string]*ValidationReportPolicy),
		clusterStats:    make(map[string]*ValidationReportClusterStats),
		jsonMarshalFunc: json.Marshal,
		validationTime:  time.Now(),
	}
}

func (m *validationReportMapperImpl) AddResult(result *policy.PolicyEvaluationResult) {
	clusterStat, ok := m.clusterStats[result.ClusterID]
	if !ok {
		clusterStat = &ValidationReportClusterStats{ClusterID: result.ClusterID}
		m.clusterStats[result.ClusterID] = clusterStat
	}
	for _, resultPolicy := range result.Policies {
		reportPolicy, ok := m.policies[resultPolicy.Name]
		if !ok {
			reportPolicy = mapResultPolicyToReportPolicy(resultPolicy)
			m.policies[resultPolicy.Name] = reportPolicy
		}
		clusterEvaluation := mapResultPolicyToReportClusterEvaluation(resultPolicy, result.ClusterID)
		reportPolicy.ClusterEvaluations = append(reportPolicy.ClusterEvaluations, clusterEvaluation)
		if clusterEvaluation.Errored {
			clusterStat.ErroredPoliciesCount++
		} else {
			if clusterEvaluation.Valid {
				clusterStat.ValidPoliciesCount++
			} else {
				clusterStat.ViolatedPoliciesCount++
				switch strings.ToLower(resultPolicy.Severity) {
				case "critical":
					clusterStat.ViolatedCriticalCount++
				case "high":
					clusterStat.ViolatedHighCount++
				case "medium":
					clusterStat.ViolatedMediumCount++
				default:
					clusterStat.ViolatedLowCount++
				}
			}
		}
	}
}

func (m *validationReportMapperImpl) AddResults(results []*policy.PolicyEvaluationResult) {
	for _, result := range results {
		m.AddResult(result)
	}
}

func (m *validationReportMapperImpl) GetReport() *ValidationReport {
	policies := make([]*ValidationReportPolicy, 0, len(m.policies))
	for _, policy := range m.policies {
		policies = append(policies, policy)
	}
	sort.SliceStable(policies, func(i, j int) bool {
		/*
			if policies[i].PolicyGroup == policies[j].PolicyGroup {
				return policies[i].PolicyName < policies[j].PolicyName
			}
			return policies[i].PolicyGroup < policies[j].PolicyGroup
		*/
		if policies[i].SeverityNumber == policies[j].SeverityNumber {
			if policies[i].PolicyGroup == policies[j].PolicyGroup {
				return policies[i].PolicyName < policies[j].PolicyName
			}
			return policies[i].PolicyGroup < policies[j].PolicyGroup
		}
		return policies[i].SeverityNumber > policies[j].SeverityNumber
	})
	stats := make([]*ValidationReportClusterStats, 0, len(m.clusterStats))
	for _, stat := range m.clusterStats {
		stats = append(stats, stat)
	}
	return &ValidationReport{
		ValidationTime: m.validationTime,
		Policies:       policies,
		ClusterStats:   stats,
	}
}

func (m *validationReportMapperImpl) GetJSONReport() ([]byte, error) {
	report := m.GetReport()
	return m.jsonMarshalFunc(report)
}

func mapResultPolicyToReportPolicy(policy *policy.Policy) *ValidationReportPolicy {
	reportPolicy := &ValidationReportPolicy{
		PolicyName:        policy.Name,
		PolicyTitle:       policy.Title,
		PolicyDescription: policy.Description,
		PolicyGroup:       policy.Group,
		Recommendation:    policy.Recommendation,
		ExternalURI:       policy.ExternalURI,
		Severity:          policy.Severity,
		SeverityNumber:    mapSeverityToNumber(policy.Severity),
	}
	return reportPolicy
}

func mapResultPolicyToReportClusterEvaluation(policy *policy.Policy, clusterName string) *ValidationReportClusterEvaluation {
	clusterEvaluation := &ValidationReportClusterEvaluation{
		ClusterID:        clusterName,
		Valid:            policy.Valid,
		Violations:       policy.Violations,
		ProcessingErrors: mapErrorSliceToStringSlice(policy.ProcessingErrors),
	}

	if len(clusterEvaluation.ProcessingErrors) > 0 {
		clusterEvaluation.Errored = true
	}
	return clusterEvaluation
}

func mapErrorSliceToStringSlice(errors []error) []string {
	strings := make([]string, len(errors))
	for i := range errors {
		strings[i] = errors[i].Error()
	}
	return strings
}

func mapSeverityToNumber(severity string) int {
	switch strings.ToLower(severity) {
	case "critical":
		return SeverityCritical
	case "high":
		return SeverityHigh
	case "medium":
		return SeverityMedium
	case "low":
		return SeverityLow
	default:
		return SeverityUnknown
	}
}
