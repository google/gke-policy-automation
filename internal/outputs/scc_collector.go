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
	"sync"
	"time"

	"github.com/google/gke-policy-automation/internal/log"
	"github.com/google/gke-policy-automation/internal/outputs/scc"
	"github.com/google/gke-policy-automation/internal/policy"
)

const (
	defaultNoThreads = 5
	collectorName    = "Security Command Center"
)

type sccCollector struct {
	ctx          context.Context
	cli          scc.SecurityCommandCenterClient
	createSource bool
	threadsNo    int
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
	return newSccCollector(ctx, createSource, cli), nil
}

func newSccCollector(ctx context.Context, createSource bool, cli scc.SecurityCommandCenterClient) ValidationResultCollector {
	return &sccCollector{
		ctx:          ctx,
		cli:          cli,
		threadsNo:    defaultNoThreads,
		createSource: createSource,
	}
}

func (c *sccCollector) Close() error {
	source, err := c.getSccSource()
	if err != nil {
		return err
	}
	errors := c.processFindings(source)
	if len(errors) > 0 {
		return fmt.Errorf("failed to upsert all findings: %d out of %d failed", len(errors), len(c.findings))
	}
	err = c.cli.Close()
	return err
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

func (c *sccCollector) Name() string {
	return collectorName
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
	newSource, err := c.cli.CreateSource()
	if err != nil {
		return "", err
	}
	return newSource, nil
}

func (c *sccCollector) processFindings(source string) []error {
	divided := divideFindings(c.threadsNo, c.findings)
	var wg sync.WaitGroup
	results := make(chan error, len(c.findings))
	log.Debugf("Starting %d upsert goroutines", len(divided))
	for i := range divided {
		wg.Add(1)
		go c.upsertFindings(i, &wg, divided[i], source, results)
	}
	log.Debugf("waiting for upsert goroutines to finish")
	wg.Wait()
	log.Debugf("all upsert goroutines finished")
	close(results)
	log.Debugf("results channel closed")
	var errors []error
	for result := range results {
		errors = append(errors, result)
	}
	return errors
}

func (c *sccCollector) upsertFindings(i int, wg *sync.WaitGroup, findings []*scc.Finding, source string, results chan error) {
	defer wg.Done()
	log.Debugf("Upsert goroutine %d starting", i)
	for i := range findings {
		if err := c.cli.UpsertFinding(source, findings[i]); err != nil {
			log.Warnf("failed to upsert finding %+v: %s", findings[i], err)
			results <- err
		}
	}
	log.Debugf("Upsert goroutine %d finished", i)
}

func divideFindings(chunksNo int, findings []*scc.Finding) [][]*scc.Finding {
	var result [][]*scc.Finding
	chunkSize := len(findings) / chunksNo
	for i := 0; i < len(findings); i += chunkSize {
		end := i + chunkSize
		if end > len(findings) {
			end = len(findings)
		}
		result = append(result, findings[i:end])
	}
	return result
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
