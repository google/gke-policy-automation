package policy

import (
	"errors"
	"reflect"
	"testing"

	"github.com/open-policy-agent/opa/rego"
)

func TestGetStringFromInterfaceMap(t *testing.T) {
	inputNames := []string{"test", "testTwo", "missing"}
	inputMaps := []map[string]interface{}{{"test": "value"}, {"testTwo": 101}, nil}
	expectedResults := []interface{}{"value", errors.New(""), errors.New("")}

	for i := range inputNames {
		result, err := getStringFromInterfaceMap(inputNames[i], inputMaps[i])
		if expected, ok := expectedResults[i].(string); ok {
			if err != nil {
				t.Errorf("err = %q; want nil", err)
			}
			if result != expected {
				t.Errorf("result = %q; want %q", result, expected)
			}
		}
		if _, ok := expectedResults[i].(error); ok {
			if err == nil {
				t.Errorf("err = nil; want error")
			}
		}
	}
}

func TestGetBoolFromInterfaceMap(t *testing.T) {
	inputNames := []string{"test", "testTwo", "missing"}
	inputMaps := []map[string]interface{}{{"test": true}, {"testTwo": 101}, nil}
	expectedResults := []interface{}{true, errors.New("error"), errors.New("error")}

	for i := range inputNames {
		result, err := getBoolFromInterfaceMap(inputNames[i], inputMaps[i])
		if expected, ok := expectedResults[i].(bool); ok {
			if err != nil {
				t.Errorf("err = %q; want nil", err)
			}
			if result != expected {
				t.Errorf("result = %v; want %v", result, expected)
			}
		}
		if _, ok := expectedResults[i].(error); ok {
			if err == nil {
				t.Errorf("err = nil; want error")
			}
		}
	}
}

func TestGetStringListFromInterfaceMap(t *testing.T) {
	inputNames := []string{"test", "testTwo", "testThree", "missing"}
	inputMaps := []map[string]interface{}{
		{"test": []interface{}{"str1", "str2"}},
		{"testTwo": nil},
		{"testThree": []interface{}{"str1", 100}},
		nil}
	expectedResults := []interface{}{
		[]interface{}{"str1", "str2"},
		errors.New(""),
		errors.New(""),
		errors.New(""),
	}
	for i := range inputNames {
		result, err := getStringListFromInterfaceMap(inputNames[i], inputMaps[i])
		if expected, ok := expectedResults[i].([]string); ok {
			if err != nil {
				t.Errorf("err = %q; want nil", err)
			}
			if !reflect.DeepEqual(result, expected) {
				t.Errorf("result = %v; want %v", result, expected)
			}
		}
		if _, ok := expectedResults[i].(error); ok {
			if err == nil {
				t.Errorf("err = nil; want error")
			}
		}
	}
}

func TestMapRegoPolicyData(t *testing.T) {
	inputData := []interface{}{
		map[string]interface{}{
			"name":        "Test Name",
			"description": "Test Description",
			"group":       "Test Group",
			"valid":       true,
			"violation":   []interface{}{"violation"},
		},
		nil,
		map[string]interface{}{},
	}
	expectedResults := []interface{}{
		&Policy{
			FullName:    "Test Name",
			Description: "Test Description",
			Group:       "Test Group",
			Valid:       true,
			Violations:  []string{"violation"},
		},
		errors.New(""),
		errors.New(""),
	}
	for i := range inputData {
		policy := Policy{}
		err := policy.mapRegoPolicyData(inputData[i])
		if expectedPolicy, ok := expectedResults[i].(*Policy); ok {
			if policy.FullName != expectedPolicy.FullName {
				t.Errorf("name = %s; want %s", policy.Name, expectedPolicy.FullName)
			}
			if policy.Description != expectedPolicy.Description {
				t.Errorf("description = %s; want %s", policy.Description, expectedPolicy.Description)
			}
			if policy.Group != expectedPolicy.Group {
				t.Errorf("group = %s; want %s", policy.Group, expectedPolicy.Group)
			}
			if policy.Valid != expectedPolicy.Valid {
				t.Errorf("valid = %v; want %v", policy.Valid, expectedPolicy.Valid)
			}
			if !reflect.DeepEqual(policy.Violations, expectedPolicy.Violations) {
				t.Errorf("violations = %v; want %v", policy.Violations, expectedPolicy.Violations)
			}
		}
		if _, ok := expectedResults[i].(error); ok {
			if err == nil {
				t.Errorf("err is nil; want error")
			}
		}
	}
}

func TestParseRegoExpressionValue(t *testing.T) {
	inputData := []interface{}{
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
		nil,
		map[string]interface{}{
			"data": nil,
		},
		map[string]interface{}{
			"name": "test_policy",
		},
		map[string]interface{}{
			"name": "test_policy",
			"data": nil,
		},
	}
	expectedResults := []interface{}{
		&Policy{
			Name: "test_policy",
		},
		errors.New(""),
		errors.New(""),
		&Policy{
			Name:             "test_policy",
			ProcessingErrors: []error{errors.New("")},
		},
		&Policy{
			Name:             "test_policy",
			ProcessingErrors: []error{errors.New("")},
		},
	}
	for i := range inputData {
		policy, err := parseRegoExpressionValue(inputData[i])
		if expectedPolicy, ok := expectedResults[i].(*Policy); ok {
			if policy.Name != expectedPolicy.Name {
				t.Errorf("name = %s; want %s", policy.Name, expectedPolicy.Name)
			}
			if len(policy.ProcessingErrors) != len(expectedPolicy.ProcessingErrors) {
				t.Errorf("len(ProcessingErrors) = %d; want %d", len(policy.ProcessingErrors), len(expectedPolicy.ProcessingErrors))
			}
		}
		if _, ok := expectedResults[i].(error); ok {
			if err == nil {
				t.Errorf("err is nil; want error")
			}
		}
	}
}

func TestProcessRegoResult(t *testing.T) {
	inputData := []rego.Result{
		{Expressions: []*rego.ExpressionValue{
			{Value: []interface{}{}},
		}},
		{},
		{Expressions: []*rego.ExpressionValue{
			{Value: nil},
		}},
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
	expectedResults := []interface{}{
		&PolicyEvaluationResult{TotalCount: 0, ErroredCount: 0},
		errors.New(""),
		errors.New(""),
		&PolicyEvaluationResult{TotalCount: 1, ErroredCount: 0, SuccessFull: make([]*Policy, 1)},
		&PolicyEvaluationResult{TotalCount: 1, ErroredCount: 1},
		&PolicyEvaluationResult{TotalCount: 1, ErroredCount: 0, Failed: make([]*Policy, 1)},
	}
	for i := range inputData {
		result, err := processRegoResult(inputData[i])
		if expectedResult, ok := expectedResults[i].(*PolicyEvaluationResult); ok {
			if result.TotalCount != expectedResult.TotalCount {
				t.Errorf("totalCount = %d; want %d", result.TotalCount, expectedResult.TotalCount)
			}
			if result.ErroredCount != expectedResult.ErroredCount {
				t.Errorf("erroredCount = %d; want %d", result.ErroredCount, expectedResult.ErroredCount)
			}
			if len(result.SuccessFull) != len(expectedResult.SuccessFull) {
				t.Errorf("len(successFull) = %d; want %d", len(result.SuccessFull), len(expectedResult.SuccessFull))
			}
			if len(result.Failed) != len(expectedResult.Failed) {
				t.Errorf("len(successFull) = %d; want %d", len(result.Failed), len(expectedResult.Failed))
			}
		}
		if _, ok := expectedResults[i].(error); ok {
			if err == nil {
				t.Errorf("err is nil; want error")
			}
		}
	}
}
