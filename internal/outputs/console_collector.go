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

	"github.com/google/gke-policy-automation/internal/log"
	"github.com/google/gke-policy-automation/internal/policy"
)

type consoleResultCollector struct {
	out          *Output
	reportMapper ValidationReportMapper
}

func NewConsoleResultCollector(output *Output) ValidationResultCollector {
	return &consoleResultCollector{
		out:          output,
		reportMapper: NewValidationReportMapper(),
	}
}

func (p *consoleResultCollector) RegisterResult(results []*policy.PolicyEvaluationResult) error {
	p.reportMapper.AddResults(results)
	return nil
}

func (p *consoleResultCollector) Close() error {
	report := p.reportMapper.GetReport()
	p.out.Printf("\n")
	p.out.InitTabs(95)
	for _, policy := range report.Policies {
		policyTitle := policy.PolicyTitle
		if policy.ExternalURI != "" {
			policyTitle = fmt.Sprintf("%s \x1b]8;;%s\x07%s\x1b]8;;\x07", ICON_HYPERLINK, policy.ExternalURI, policy.PolicyTitle)
		}
		p.out.ColorPrintf("%s [bold][light_gray][%s][yellow] %s[reset]: %s\n", ICON_MAGNIFIER, policy.PolicyGroup, policy.PolicyName, policyTitle)

		for _, evaluation := range policy.ClusterEvaluations {
			statusString := "[ \033[1m\033[32mOK\033[0m ]"
			if !evaluation.Valid {
				statusString = "[\033[1m\033[31mFAIL\033[0m]"
			}
			p.out.TabPrintf("  - %s\t"+statusString+"\n", evaluation.ClusterID)
			if !evaluation.Valid {
				for _, violation := range evaluation.Violations {
					p.out.TabPrintf("    \033[1m\033[31m%s\033[0m\t\n", violation)
				}
			}
			log.Infof("Policy: %s, Cluster: %s, Valid: %v", policy.PolicyName, evaluation.ClusterID, evaluation.Valid)
		}
		p.out.TabFlush()
		p.out.Printf("\n")
	}
	p.out.ColorPrintf("%s [light_gray][bold]Evaluated %d policies on %d clusters\n", ICON_INFO, len(report.Policies), len(report.ClusterStats))
	p.out.InitTabs(0)
	for _, stat := range report.ClusterStats {
		p.out.TabPrintf("  - %s:\t\033[32m%d valid, \033[31m%d violated, \033[33m%d errored\033[0m\n", stat.ClusterID, stat.ValidPoliciesCount, stat.ViolatedPoliciesCount, stat.ErroredPoliciesCount)
	}
	p.out.TabFlush()
	p.out.Printf("\n")
	return nil
}

func (p *consoleResultCollector) Name() string {
	return "console"
}
