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
	"fmt"
	"strings"
	"testing"

	"github.com/google/gke-policy-automation/internal/policy"
)

func buildPoliciesMetadata() []*policy.Policy {

	return []*policy.Policy{
		{
			Title:       "Second policy",
			Description: "Second description",
			Group:       "Group 2",
			File:        "gke-policies/file2.rego",
		},
		{
			Title:       "Third policy",
			Description: "Third description",
			Group:       "Group 1",
			File:        "gke-policies/file3.rego",
		},
		{
			Title:       "First policy",
			Description: "First description",
			Group:       "Group 1",
			File:        "gke-policies/file1.rego",
		},
	}
}

func TestMarkdownDocumention(t *testing.T) {
	var sb strings.Builder
	fmt.Fprintf(&sb, "|Group 1|First policy|First description|[gke-policies/file1.rego](%sgke-policies/file1.rego)|\n", defaultPolicyDocFileURLPrefix)
	fmt.Fprintf(&sb, "|Group 1|Third policy|Third description|[gke-policies/file3.rego](%sgke-policies/file3.rego)|\n", defaultPolicyDocFileURLPrefix)
	fmt.Fprintf(&sb, "|Group 2|Second policy|Second description|[gke-policies/file2.rego](%sgke-policies/file2.rego)|\n", defaultPolicyDocFileURLPrefix)
	expected := sb.String()

	generator := NewMarkdownPolicyDocumentation(buildPoliciesMetadata())
	documentation := generator.GenerateDocumentation()

	if !strings.HasSuffix(documentation, expected) {
		t.Fatalf("documentation is %v; want suffix to be %v", documentation, expected)
	}
}
