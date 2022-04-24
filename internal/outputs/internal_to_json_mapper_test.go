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
	"errors"
	"testing"

	"github.com/google/gke-policy-automation/internal/policy"
)

func TestMapValidPolicyToJson(t *testing.T) {
	r := policy.NewPolicyEvaluationResult()

	policyTitle := "title"
	policyDescription := "description"
	policyGroup := "group"
	policyValid := true

	r.AddPolicy(&policy.Policy{
		Title:            policyTitle,
		Description:      policyDescription,
		Group:            policyGroup,
		Valid:            policyValid,
		Violations:       []string{},
		ProcessingErrors: []error{},
	})

	result := MapClusterToJson(r)

	if result.ValidationResults[0].PolicyTitle != policyTitle {
		t.Errorf("policy title not mapped correctly: policyTitle = %s; want %s", result.ValidationResults[0].PolicyTitle, policyTitle)
	}
	if result.ValidationResults[0].PolicyDescription != policyDescription {
		t.Errorf("policy description not mapped correctly: policyDescription = %s; want %s", result.ValidationResults[0].PolicyDescription, policyDescription)
	}
	if result.ValidationResults[0].PolicyGroup != policyGroup {
		t.Errorf("policy group not mapped correctly: policyGroup = %s; want %s", result.ValidationResults[0].PolicyGroup, policyGroup)
	}
	if result.ValidationResults[0].IsValid == false {
		t.Errorf("policy not mapped correctly: IsValid = %t; want %t", result.ValidationResults[0].IsValid, policyValid)
	}
}

func TestMapViolatedPolicyToJson(t *testing.T) {
	r := policy.NewPolicyEvaluationResult()

	policyTitle := "title"
	policyDescription := "description"
	policyGroup := "group"
	policyValid := false
	violationMessage := "error"

	r.AddPolicy(&policy.Policy{
		Title:            policyTitle,
		Description:      policyDescription,
		Group:            policyGroup,
		Valid:            policyValid,
		Violations:       []string{violationMessage},
		ProcessingErrors: []error{},
	})

	result := MapClusterToJson(r)

	if result.ValidationResults[0].PolicyTitle != policyTitle {
		t.Errorf("policy title not mapped correctly: policyTitle = %s; want %s", result.ValidationResults[0].PolicyTitle, policyTitle)
	}
	if result.ValidationResults[0].PolicyDescription != policyDescription {
		t.Errorf("policy description not mapped correctly: policyDescription = %s; want %s", result.ValidationResults[0].PolicyDescription, policyDescription)
	}
	if result.ValidationResults[0].PolicyGroup != policyGroup {
		t.Errorf("policy group not mapped correctly: policyGroup = %s; want %s", result.ValidationResults[0].PolicyGroup, policyGroup)
	}
	if result.ValidationResults[0].IsValid == true {
		t.Errorf("policy not mapped correctly: IsValid = %t; want %t", result.ValidationResults[0].IsValid, policyValid)
	}
	if len(result.ValidationResults[0].Violations) != 1 {
		t.Errorf("policy not mapped correctly: violations length != %d", len(result.ValidationResults[0].Violations))
	}
	if result.ValidationResults[0].Violations[0].ErrorMessage != violationMessage {
		t.Errorf("policy not mapped correctly: violation error message != %s", violationMessage)
	}
}

func TestErroredMapping(t *testing.T) {
	r := policy.NewPolicyEvaluationResult()

	policyTitle := "title"
	policyDescription := "description"
	policyGroup := "group"
	errorMessage := "error"

	r.AddPolicy(&policy.Policy{
		Title:            policyTitle,
		Description:      policyDescription,
		Group:            policyGroup,
		ProcessingErrors: []error{errors.New(errorMessage)},
	})

	result := MapClusterToJson(r)

	if len(result.ProcessingErrors) != 1 {
		t.Errorf("policy not mapped correctly: errors length != %d", len(result.ProcessingErrors))
	}
	if result.ProcessingErrors[0].Error() != errorMessage {
		t.Errorf("policy not mapped correctly: error message != %s", errorMessage)
	}
}
