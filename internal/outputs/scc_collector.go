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
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/gke-policy-automation/internal/log"
	"github.com/google/gke-policy-automation/internal/outputs/scc"
	"github.com/google/gke-policy-automation/internal/policy"
)

type sccCollector struct {
	ctx          context.Context
	cli          scc.SecurityCommandCenterClient
	createSource bool
	findings     []*scc.Finding
}

func NewSccCollector(ctx context.Context, orgNumber string, createSource bool, credsFile string) (ValidationResultCollector, error) {
	var cli scc.SecurityCommandCenterClient
	var err error
	if credsFile != "" {
		cli, err = scc.NewSecurityCommandCenterClient(ctx, orgNumber)
	} else {
		cli, err = scc.NewSecurityCommandCenterClientWithCredentialsFile(ctx, orgNumber, credsFile)
	}
	if err != nil {
		return nil, err
	}
	return &sccCollector{
		ctx:          ctx,
		cli:          cli,
		createSource: createSource,
	}, nil
}

func (c *sccCollector) Close() error {
	source, err := c.getSccSource()
	if err != nil {
		return err
	}
	errors := make([]error, 0)
	for _, finding := range c.findings {
		if err := c.cli.UpsertFinding(source, finding); err != nil {
			log.Warnf("failed to upsert finding %v: %s", finding, err)
			errors = append(errors, err)
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf("failed to upsert all findings: %d out of %d failed", len(errors), len(c.findings))
	}
	return nil
}

func (c *sccCollector) RegisterResult(results []*policy.PolicyEvaluationResult) error {
	eventTime := time.Now()
	for _, result := range results {
		for _, policy := range result.Policies {
			finding := mapPolicyToFinding(result.ClusterName, eventTime, policy)
			c.findings = append(c.findings, finding)
		}
	}
	return nil
}

func (c *sccCollector) getSccSource() (string, error) {
	source, err := c.cli.FindSource()
	if err != nil {
		return "", err
	}
	if source != nil {
		return *source, nil
	}
	log.Debugf("SCC source was not found")
	if !c.createSource {
		return "", errors.New("SCC source was not found and its provisioning is disabled")
	}
	log.Debugf("Creating SCC source")
	*source, err = c.cli.CreateSource()
	if err != nil {
		return "", err
	}
	return *source, nil
}

func mapPolicyToFinding(resourceName string, eventTime time.Time, policy *policy.Policy) *scc.Finding {
	return &scc.Finding{
		Time:         eventTime,
		ResourceName: fmt.Sprintf("//container.googleapis.com/%s", resourceName),
		Category:     policy.Category,
		Description:  policy.Description,
		State:        mapPolicyEvaluationToFindingState(policy),
		Severity:     strings.ToUpper(policy.Severity),
	}
}

func mapPolicyEvaluationToFindingState(policy *policy.Policy) string {
	if policy.Valid {
		return scc.FINDING_STATE_STRING_INACTIVE
	}
	if len(policy.Violations) > 0 {
		return scc.FINDING_STATE_STRING_ACTIVE
	}
	return scc.FINDING_STATE_STRING_UNSPECIFIED
}
