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
	"strings"

	cfg "github.com/google/gke-policy-automation/internal/config"
	"github.com/google/gke-policy-automation/internal/log"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
)

const regoTestFileSuffix = "_test.rego"

type PolicyAgent interface {
	Compile(files []*PolicyFile) error
	WithFiles(files []*PolicyFile, excludes cfg.ConfigPolicyExclusions) error
	Evaluate(input interface{}, packageBase string) (*PolicyEvaluationResult, error)
	GetPolicies() []*Policy
}

type GKEPolicyAgent struct {
	ctx               context.Context
	compiler          *ast.Compiler
	policies          []*Policy
	evalCache         map[string]*Policy
	excludes          cfg.ConfigPolicyExclusions
	parserIgnoredPkgs []string
}

type Policy struct {
	Name             string
	File             string
	Title            string
	Description      string
	Group            string
	Severity         string
	Category         string
	Valid            bool
	Violations       []string
	ProcessingErrors []error
	CisVersion       string
	CisID            string
	ExternalURI      string
	Recommendation   string
}

type PolicyEvaluationResult struct {
	ClusterID string
	Policies  []*Policy
}

type RegoEvaluationResult struct {
	Name       string
	Valid      bool
	Violations []string
}

func NewPolicyAgent(ctx context.Context) PolicyAgent {
	return &GKEPolicyAgent{
		ctx:               ctx,
		policies:          make([]*Policy, 0),
		evalCache:         make(map[string]*Policy),
		parserIgnoredPkgs: []string{"gke.rule"},
	}
}

func (pa *GKEPolicyAgent) Compile(files []*PolicyFile) error {
	modules := make(map[string]string)
	for _, file := range files {
		modules[file.FullName] = file.Content
	}
	compiler, err := pa.compileModulesWithOpt(modules,
		ast.CompileOpts{ParserOptions: ast.ParserOptions{ProcessAnnotation: true}})
	if err != nil {
		return err
	}
	pa.compiler = compiler
	return nil
}

func (pa *GKEPolicyAgent) initPolicyExcludeCache() map[string]bool {
	cache := make(map[string]bool)
	for _, policy := range pa.excludes.Policies {
		cache["data."+policy] = true
	}
	return cache
}

func (pa *GKEPolicyAgent) initGroupExcludeCache() map[string]bool {
	cache := make(map[string]bool)
	for _, g := range pa.excludes.PolicyGroups {
		cache[g] = true
	}
	return cache
}

func isExcluded(s string, m map[string]bool) (bool, error) {
	result, ok := m[s]
	if !ok {
		return false, fmt.Errorf("cache does not contain key: %q", s)
	}
	return result, nil
}

func (pa *GKEPolicyAgent) compileModulesWithOpt(modules map[string]string, opts ast.CompileOpts) (*ast.Compiler, error) {

	parsed := make(map[string]*ast.Module, len(modules))

	policyExcludeCache := pa.initPolicyExcludeCache()
	groupExcludeCache := pa.initGroupExcludeCache()

module:
	for f, module := range modules {
		// Filter out tests
		if strings.HasSuffix(f, regoTestFileSuffix) {
			log.Debugf("Skipped policy test file %s", f)
			continue
		}
		var pm *ast.Module
		var err error
		if pm, err = ast.ParseModuleWithOpts(f, module, opts.ParserOptions); err != nil {
			return nil, err
		}

		// Check if the policy is excluded
		if _, err := isExcluded(pm.Package.Path.String(), policyExcludeCache); err == nil {
			continue
		}

		// Check if the group is excluded
		for _, annot := range pm.Annotations {
			if group, ok := annot.Custom["group"]; ok {
				if _, err := isExcluded(fmt.Sprint(group), groupExcludeCache); err == nil {
					continue module
				}
			}
		}
		parsed[f] = pm
	}

	compiler := ast.NewCompiler().WithEnablePrintStatements(opts.EnablePrintStatements)
	compiler.Compile(parsed)

	if compiler.Failed() {
		return nil, compiler.Errors
	}

	return compiler, nil
}

func (pa *GKEPolicyAgent) ParseCompiled() []error {
	if pa.compiler == nil {
		return []error{fmt.Errorf("compiler is nil")}
	}
	errors := make([]error, 0)
module:
	for _, m := range pa.compiler.Modules {
		policy := Policy{}
		policy.mapModule(m)
		if strings.HasSuffix(policy.File, regoTestFileSuffix) {
			continue
		}
		for _, ignored := range pa.parserIgnoredPkgs {
			if strings.HasPrefix(policy.Name, ignored) {
				continue module
			}
		}
		metaErrs := policy.MetadataErrors()
		if len(metaErrs) > 0 {
			errors = append(errors, fmt.Errorf("policy %s has metadata errors: %s", policy.Name, strings.Join(metaErrs, ", ")))
		} else {
			pa.policies = append(pa.policies, &policy)
		}
	}
	return errors
}

func (pa *GKEPolicyAgent) WithFiles(files []*PolicyFile, excludes cfg.ConfigPolicyExclusions) error {
	pa.excludes = excludes
	if err := pa.Compile(files); err != nil {
		return err
	}
	if errors := pa.ParseCompiled(); len(errors) > 0 {
		log.Debugf("parsing compiled policies resulted in %d errors", len(errors))
		for _, err := range errors {
			log.Warnf("parsing compiled policies error: %s", err)
		}
		return errors[0]
	}
	return nil
}

func (pa *GKEPolicyAgent) Evaluate(input interface{}, packageBase string) (*PolicyEvaluationResult, error) {
	query := getRegoQueryForPackageBase(packageBase)
	var rgo *rego.Rego
	if pa.compiler == nil {
		rgo = rego.New(
			rego.Input(input),
			rego.Query(query))
	} else {
		rgo = rego.New(
			rego.Compiler(pa.compiler),
			rego.Input(input),
			rego.Query(query))
	}
	results, err := rgo.Eval(pa.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate rego: %s", err)
	}
	return pa.processRegoResultSet(packageBase, results)
}

func (pa *GKEPolicyAgent) GetPolicies() []*Policy {
	return pa.policies
}

func (pa *GKEPolicyAgent) processRegoResultSet(packageBase string, results rego.ResultSet) (*PolicyEvaluationResult, error) {
	pa.initEvalCache()
	evalResults := &PolicyEvaluationResult{}
	for i, result := range results {
		value, bindings, err := getResultDataForEval(result)
		if err != nil {
			log.Debugf("failed to get data from Rego result at index %d: %s", i, err)
			evalResults.Policies = append(evalResults.Policies, NewPolicyFromEvalResult(&RegoEvaluationResult{}, []error{err}))
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
		policyName := packageBase + "." + regoEvalResult.Name
		if compiledPolicy, ok := pa.evalCache[policyName]; ok {
			compiledPolicy.Valid = policy.Valid
			compiledPolicy.Violations = policy.Violations
			compiledPolicy.ProcessingErrors = policy.ProcessingErrors
			policy = compiledPolicy
		} else {
			log.Warnf("rego policy %q has no match with any compiled policy", policyName)
		}
		evalResults.Policies = append(evalResults.Policies, policy)
	}
	return evalResults, nil
}

func (pa *GKEPolicyAgent) initEvalCache() {
	pa.evalCache = make(map[string]*Policy)
	for _, policy := range pa.policies {
		policyCopy := *policy
		pa.evalCache[policyCopy.Name] = &policyCopy
	}
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

func (p *Policy) mapModule(module *ast.Module) {
	p.Name = module.Package.String()[8:]
	p.File = module.Package.Location.File
	for _, annot := range module.Annotations {
		if annot.Scope != "package" {
			continue
		}
		p.Title = annot.Title
		p.Description = annot.Description
		if group, ok := getStringFromInterfaceMap("group", annot.Custom); ok {
			p.Group = group
		}
		if severity, ok := getStringFromInterfaceMap("severity", annot.Custom); ok {
			p.Severity = severity
		}
		if category, ok := getStringFromInterfaceMap("sccCategory", annot.Custom); ok {
			p.Category = category
		}
		if cis, ok := annot.Custom["cis"]; ok {
			if cisMap, ok := cis.(map[string]interface{}); ok {
				if cisVersion, ok := getStringFromInterfaceMap("version", cisMap); ok {
					p.CisVersion = cisVersion
				}
				if cisID, ok := getStringFromInterfaceMap("id", cisMap); ok {
					p.CisID = cisID
				}
			}
		}
		if recommendation, ok := getStringFromInterfaceMap("recommendation", annot.Custom); ok {
			p.Recommendation = recommendation
		}
		if externalURI, ok := getStringFromInterfaceMap("externalURI", annot.Custom); ok {
			p.ExternalURI = externalURI
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
	if p.Severity == "" {
		errs = append(errs, "severity is not set")
	}
	if p.Category == "" {
		errs = append(errs, "category is not set")
	}
	if p.CisVersion != "" && p.CisID == "" {
		errs = append(errs, "CIS version is set without CIS identifier")
	}
	if p.CisID != "" && p.CisVersion == "" {
		errs = append(errs, "CIS identifier is set without CIS version")
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

func getRegoQueryForPackageBase(packageBase string) string {
	return "data." + packageBase + "[name]"
}

func getStringFromInterfaceMap(key string, m map[string]interface{}) (string, bool) {
	if value, ok := m[key]; ok {
		valueString, ok := value.(string)
		return valueString, ok
	}
	return "", false
}
