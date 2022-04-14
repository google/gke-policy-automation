package outputs

import (
	"github.com/google/gke-policy-automation/internal/policy"
)

func MapClusterToJson(evaluationResult *policy.PolicyEvaluationResult) ClusterValidationResult {

	policyList := make([]PolicyValidationResult, evaluationResult.ValidCount()+evaluationResult.ViolatedCount())
	result := evaluationResult
	for _, group := range result.Groups() {
		for _, policy := range result.Valid[group] {
			policyList = append(policyList, MapPolicyToJson(policy, true))
		}
		for _, policy := range result.Violated[group] {
			policyList = append(policyList, MapPolicyToJson(policy, false))
		}
	}
	return ClusterValidationResult{
		ClusterPath:       evaluationResult.ClusterName,
		ValidationResults: policyList,
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
