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
	for _, policy := range report.Policies {
		p.out.ColorPrintf("\U0001f50e [bold][white][%s][yellow] %s[reset]: %s\n", policy.PolicyGroup, policy.PolicyName, policy.PolicyTitle)
		for _, evaluation := range policy.ClusterEvaluations {
			statusString := "[reset][ [bold][green]OK[reset] ]\n"
			if !evaluation.Valid {
				statusString = "[reset][[bold][red]FAIL[reset]]\n"
			}
			p.out.ColorPrintf("\t- %s\t\t\t"+statusString, evaluation.ClusterID)
			if !evaluation.Valid {
				for _, violation := range evaluation.Violations {
					p.out.ColorPrintf("\t  [bold][red]%s\n", violation)
				}
			}
		}
		p.out.Printf("\n")
	}
	p.out.ColorPrintf("\u2139 [white][bold]Evaluated %d policies on %d clusters\n", len(report.Policies), len(report.ClusterStats))
	for _, stat := range report.ClusterStats {
		p.out.ColorPrintf(" - %s: [green]%d valid, [red]%d violated, [yellow]%d errored\n", stat.ClusterID, stat.ValidPoliciesCount, stat.ViolatedPoliciesCount, stat.ErroredPoliciesCount)
	}
	return nil
}
