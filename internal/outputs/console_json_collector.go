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

type consoleJsonResultCollector struct {
	out          *Output
	reportMapper ValidationReportMapper
}

func NewConsoleJsonResultCollector(output *Output) ValidationResultCollector {
	return &consoleJsonResultCollector{
		out:          output,
		reportMapper: NewValidationReportMapper(),
	}
}

func (p *consoleJsonResultCollector) RegisterResult(results []*policy.PolicyEvaluationResult) error {
	p.reportMapper.AddResults(results)
	return nil
}

func (p *consoleJsonResultCollector) Close() error {

	jsonResult, err := p.reportMapper.GetJsonReport()

	if err != nil {
		return err
	}

	p.out.Printf(string(jsonResult))
	return nil
}

func (p *consoleJsonResultCollector) Name() string {
	return "console json"
}
