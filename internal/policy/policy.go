// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
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
	"strings"

	"github.com/google/gke-policy-automation/internal/log"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
)

const regoPolicyPackage = "gke.policy"
const regoQuery = "data." + regoPolicyPackage + "[name]"
const regoTestFileSuffix = "_test.rego"

type PolicyAgent struct {
	ctx      context.Context
	compiler *ast.Compiler
	compiled map[string]*Policy
}

type Policy struct {
	Name             string
	File             string
	Title            string
	Description      string
	Group            string
	Valid            bool
	Violations       []string
	ProcessingErrors []error
}

type PolicyEvaluationResult struct {
	ClusterName string
	Valid       map[string][]*Policy
	Violated    map[string][]*Policy
	Errored     []*Policy
}

type RegoEvaluationResult struct {
	Name       string
	Valid      bool
	Violations []string
}

func NewPolicyAgent(ctx context.Context) *PolicyAgent {
	return &PolicyAgent{
		ctx:      ctx,
		compiled: make(map[string]*Policy),
	}
}

func NewPolicyEvaluationResult() *PolicyEvaluationResult {
	return &PolicyEvaluationResult{
		Valid:    make(map[string][]*Policy),
		Violated: make(map[string][]*Policy),
		Errored:  make([]*Policy, 0),
	}
}

func (r *PolicyEvaluationResult) Groups() []string {
	groupMap := make(map[string]bool)
	for k := range r.Valid {
		groupMap[k] = true
	}
	for k := range r.Violated {
		groupMap[k] = true
	}
	groups := make([]string, len(groupMap))
	i := 0
	for k := range groupMap {
		groups[i] = k
		i++
	}
	return groups
}

func (r *PolicyEvaluationResult) AddPolicy(policy *Policy) {
	if len(policy.ProcessingErrors) > 0 {
		r.Errored = append(r.Errored, policy)
		return
	}
	if policy.Valid {
		r.Valid[policy.Group] = append(r.Valid[policy.Group], policy)
	} else {
		r.Violated[policy.Group] = append(r.Violated[policy.Group], policy)
	}
}

func (r *PolicyEvaluationResult) ValidCount() int {
	cnt := 0
	for _, v := range r.Valid {
		cnt += len(v)
	}
	return cnt
}

func (r *PolicyEvaluationResult) ViolatedCount() int {
	cnt := 0
	for _, v := range r.Violated {
		cnt += len(v)
	}
	return cnt
}

func (r *PolicyEvaluationResult) ErroredCount() int {
	return len(r.Errored)
}

func (pa *PolicyAgent) Compile(files []*PolicyFile) error {
	modules := make(map[string]string)
	for _, file := range files {
		modules[file.FullName] = file.Content
	}
	compiler, err := ast.CompileModulesWithOpt(modules,
		ast.CompileOpts{ParserOptions: ast.ParserOptions{ProcessAnnotation: true}})
	if err != nil {
		return err
	}
	pa.compiler = compiler
	return nil
}

func (pa *PolicyAgent) ParseCompiled() ([]*Policy, []error) {
	if pa.compiler == nil {
		return nil, []error{fmt.Errorf("compiler is nil")}
	}
	policies := make([]*Policy, 0)
	errors := make([]error, 0)
	for _, m := range pa.compiler.Modules {
		policy := Policy{}
		policy.MapModule(m)
		if !strings.HasPrefix(policy.Name, regoPolicyPackage) || strings.HasSuffix(policy.File, regoTestFileSuffix) {
			continue
		}
		metaErrs := policy.MetadataErrors()
		if len(metaErrs) > 0 {
			errors = append(errors, fmt.Errorf("policy %s has metadata errors: %s", policy.Name, strings.Join(metaErrs, ", ")))
		} else {
			policies = append(policies, &policy)
		}
	}
	return policies, errors
}

func (pa *PolicyAgent) WithFiles(files []*PolicyFile) error {
	if err := pa.Compile(files); err != nil {
		return err
	}
	policies, errors := pa.ParseCompiled()
	if len(errors) > 0 {
		return errors[0]
	}
	pa.compiled = make(map[string]*Policy)
	for _, policy := range policies {
		pa.compiled[policy.Name] = policy
	}
	return nil
}

func (pa *PolicyAgent) Evaluate(input interface{}) (*PolicyEvaluationResult, error) {
	var rgo *rego.Rego
	if pa.compiler == nil {
		rgo = rego.New(
			rego.Input(input),
			rego.Query(regoQuery))
	} else {
		rgo = rego.New(
			rego.Compiler(pa.compiler),
			rego.Input(input),
			rego.Query(regoQuery))
	}
	results, err := rgo.Eval(pa.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate rego: %s", err)
	}
	return pa.processRegoResultSet(results)
}

func (pa *PolicyAgent) processRegoResultSet(results rego.ResultSet) (*PolicyEvaluationResult, error) {
	evalResults := NewPolicyEvaluationResult()
	for _, result := range results {
		value, bindings, err := getResultDataForEval(result)
		if err != nil {
			evalResults.AddPolicy(NewPolicyFromEvalResult(&RegoEvaluationResult{}, []error{err}))
			continue
		}
		regoEvalResult := RegoEvaluationResult{}
		regoEvalResultErrors := make([]error, 0)
		if err := regoEvalResult.mapExpressionBindings(bindings); err != nil {
			regoEvalResultErrors = append(regoEvalResultErrors, err)
		}
		if err := regoEvalResult.mapExpressionValue(value); err != nil {
			regoEvalResultErrors = append(regoEvalResultErrors, err)
		}
		policy := NewPolicyFromEvalResult(&regoEvalResult, regoEvalResultErrors)
		policyName := regoPolicyPackage + "." + regoEvalResult.Name
		if compiledPolicy, ok := pa.compiled[policyName]; ok {
			compiledPolicy.Valid = policy.Valid
			compiledPolicy.Violations = policy.Violations
			compiledPolicy.ProcessingErrors = policy.ProcessingErrors
			policy = compiledPolicy
		} else {
			log.Warnf("rego policy %q has no match with any compiled policy", policyName)
		}
		evalResults.AddPolicy(policy)
	}
	return evalResults, nil
}

func getResultDataForEval(regoResult rego.Result) (value interface{}, bindings map[string]interface{}, err error) {
	if len(regoResult.Expressions) < 1 {
		err = fmt.Errorf("result has no expressions")
		return
	}
	if len(regoResult.Bindings) < 1 {
		err = fmt.Errorf("result has no bindings")
		return
	}
	value = regoResult.Expressions[0].Value
	bindings = regoResult.Bindings
	return
}

func (r *RegoEvaluationResult) mapExpressionBindings(bindings map[string]interface{}) error {
	name, ok := bindings["name"]
	if !ok {
		return fmt.Errorf("expression has no binding for key %q", "name")
	}
	nameStr, ok := name.(string)
	if !ok {
		return fmt.Errorf("expression binding for key %q is %q (expected string) ", "name", reflect.TypeOf(name))
	}
	r.Name = nameStr
	return nil
}

func (r *RegoEvaluationResult) mapExpressionValue(value interface{}) error {
	valueMap, ok := value.(map[string]interface{})
	if !ok {
		return fmt.Errorf("rego expression value type is %q (expected map[string]interface{})", reflect.TypeOf(value))
	}
	valid, violations, err := parseRegoPolicyData(valueMap)
	if err != nil {
		return err
	}
	r.Valid = valid
	r.Violations = violations
	return nil
}

func parseRegoPolicyData(data interface{}) (valid bool, violations []string, err error) {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		err = fmt.Errorf("failed to convert value of type %q to map[string]interface{}", reflect.TypeOf(data))
		return
	}
	if valid, err = getBoolFromInterfaceMap("valid", dataMap); err != nil {
		return
	}
	if violations, err = getStringListFromInterfaceMap("violation", dataMap); err != nil {
		return
	}
	return
}

func NewPolicyFromEvalResult(result *RegoEvaluationResult, errors []error) *Policy {
	policy := &Policy{
		Name:       result.Name,
		Valid:      result.Valid,
		Violations: result.Violations,
	}
	if len(errors) > 0 {
		policy.ProcessingErrors = errors
	}
	return policy
}

func (p *Policy) MapModule(module *ast.Module) {
	p.Name = module.Package.String()[8:]
	p.File = module.Package.Location.File
	for _, annot := range module.Annotations {
		if annot.Scope != "package" {
			continue
		}
		p.Title = annot.Title
		p.Description = annot.Description
		if group, ok := annot.Custom["group"]; ok {
			if groupS, okS := group.(string); okS {
				p.Group = groupS
			}
		}
	}
}

func (p Policy) MetadataErrors() []string {
	errs := make([]string, 0)
	if p.Title == "" {
		errs = append(errs, "title is not set")
	}
	if p.Description == "" {
		errs = append(errs, "description is not set")
	}
	if p.Group == "" {
		errs = append(errs, "group is not set")
	}
	return errs
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
