package policy

import (
	"errors"
	"reflect"
	"testing"

	"github.com/open-policy-agent/opa/rego"
)

func TestPolicyEvaluationResultCounts(t *testing.T) {
	input := PolicyEvaluationResult{
		errored:       make([]*Policy, 8),
		validCount:    10,
		violatedCount: 5,
	}
	if input.ValidCount() != input.validCount {
		t.Errorf("ValidCount is %d; want %d", input.ValidCount(), input.validCount)
	}
	if input.ViolatedCount() != input.violatedCount {
		t.Errorf("ViolatedCount is %d; want %d", input.ViolatedCount(), input.violatedCount)

	}
	if input.ErroredCount() != len(input.errored) {
		t.Errorf("ErroredCount is %d; want %d", input.ErroredCount(), len(input.errored))
	}
}

func TestProcessRegoResult(t *testing.T) {
	inputData := []*rego.Result{
		{Expressions: []*rego.ExpressionValue{
			{Value: []interface{}{
				map[string]interface{}{
					"name": "test_policy",
					"data": map[string]interface{}{
						"name":        "Test Name",
						"description": "Test Description",
						"group":       "Test Group",
						"valid":       true,
						"violation":   []interface{}{"violation"},
					},
				},
			}},
		}},
		{Expressions: []*rego.ExpressionValue{
			{Value: []interface{}{
				map[string]interface{}{},
			},
			},
		}},
		{Expressions: []*rego.ExpressionValue{
			{Value: []interface{}{
				map[string]interface{}{
					"name": "test_policy",
				},
			},
			}},
		},
	}
	expectedResults := []*PolicyEvaluationResult{
		{successful: map[string][]*Policy{"Test Group": make([]*Policy, 1)}},
		{errored: make([]*Policy, 1)},
		{errored: make([]*Policy, 1)},
	}
	for i := range inputData {
		result, err := processRegoResult(inputData[i])
		if err != nil {
			t.Errorf("err is not nil; want nil")
		}
		if len(result.successful) != len(expectedResults[i].successful) {
			t.Errorf("len(successful) = %d; want %d", len(result.successful), len(expectedResults[i].successful))
		}
		if len(result.errored) != len(expectedResults[i].errored) {
			t.Errorf("len(errored) = %d; want %d", len(result.errored), len(expectedResults[i].errored))
		}
	}
}

func TestProcessRegoResult_negative(t *testing.T) {
	_, err := processRegoResult(&rego.Result{})
	if err == nil {
		t.Errorf("err is nil; want error")
	}
}

func TestGetExpressionValueList(t *testing.T) {
	input := &rego.Result{Expressions: []*rego.ExpressionValue{{Value: []interface{}{"test"}}}}
	expected := []interface{}{"test"}
	result, err := getExpressionValueList(input, 0)
	if err != nil {
		t.Errorf("err is not nil; want nil")
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("result = %v; want %v", result, expected)
	}
}

func TestGetExpressionValueList_negative(t *testing.T) {
	inputs := []*rego.Result{
		{Expressions: []*rego.ExpressionValue{{}}},
		{},
	}
	for _, input := range inputs {
		_, err := getExpressionValueList(input, 0)
		if err == nil {
			t.Errorf("err is nil; want error")
		}
	}
}

func TestParseRegoExpressionValue(t *testing.T) {
	inputData := []map[string]interface{}{
		{
			"name": "test_policy",
			"data": map[string]interface{}{
				"name":        "Test Name",
				"description": "Test Description",
				"group":       "Test Group",
				"valid":       true,
				"violation":   []interface{}{"violation"},
			},
		},
		{"name": "test_policy"},
		{"name": "test_policy", "data": nil},
	}
	expectedResults := []*Policy{
		{Name: "test_policy"},
		{Name: "test_policy", ProcessingErrors: []error{errors.New("")}},
		{Name: "test_policy", ProcessingErrors: []error{errors.New("")}},
	}
	for i := range inputData {
		policy, err := parseRegoExpressionValue(inputData[i])
		if err != nil {
			t.Errorf("err is not nil; want nil")
		}
		if policy.Name != expectedResults[i].Name {
			t.Errorf("name = %s; want %s", policy.Name, expectedResults[i].Name)
		}
		if len(policy.ProcessingErrors) != len(expectedResults[i].ProcessingErrors) {
			t.Errorf("len(ProcessingErrors) = %d; want %d", len(policy.ProcessingErrors), len(expectedResults[i].ProcessingErrors))
		}
	}
}

func TestParseRegoExpressionValue_negative(t *testing.T) {
	inputData := []map[string]interface{}{
		nil,
		{"data": nil},
	}
	for _, input := range inputData {
		_, err := parseRegoExpressionValue(input)
		if err == nil {
			t.Errorf("err is nil; want error")
		}
	}
}

func TestMapRegoPolicyData(t *testing.T) {
	input := map[string]interface{}{
		"name":        "Test Name",
		"description": "Test Description",
		"group":       "Test Group",
		"valid":       true,
		"violation":   []interface{}{"violation"},
	}
	expected := &Policy{
		FullName:    "Test Name",
		Description: "Test Description",
		Group:       "Test Group",
		Valid:       true,
		Violations:  []string{"violation"},
	}

	policy := Policy{}
	err := policy.mapRegoPolicyData(input)
	if err != nil {
		t.Errorf("err = %q; want nil", err)
	}
	if policy.FullName != expected.FullName {
		t.Errorf("name = %s; want %s", policy.Name, expected.FullName)
	}
	if policy.Description != expected.Description {
		t.Errorf("description = %s; want %s", policy.Description, expected.Description)
	}
	if policy.Group != expected.Group {
		t.Errorf("group = %s; want %s", policy.Group, expected.Group)
	}
	if policy.Valid != expected.Valid {
		t.Errorf("valid = %v; want %v", policy.Valid, expected.Valid)
	}
	if !reflect.DeepEqual(policy.Violations, expected.Violations) {
		t.Errorf("violations = %v; want %v", policy.Violations, expected.Violations)
	}
}

func TestMapRegoPolicyData_negative(t *testing.T) {
	inputData := []interface{}{
		nil,
		map[string]interface{}{},
	}
	for _, input := range inputData {
		policy := Policy{}
		err := policy.mapRegoPolicyData(input)
		if err == nil {
			t.Errorf("err is nil; want error")
		}
	}
}

func TestGetStringFromInterfaceMap(t *testing.T) {
	inputName := "test"
	inputMap := map[string]interface{}{"test": "value"}
	expected := "value"

	result, err := getStringFromInterfaceMap(inputName, inputMap)
	if err != nil {
		t.Errorf("err = %q; want nil", err)
	}
	if result != expected {
		t.Errorf("result = %q; want %q", result, expected)
	}
}

func TestGetStringFromInterfaceMap_negative(t *testing.T) {
	inputNames := []string{"testTwo", "missing"}
	inputMaps := []map[string]interface{}{{"testTwo": 101}, nil}
	for i := range inputNames {
		_, err := getStringFromInterfaceMap(inputNames[i], inputMaps[i])
		if err == nil {
			t.Errorf("err = nil; want error")
		}
	}
}

func TestGetBoolFromInterfaceMap(t *testing.T) {
	inputName := "test"
	inputMap := map[string]interface{}{"test": true}
	expected := true

	result, err := getBoolFromInterfaceMap(inputName, inputMap)
	if err != nil {
		t.Errorf("err = %q; want nil", err)
	}
	if result != expected {
		t.Errorf("result = %v; want %v", result, expected)
	}
}

func TestGetBoolFromInterfaceMap_negative(t *testing.T) {
	inputNames := []string{"testTwo", "missing"}
	inputMaps := []map[string]interface{}{{"testTwo": 101}, nil}
	for i := range inputNames {
		_, err := getBoolFromInterfaceMap(inputNames[i], inputMaps[i])
		if err == nil {
			t.Errorf("err = nil; want error")
		}
	}
}

func TestGetStringListFromInterfaceMap(t *testing.T) {
	inputName := "test"
	inputMap := map[string]interface{}{"test": []interface{}{"str1", "str2"}}
	expected := []string{"str1", "str2"}

	result, err := getStringListFromInterfaceMap(inputName, inputMap)
	if err != nil {
		t.Errorf("err = %q; want nil", err)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("result = %v; want %v", result, expected)
	}
}

func TestGetStringListFromInterfaceMap_negative(t *testing.T) {
	inputNames := []string{"testTwo", "testThree", "missing"}
	inputMaps := []map[string]interface{}{
		{"testTwo": nil},
		{"testThree": []interface{}{"str1", 100}},
		nil}
	for i := range inputNames {
		_, err := getStringListFromInterfaceMap(inputNames[i], inputMaps[i])
		if err == nil {
			t.Errorf("err = nil; want error")
		}
	}
}
