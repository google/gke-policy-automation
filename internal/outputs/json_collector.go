package outputs

import (
	"encoding/json"
	"os"

	"github.com/google/gke-policy-automation/internal/policy"
)

type JSONResultCollector struct {
	filename          string
	validationResults ValidationResults
}

func NewJSONResultCollector(filename string) ValidationResultCollector {
	return &JSONResultCollector{
		filename: filename,
	}
}

func (p *JSONResultCollector) RegisterResult(results []*policy.PolicyEvaluationResult) error {

	for _, r := range results {
		p.validationResults.ClusterValidationResults = append(p.validationResults.ClusterValidationResults, MapClusterToJson(r))
	}
	return nil
}

func (p *JSONResultCollector) Close() error {

	res, err := json.Marshal(p.validationResults.ClusterValidationResults)
	if err != nil {
		return err
	}

	d1 := []byte(res)
	os.WriteFile(p.filename, d1, 0644)

	return err
}
