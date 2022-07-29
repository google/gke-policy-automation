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
	"strings"
	"testing"

	"github.com/google/gke-policy-automation/internal/policy"
)

func buildPoliciesMetadata() []*policy.Policy {

	return []*policy.Policy{
		{
			Title:       "First policy",
			Description: "First description",
			Group:       "Group 1",
			File:        "file1.rego",
		},
		{
			Title:       "Second policy",
			Description: "Second description",
			Group:       "Group 2",
			File:        "file2.rego",
		},
	}
}

func TestMarkdownDocumention(t *testing.T) {

	expected := "\n |Group 1|First policy|First description|[file1.rego](https://github.com/google/gke-policy-automation/blob/main/file1.rego)|\n |Group 2|Second policy|Second description|[file2.rego](https://github.com/google/gke-policy-automation/blob/main/file2.rego)|"

	generator := NewMarkdownPolicyDocumentation(buildPoliciesMetadata())
	documentation := generator.GenerateDocumentation()

	if !strings.HasSuffix(documentation, expected) {
		t.Fatalf("documentation is %v; want suffix to be %v", documentation, expected)
	}
}
