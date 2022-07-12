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
	"os"
	"testing"

	"github.com/google/gke-policy-automation/internal/policy"
)

type MockFileWriter struct {
}

func (f MockFileWriter) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return nil
}

func TestCollectingToJson(t *testing.T) {

	evalResults := make([]*policy.PolicyEvaluationResult, 0)
	r := &policy.PolicyEvaluationResult{}

	policyTitle := "title"
	policyDescription := "description"
	policyGroup := "group"
	policyValid := true

	r.Policies = append(r.Policies, &policy.Policy{
		Title:            policyTitle,
		Description:      policyDescription,
		Group:            policyGroup,
		Valid:            policyValid,
		Violations:       []string{},
		ProcessingErrors: []error{},
	})

	evalResults = append(evalResults, r)

	var collector = NewJSONResultToCustomWriterCollector("sample.json", MockFileWriter{})

	err := collector.RegisterResult(evalResults)

	if err != nil {
		t.Errorf("registering result failed: %s", err)
	}

	err = collector.Close()

	if err != nil {
		t.Errorf("json serialization failed: %s", err)
	}
}
