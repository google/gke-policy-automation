package policy

import (
	"context"
	"fmt"
	"reflect"

	"github.com/open-policy-agent/opa/rego"
)

type PolicyAgent struct {
	ctx     context.Context
	dataDir string
}

type Policy struct {
	Name             string
	FullName         string
	Description      string
	Group            string
	Valid            bool
	Violations       []string
	ProcessingErrors []error
}

type PolicyEvaluationResult struct {
	SuccessFull  []*Policy
	Failed       []*Policy
	TotalCount   int
	ErroredCount int
}

func NewPolicyAgent(ctx context.Context, dataDir string) *PolicyAgent {
	return &PolicyAgent{
		ctx:     ctx,
		dataDir: dataDir,
	}
}

func (p *PolicyAgent) EvaluatePolicies(input interface{}) (*PolicyEvaluationResult, error) {
	rgo := rego.New(rego.Load([]string{p.dataDir}, nil),
		rego.Input(input),
		rego.Query("data.gke.policies_data"))

	rs, err := rgo.Eval(p.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate rego: %s", err)
	}
	if len(rs) < 1 {
		return nil, fmt.Errorf("rego evaluation returned empty result set")
	}
	return processRegoResult(rs[0])
}

func processRegoResult(regoResult rego.Result) (*PolicyEvaluationResult, error) {
	if len(regoResult.Expressions) < 1 {
		return nil, fmt.Errorf("rego result has empty expression list")
	}
	regoResultExpressionValue := regoResult.Expressions[0].Value
	regoResultExpressionValueList, ok := regoResultExpressionValue.([]interface{})
	if !ok {
		return nil, fmt.Errorf("rego expression value type is %q (expected []interface{})", reflect.TypeOf(regoResultExpressionValue))
	}
	results := &PolicyEvaluationResult{}
	results.TotalCount = len(regoResultExpressionValueList)
	for _, result := range regoResultExpressionValueList {
		policy, err := parseRegoExpressionValue(result)
		if err != nil {
			//TODO add warn logging or something
			results.ErroredCount++
			continue
		}
		if len(policy.ProcessingErrors) > 0 {
			results.Failed = append(results.Failed, policy)
		} else {
			results.SuccessFull = append(results.SuccessFull, policy)
		}
	}
	return results, nil
}

func parseRegoExpressionValue(value interface{}) (*Policy, error) {
	valueMap, ok := value.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("rego expression value type is %q (expected map[string]interface{})", reflect.TypeOf(value))
	}
	policy := &Policy{}
	if v, err := getStringFromInterfaceMap("name", valueMap); err == nil {
		policy.Name = v
	} else {
		return nil, fmt.Errorf("policy map does not contain key: %q", "name")
	}
	policyData, ok := valueMap["data"]
	if !ok {
		policy.ProcessingErrors = []error{fmt.Errorf("policy map does not contain key: %q", "data")}
		return policy, nil
	}
	if err := policy.mapRegoPolicyData(policyData); err != nil {
		policy.ProcessingErrors = []error{err}
	}
	return policy, nil
}

func (p *Policy) mapRegoPolicyData(data interface{}) error {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("failed to convert value of type %q to map[string]interface{}", reflect.TypeOf(data))
	}
	if v, err := getStringFromInterfaceMap("name", dataMap); err == nil {
		p.FullName = v
	} else {
		return err
	}
	if v, err := getStringFromInterfaceMap("description", dataMap); err == nil {
		p.Description = v
	} else {
		return err
	}
	if v, err := getStringFromInterfaceMap("group", dataMap); err == nil {
		p.Group = v
	} else {
		return err
	}
	if v, err := getBoolFromInterfaceMap("valid", dataMap); err == nil {
		p.Valid = v
	} else {
		return err
	}
	if v, err := getStringListFromInterfaceMap("violation", dataMap); err == nil {
		p.Violations = v
	} else {
		return err
	}
	return nil
}

func getStringFromInterfaceMap(name string, m map[string]interface{}) (string, error) {
	v, ok := m[name]
	if !ok {
		return "", fmt.Errorf("map does not contain key: %q", name)
	}
	vString, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("key %q type is %q (not a string)", name, reflect.ValueOf(v))
	}
	return vString, nil
}

func getBoolFromInterfaceMap(name string, m map[string]interface{}) (bool, error) {
	v, ok := m[name]
	if !ok {
		return false, fmt.Errorf("map does not contain key: %q", name)
	}
	vBool, ok := v.(bool)
	if !ok {
		return false, fmt.Errorf("key %q type is %q (not a string)", name, reflect.ValueOf(v))
	}
	return vBool, nil
}

func getStringListFromInterfaceMap(name string, m map[string]interface{}) ([]string, error) {
	v, ok := m[name]
	if !ok {
		return nil, fmt.Errorf("map does not contain key: %q", name)
	}
	vList, ok := v.([]interface{})
	if !ok {
		return nil, fmt.Errorf("key %q type is %q (not a []interface{})", name, reflect.ValueOf(v))
	}
	vStringList := make([]string, len(vList))
	for i := range vList {
		vStringListItem, ok := vList[i].(string)
		if !ok {
			return nil, fmt.Errorf("key's %q list element %d is not a string", name, i)
		}
		vStringList[i] = vStringListItem
	}
	return vStringList, nil
}
