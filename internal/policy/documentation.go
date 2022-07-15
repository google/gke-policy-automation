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
	"fmt"
	"strings"
)

type PolicyDocumentation interface {
	GenerateDocumentation() string
}

type MarkdownPolicyDocumentation struct {
	policies []*Policy
}

type DocumentationBuilder func(policies []*Policy) PolicyDocumentation

func NewMarkdownPolicyDocumentation(policies []*Policy) PolicyDocumentation {
	return &MarkdownPolicyDocumentation{policies}
}

func (m *MarkdownPolicyDocumentation) GenerateDocumentation() string {

	var sb strings.Builder

	sb.WriteString("# Available Policies\n\n|Title|Description|Group|File|\n|-|-|-|-|")

	for _, p := range m.policies {
		sb.WriteString(fmt.Sprintf("\n |%s|%s|%s|%s|", p.Title, p.Description, p.Group, p.File))
	}

	return sb.String()
}
