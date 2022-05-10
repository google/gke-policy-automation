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
	"time"

	"github.com/google/gke-policy-automation/internal/policy"
)

func MapEvaluationResultsToJsonWithTime(evaluationResult []*policy.PolicyEvaluationResult, time time.Time) ([]byte, error) {

	validationResults := ValidationResults{
		ValidationDate: time,
	}

	for _, r := range evaluationResult {
		validationResults.ClusterValidationResults = append(validationResults.ClusterValidationResults, MapClusterToJson(r))
	}

	res, err := json.Marshal(validationResults)

	if err != nil {
		return nil, err
	}

	return res, nil
}

func MapClusterToJson(evaluationResult *policy.PolicyEvaluationResult) ClusterValidationResult {

	policyList := make([]PolicyValidationResult, 0)
	errorList := make([]error, 0)
	result := evaluationResult
	for _, group := range result.Groups() {
		for _, policy := range result.Valid[group] {
			policyList = append(policyList, MapPolicyToJson(policy, true))
		}
		for _, policy := range result.Violated[group] {
			policyList = append(policyList, MapPolicyToJson(policy, false))
		}
	}
	for _, policy := range result.Errored {
		errorList = append(errorList, policy.ProcessingErrors...)
	}
	return ClusterValidationResult{
		ClusterPath:       evaluationResult.ClusterName,
		ValidationResults: policyList,
		ProcessingErrors:  errorList,
	}
}

func MapPolicyToJson(policy *policy.Policy, isValid bool) PolicyValidationResult {

	violationsList := make([]Violation, len(policy.Violations))
	for v := range policy.Violations {
		violationsList[v] = MapViolationToJson(policy.Violations[v])
	}

	return PolicyValidationResult{
		PolicyGroup:       policy.Group,
		PolicyTitle:       policy.Title,
		PolicyDescription: policy.Description,
		IsValid:           isValid,
		Violations:        violationsList,
	}
}

func MapViolationToJson(violation string) Violation {

	return Violation{
		ErrorMessage: violation,
	}
}
