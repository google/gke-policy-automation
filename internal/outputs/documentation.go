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
	"sort"
	"strings"

	"github.com/google/gke-policy-automation/internal/config"
	"github.com/google/gke-policy-automation/internal/policy"
)

type PolicyDocumentation interface {
	GenerateDocumentation() string
}

type MarkdownPolicyDocumentation struct {
	policies []*policy.Policy
}

type DocumentationBuilder func(policies []*policy.Policy) PolicyDocumentation

func NewMarkdownPolicyDocumentation(policies []*policy.Policy) PolicyDocumentation {
	return &MarkdownPolicyDocumentation{policies}
}

func (m *MarkdownPolicyDocumentation) GenerateDocumentation() string {
	sort.SliceStable(m.policies, func(i, j int) bool {
		return m.policies[i].Group < m.policies[j].Group
	})
	var sb strings.Builder

	sb.WriteString("# Available Policies\n\n|Group|`Title|Description|File|\n|-|-|-|-|")

	for _, p := range m.policies {
		policyFileURL := fmt.Sprintf("%s/blob/%s/%s", config.DefaultGitRepository, config.DefaultGitBranch, p.File)
		sb.WriteString(fmt.Sprintf("\n |%s|%s|%s|[%s](%s)|", p.Group, p.Title, p.Description, p.File, policyFileURL))
	}

	return sb.String()
}
