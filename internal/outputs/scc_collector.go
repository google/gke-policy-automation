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
	"context"
	"time"

	"github.com/google/gke-policy-automation/internal/outputs/scc"
	"github.com/google/gke-policy-automation/internal/policy"
)

type sccCollector struct {
	ctx context.Context
	cli scc.SecurityCommandCenterClient
}

func NewSccCollector(ctx context.Context, orgNumber string) (ValidationResultCollector, error) {
	cli, err := scc.NewSecurityCommandCenterClient(ctx, orgNumber)
	if err != nil {
		return nil, err
	}
	return &sccCollector{
		ctx: ctx,
		cli: cli,
	}, nil
}

// Close implements ValidationResultCollector
func (c *sccCollector) Close() error {
	panic("unimplemented")
}

// RegisterResult implements ValidationResultCollector
func (c *sccCollector) RegisterResult(results []*policy.PolicyEvaluationResult) error {
	return nil
}

func mapResultToFinding(result *policy.PolicyEvaluationResult) *scc.Finding {
	return &scc.Finding{
		Time:         time.Now(),
		SourceName:   "TODO",
		ResourceName: "TODO",
		Category:     "TODO",
		Description:  "TODO",
		State:        "TODO",
		Severity:     "TODO",
	}
}
