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

package policy

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	cfg "github.com/google/gke-policy-automation/internal/config"
	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/rego"
)

func TestCompile(t *testing.T) {
	policyFiles := []*PolicyFile{
		{"test_one.rego", "folder/test_one.rego", `
package test_one
p = 1`},
		{"test_two.rego", "folder/test_two.rego", `
package bla.test_two
p = 2`}}
	pa := NewPolicyAgent(context.Background())

	err := pa.Compile(policyFiles)
	if err != nil {
		t.Fatalf("err = %q; want nil", err)
	}
	gkePa, ok := pa.(*GKEPolicyAgent)
	if !ok {
		t.Fatalf("policy agent type is not *GKEPolicyAgent")
	}
	if gkePa.compiler == nil {
		t.Fatalf("compiler = nil; want compiler")
	}
	if len(gkePa.compiler.Modules) != len(policyFiles) {
		t.Errorf("number of compiled policies = %d; want %d", len(gkePa.compiler.Modules), len(policyFiles))
	}
	for _, file := range policyFiles {
		if _, ok := gkePa.compiler.Modules[file.FullName]; !ok {
			t.Errorf("compiler has no module for file %s", file)
		}
	}
}

func TestCompile_parseError(t *testing.T) {
	policyFiles := []*PolicyFile{
		{"test_one.rego", "folder/test_one.rego", `
bla bla`}}
	pa := GKEPolicyAgent{}
	err := pa.Compile(policyFiles)
	if err == nil {
		t.Errorf("err is nil; want error")
	}
}

func TestParseCompiled(t *testing.T) {
	goodPackage := "gke.policy.testOk"
	policyContentOk := fmt.Sprintf("# METADATA\n"+
		"# title: TestTitle\n"+
		"# description: TestDescription\n"+
		"# custom:\n"+
		"#   group: TestGroup\n"+
		"#   severity: High\n"+
		"#   sccCategory: Category\n"+
		"package %s\n"+
		"p = 1", goodPackage)
	policyContentBadMeta := `# METADATA
# title:  TestTitle
package gke.policy.badMeta
p = 1`
	policyContentBadMetaTwo := `# METADATA
# title: TestTitle
# description: TestDescription
package gke.policy.badMetaTwo
p = 1`

	policyFiles := []*PolicyFile{
		{"test_one.rego", "folder/test_one.rego", policyContentOk},
		{"test_two.rego", "folder/test_two.rego", policyContentBadMeta},
		{"test_three.rego", "folder/test_three.rego", policyContentBadMetaTwo},
	}
	pa := GKEPolicyAgent{}
	if err := pa.Compile(policyFiles); err != nil {
		t.Fatalf("err is %s; expected nil", err)
	}
	errors := pa.ParseCompiled()
	if len(pa.policies) != 1 {
		t.Fatalf("len(policies) = %v; want %v", len(pa.policies), 1)
	}
	if len(errors) != 2 {
		t.Fatalf("len(errors) = %v; want %v", len(errors), 2)
	}
	if pa.policies[0].Name != goodPackage {
		t.Errorf("policy[0] name = %v; want %v", pa.policies[0].Name, goodPackage)
	}
}

func TestParseCompiled_noCompiler(t *testing.T) {
	pa := GKEPolicyAgent{}
	if err := pa.ParseCompiled(); err == nil {
		t.Fatalf("err is nil; want error")
	}
}

func TestWithFiles(t *testing.T) {
	ignoredPkg := "gke.invalid"
	packageOne := "gke.policy.package_one"
	titleOne := "TitleOne"
	contentOne := fmt.Sprintf("# METADATA\n"+
		"# title: %s\n"+
		"# description: Test\n"+
		"# custom:\n"+
		"#   group: Test\n"+
		"#   severity: High\n"+
		"#   sccCategory: Category\n"+
		"package %s\n"+
		"p = 1", titleOne, packageOne)
	packageTwo := "gke.scalability.package_two"
	titleTwo := "TitleTwo"
	contentTwo := fmt.Sprintf("# METADATA\n"+
		"# title: %s\n"+
		"# description: Test\n"+
		"# custom:\n"+
		"#   group: Test\n"+
		"#   severity: High\n"+
		"#   sccCategory: Category\n"+
		"package %s\n"+
		"p = 1", titleTwo, packageTwo)
	contentThree := fmt.Sprintf("# METADATA\n"+
		"# title: TitleThree\n"+
		"# description: Test\n"+
		"# custom:\n"+
		"#   group: Test\n"+
		"package %s\n"+
		"p = 1", ignoredPkg+".test")
	policyFiles := []*PolicyFile{
		{"test_one.rego", "folder/test_one.rego", contentOne},
		{"test_two.rego", "folder/test_two.rego", contentTwo},
		{"test_three.rego", "folder/test_three.rego", contentThree},
		{"test_one_test.rego", "folder/test_one_test.rego", contentThree},
	}
	policyExclusions := &cfg.ConfigPolicyExclusions{
		Policies:     []string{"gke.policy.enable_ilb_subsetting"},
		PolicyGroups: []string{"security"},
	}
	pa := GKEPolicyAgent{parserIgnoredPkgs: []string{ignoredPkg}}
	if err := pa.WithFiles(policyFiles, *policyExclusions); err != nil {
		t.Fatalf("error = %v; want nil", err)
	}
	if len(pa.policies) != 2 {
		t.Fatalf("len(pa.compiled) = %v; want %v", len(pa.policies), 2)
	}
}

func TestProcessRegoResultSet(t *testing.T) {
	regoPackageBase := "gke.policy"
	policyOneCompiled := &Policy{
		Name:        regoPackageBase + ".policy_one",
		File:        "rego/policy_one.rego",
		Title:       "Policy One test",
		Description: "This is just for test",
		Group:       "policy_one",
	}
	policyOneResult := rego.Result{
		Expressions: []*rego.ExpressionValue{
			{Value: map[string]interface{}{
				"valid":     true,
				"violation": []interface{}{},
			}},
		},
		Bindings: map[string]interface{}{
			"name": "policy_one",
		},
	}
	policyTwoCompiled := &Policy{
		Name:        regoPackageBase + ".policy_two",
		File:        "rego/policy_two.rego",
		Title:       "Policy Two test",
		Description: "This is just for test",
		Group:       "policy_two",
	}
	policyTwoResult := rego.Result{
		Expressions: []*rego.ExpressionValue{
			{Value: map[string]interface{}{
				"valid":     false,
				"violation": []interface{}{"error"},
			}},
		},
		Bindings: map[string]interface{}{
			"name": "policy_two",
		},
	}
	policyThreeCompiled := &Policy{
		Name:        regoPackageBase + ".policy_three",
		File:        "rego/policy_three.rego",
		Title:       "Policy Three test",
		Description: "This is just for test",
	}
	policyThreeResult := rego.Result{
		Expressions: []*rego.ExpressionValue{
			{Value: map[string]interface{}{
				"valid": false,
			}},
		},
		Bindings: map[string]interface{}{
			"name": "policy_three",
		},
	}
	resultSet := []rego.Result{policyOneResult, policyTwoResult, policyThreeResult}
	pa := GKEPolicyAgent{}
	pa.policies = []*Policy{policyOneCompiled, policyTwoCompiled, policyThreeCompiled}

	result, err := pa.processRegoResultSet(regoPackageBase, resultSet)
	if err != nil {
		t.Fatalf("got error; expected nil")
	}
	if len(result.Policies) != 3 {
		t.Errorf("result policies number = %v; want %v", len(result.Policies), 3)
	}
	if len(pa.evalCache) != len(pa.policies) {
		t.Fatalf("number of policies in eval cache = %v; want %v", len(pa.evalCache), len(pa.policies))
	}
}

func TestInitEvalCache(t *testing.T) {
	pa := &GKEPolicyAgent{}
	pa.policies = []*Policy{
		{
			Name:  "gke.scalability.policy_one",
			Title: "policy one",
		},
		{
			Name:  "gke.scalability.policy_twp",
			Title: "policy two",
		},
		{
			Name:  "gke.scalability.policy_three",
			Title: "policy two",
		},
	}
	pa.initEvalCache()
	if len(pa.evalCache) != len(pa.policies) {
		t.Fatalf("number of policies in eval cache = %v; want %v", len(pa.evalCache), len(pa.policies))
	}
	for _, policy := range pa.policies {
		evalPolicy, ok := pa.evalCache[policy.Name]
		if !ok {
			t.Fatalf("policy with name %v missing in eval cache", policy.Name)
		}
		if evalPolicy.Name != policy.Name {
			t.Errorf("evalPolicy name = %v; want %v", evalPolicy.Name, policy.Name)
		}
		if evalPolicy.Title != policy.Title {
			t.Errorf("evalPolicy title = %v; want %v", evalPolicy.Title, policy.Title)
		}
	}
}

func TestGetResultDataForEval(t *testing.T) {
	input := []rego.Result{
		{Expressions: []*rego.ExpressionValue{{Value: "test"}},
			Bindings: map[string]interface{}{"name": "test"}},
		{Expressions: []*rego.ExpressionValue{{Text: "test"}}},
		{Bindings: map[string]interface{}{"name": "test"}},
		{},
	}
	expected := []interface{}{
		rego.Result{
			Expressions: []*rego.ExpressionValue{{Value: "test"}},
			Bindings:    map[string]interface{}{"name": "test"}},
		nil,
		nil,
		nil,
	}
	for i := range input {
		value, bindings, err := getResultDataForEval(input[i])
		if err == nil {
			expectedResult := expected[i].(rego.Result)
			if !reflect.DeepEqual(value, expectedResult.Expressions[0].Value) {
				t.Errorf("value = %v; want %v", value, expectedResult.Expressions[0].Value)
			}
			if !reflect.DeepEqual(bindings["name"], expectedResult.Bindings["name"]) {
				t.Errorf("bindings[name] = %v; want %v", bindings["name"], expectedResult.Bindings["name"])
			}
		} else {
			if expected[i] != nil {
				t.Errorf("did not expect error; got error")
			}
		}
	}
}

func TestMapExpressionBindings(t *testing.T) {
	bindings := []map[string]interface{}{
		{"name": "policy_name"},
		{"name": 20},
		{"bogus": "value"},
	}
	expected := []interface{}{
		"policy_name",
		nil,
		nil,
	}
	result := RegoEvaluationResult{}
	for i := range bindings {
		err := result.mapExpressionBindings(bindings[i])
		if err == nil {
			if result.Name != expected[i] {
				t.Errorf("name = %v; want %v", result.Name, expected[i])
			}
		} else {
			if expected[i] != nil {
				t.Errorf("did not expect error; got error")
			}
		}
	}
}

func TestMapExpressionValue(t *testing.T) {
	input := map[string]interface{}{
		"valid":     true,
		"violation": []interface{}{"violation"},
	}
	expectedValid := true
	expectedViolations := []string{"violation"}

	result := RegoEvaluationResult{}
	if err := result.mapExpressionValue(input); err != nil {
		t.Errorf("err = %q; want nil", err)
	}
	if result.Valid != expectedValid {
		t.Errorf("valid = %v; want %v", result.Valid, expectedValid)
	}
	if !reflect.DeepEqual(result.Violations, expectedViolations) {
		t.Errorf("valid = %v; want %v", result.Violations, expectedViolations)
	}
}

func TestParseRegoPolicyData(t *testing.T) {
	input := map[string]interface{}{
		"valid":     true,
		"violation": []interface{}{"violation"},
	}
	expectedValid := true
	expectedViolations := []string{"violation"}

	valid, violations, err := parseRegoPolicyData(input)
	if err != nil {
		t.Errorf("err = %q; want nil", err)
	}
	if valid != input["valid"] {
		t.Errorf("valid = %v; want %v", valid, expectedValid)
	}
	if !reflect.DeepEqual(violations, expectedViolations) {
		t.Errorf("violations = %v; want %v", violations, expectedViolations)
	}
}

func TestMapModule(t *testing.T) {
	file := "folder/test_one.rego"
	pkg := "gke.policy.test"
	title := "This is title"
	desc := "This is long description"
	group := "TestGroup"
	severity := "Low"
	category := "TEST"
	cisVersion := "1.2"
	cisID := "4.1.3"
	recommendation := "do this and that"
	externalURI := "https://cloud.google.com/kubernetes-engine"

	content := fmt.Sprintf("# METADATA\n"+
		"# title: %s\n"+
		"# description: %s\n"+
		"# custom:\n"+
		"#   group: %s\n"+
		"#   severity: %s\n"+
		"#   sccCategory: %s\n"+
		"#   cis:\n"+
		"#     version: %q\n"+
		"#     id: %q\n"+
		"#   recommendation: %s\n"+
		"#   externalURI: %s\n"+
		"package %s\n"+
		"p = 1", title, desc, group, severity, category, cisVersion, cisID, recommendation, externalURI, pkg)

	modules := map[string]string{file: content}
	compiler := ast.MustCompileModulesWithOpts(modules,
		ast.CompileOpts{ParserOptions: ast.ParserOptions{ProcessAnnotation: true}})
	module := compiler.Modules[file]
	policy := Policy{}
	policy.mapModule(module)

	if policy.Name != pkg {
		t.Errorf("name = %v; want %v", policy.Name, pkg)
	}
	if policy.File != file {
		t.Errorf("file = %v; want %v", policy.File, file)
	}
	if policy.Title != title {
		t.Errorf("title = %v; want %v", policy.Title, title)
	}
	if policy.Description != desc {
		t.Errorf("description = %v; want %v", policy.Description, desc)
	}
	if policy.Group != group {
		t.Errorf("group = %v; want %v", policy.Group, group)
	}
	if policy.Severity != severity {
		t.Errorf("severity = %v; want %v", policy.Severity, severity)
	}
	if policy.Category != category {
		t.Errorf("category = %v; want %v", policy.Category, category)
	}
	if policy.CisVersion != cisVersion {
		t.Errorf("cis version = %v; want %v", policy.CisVersion, cisVersion)
	}
	if policy.CisID != cisID {
		t.Errorf("cis id = %v; want %v", policy.CisID, cisID)
	}
	if policy.ExternalURI != externalURI {
		t.Errorf("externalURI = %v; want %v", policy.ExternalURI, externalURI)
	}
	if policy.Recommendation != recommendation {
		t.Errorf("recommendation = %v; want %v", policy.Recommendation, recommendation)
	}
}

func TestMetadataErrors(t *testing.T) {
	input := []Policy{
		{Title: "title", Description: "description", Group: "group", Severity: "High", Category: "TEST"},
		{Title: "title", Description: "description", Group: "group", Severity: "High"},
		{Title: "title", Description: "description", Group: "group"},
		{Title: "title"},
		{},
		{Title: "title", Description: "description", Group: "group", Severity: "High", Category: "TEST", CisVersion: "1.0"},
		{Title: "title", Description: "description", Group: "group", Severity: "High", Category: "TEST", CisID: "1.1.1"},
	}
	expErrCnt := []int{
		0,
		1,
		2,
		4,
		5,
		1,
		1,
	}
	for i := range input {
		errors := input[i].MetadataErrors()
		if len(errors) != expErrCnt[i] {
			t.Errorf("error cnt = %v; want %v", len(errors), expErrCnt[i])
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

func TestInitPolicyExcludeCache(t *testing.T) {
	pa := &GKEPolicyAgent{}
	pa.excludes.Policies = []string{"policy_one", "policy_two"}
	policyExcludeCache := pa.initPolicyExcludeCache()
	if len(policyExcludeCache) != len(pa.excludes.Policies) {
		t.Fatalf("number of policies in exclude cache = %v; want %v", len(policyExcludeCache), len(pa.excludes.Policies))
	}
	for _, policy := range pa.excludes.Policies {
		_, ok := policyExcludeCache["data."+policy]
		if !ok {
			t.Fatalf("policy with name %v missing in exclude cache", policy)
		}
	}
}

func TestInitGroupExcludeCache(t *testing.T) {
	pa := &GKEPolicyAgent{}
	pa.excludes.PolicyGroups = []string{"group_one", "group_two"}
	groupExcludeCache := pa.initGroupExcludeCache()
	if len(groupExcludeCache) != len(pa.excludes.PolicyGroups) {
		t.Fatalf("number of policy groups in exclude cache = %v; want %v", len(groupExcludeCache), len(pa.excludes.PolicyGroups))
	}
	for _, group := range pa.excludes.PolicyGroups {
		_, ok := groupExcludeCache[group]
		if !ok {
			t.Fatalf("group with name %v missing in exclude cache", group)
		}
	}
}

func TestIsExcluded(t *testing.T) {
	inputName := "test"
	inputMap := map[string]bool{"test": true}
	expected := true

	result, err := isExcluded(inputName, inputMap)
	if err != nil {
		t.Errorf("err = %q; want nil", err)
	}
	if result != expected {
		t.Errorf("result = %v; want %v", result, expected)
	}
}

func TestIsExcluded_negative(t *testing.T) {
	inputName := "missing"
	inputMap := map[string]bool{"test": true}
	expected := false

	result, err := isExcluded(inputName, inputMap)
	if err == nil {
		t.Errorf("err = nil; want error")
	}
	if result != expected {
		t.Errorf("result = %v; want %v", result, expected)
	}
}

func TestGetRegoQueryForPackageBase(t *testing.T) {
	base := "gke.scalability"
	query := getRegoQueryForPackageBase(base)
	expected := "data." + base + "[name]"
	if query != expected {
		t.Fatalf("query = %v; want %v", query, expected)
	}
}

func TestGetStringFromInterfaceMap(t *testing.T) {
	m := map[string]interface{}{
		"keyOne": "value",
		"keyTwo": 12,
	}
	if v, ok := getStringFromInterfaceMap("keyOne", m); !ok {
		t.Errorf("ok for keyOne is false; want true")

	} else if v != m["keyOne"] {
		t.Errorf("value for keyOne = %v; want %v", v, m["keyOne"])
	}
	if _, ok := getStringFromInterfaceMap("keyTwo", m); ok {
		t.Errorf("ok for keyOne is true; want false")
	}
	if _, ok := getStringFromInterfaceMap("missing", m); ok {
		t.Errorf("ok for missing is true; want false")
	}
}
