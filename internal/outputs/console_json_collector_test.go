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
	"bytes"
	"testing"

	"github.com/google/gke-policy-automation/internal/policy"
)

func TestConsoleJsonResultCollector(t *testing.T) {
	var buff bytes.Buffer
	var expectedJSON = "json string"

	out := &Output{w: &buff}
	reportMapperMock := &validationReportMapperMock{
		addResultsFn: func(results []*policy.PolicyEvaluationResult) {},
		getJSONReportFn: func() ([]byte, error) {
			return []byte(expectedJSON), nil
		},
	}

	collector := &consoleJSONResultCollector{out: out, reportMapper: reportMapperMock}
	err := collector.RegisterResult([]*policy.PolicyEvaluationResult{{}})
	if err != nil {
		t.Fatalf("err on RegisterResult = %v; want nil", err)
	}
	err = collector.Close()
	if err != nil {
		t.Fatalf("err on Close = %v; want nil", err)
	}
	if len(buff.String()) <= 0 {
		t.Errorf("nothing was written to the output buffer")
	}
	if buff.String() != expectedJSON {
		t.Errorf("expected \"%v\" but buffer was \"%v\"", expectedJSON, buff.String())
	}
}
