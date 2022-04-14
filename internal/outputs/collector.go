package outputs

import (
	"github.com/google/gke-policy-automation/internal/policy"
)

type ValidationResultCollector interface {
	RegisterResult(results []*policy.PolicyEvaluationResult) error
	Close() error
}
