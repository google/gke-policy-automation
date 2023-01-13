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
	defaultGoroutinesNo = 20
	collectorName       = "Security Command Center"
)

type sccCollector struct {
	ctx          context.Context
	cli          scc.SecurityCommandCenterClient
	createSource bool
	goRoutinesNo int
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
		goRoutinesNo: defaultGoroutinesNo,
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
			finding := mapPolicyToFinding(result.ClusterID, eventTime, policy)
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
	log.Debugf("using %d maxGoRoutines", c.goRoutinesNo)
	findingsChan := make(chan *scc.Finding, c.goRoutinesNo)
	errorsChan := make(chan error, c.goRoutinesNo)

	log.Debugf("starting finding producing goroutine")
	go func() {
		for _, finding := range c.findings {
			findingsChan <- finding
		}
		close(findingsChan)
	}()

	log.Debugf("starting finding consuming goroutine")
	go func() {
		var wg sync.WaitGroup
		for i := 0; i < c.goRoutinesNo; i++ {
			wg.Add(1)
			go c.upsertFinding(i, &wg, findingsChan, source, errorsChan)
		}
		wg.Wait()
		close(errorsChan)
	}()
	log.Debugf("processing errors")
	var errors []error
	for err := range errorsChan {
		errors = append(errors, err)
	}
	return errors
}

func (c *sccCollector) upsertFinding(i int, wg *sync.WaitGroup, findings chan *scc.Finding, source string, errors chan error) {
	defer wg.Done()
	for finding := range findings {
		log.Debugf("goroutine %d processing finding (resName=%v category=%v)", i, finding.ResourceName, finding.Category)
		if err := c.cli.UpsertFinding(source, finding); err != nil {
			log.Warnf("failed to upsert finding (resName=%v category=%v): %s", finding.ResourceName, finding.Category, err)
			errors <- err
		}
	}
	log.Debugf("Upsert goroutine %d finished", i)
}

func mapPolicyToFinding(resourceName string, eventTime time.Time, policy *policy.Policy) *scc.Finding {
	return &scc.Finding{
		Time:              eventTime,
		ResourceName:      fmt.Sprintf("//container.googleapis.com/%s", resourceName),
		Category:          policy.Category,
		Description:       policy.Description,
		State:             mapPolicyEvaluationToFindingState(policy),
		Severity:          strings.ToUpper(policy.Severity),
		SourcePolicyName:  policy.Name,
		SourcePolicyFile:  policy.File,
		SourcePolicyGroup: policy.Group,
		CisVersion:        policy.CisVersion,
		CisID:             policy.CisID,
		ExternalURI:       policy.ExternalURI,
		Recommendation:    policy.Recommendation,
	}
}

func mapPolicyEvaluationToFindingState(policy *policy.Policy) string {
	if policy.Valid {
		return scc.FindingStateStringInactive
	}
	if len(policy.Violations) > 0 {
		return scc.FindingStateStringActive
	}
	return scc.FindingStateStringUnspecified
}
