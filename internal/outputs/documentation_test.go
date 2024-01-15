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
			CisVersion:  "1.2",
			CisID:       "5.3.1",
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
	fmt.Fprintf(&sb, "|[First policy](%sgke-policies/file1.rego)|Group 1|First description||\n", defaultPolicyDocFileURLPrefix)
	fmt.Fprintf(&sb, "|[Third policy](%sgke-policies/file3.rego)|Group 1|Third description|[CIS GKE](%s) 1.2: 5.3.1|\n", defaultPolicyDocFileURLPrefix, cisGKEURL)
	fmt.Fprintf(&sb, "|[Second policy](%sgke-policies/file2.rego)|Group 2|Second description||\n", defaultPolicyDocFileURLPrefix)
	expected := sb.String()

	generator := NewMarkdownPolicyDocumentation(buildPoliciesMetadata())
	documentation := generator.GenerateDocumentation()

	if !strings.HasSuffix(documentation, expected) {
		t.Fatalf("documentation is %v; want suffix to be %v", documentation, expected)
	}
}
