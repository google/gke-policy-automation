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
	"text/tabwriter"

	"github.com/google/gke-policy-automation/internal/policy"
)

func TestConsoleResultCollector(t *testing.T) {
	var buff bytes.Buffer
	out := &Output{w: &buff, tabWriter: tabwriter.NewWriter(&buff, 0, 0, 0, '\t', tabwriter.AlignRight)}
	reportMapperMock := &validationReportMapperMock{
		addResultsFn: func(results []*policy.PolicyEvaluationResult) {},
		getReportFn: func() *ValidationReport {
			return &ValidationReport{
				Policies: []*ValidationReportPolicy{
					{
						PolicyName:        "test-policy",
						PolicyGroup:       "test-group",
						PolicyTitle:       "test-title",
						PolicyDescription: "test-desc",
						ClusterEvaluations: []*ValidationReportClusterEvaluation{
							{ClusterID: "projects/test-proj/locations/europe-central2/clusters/cluster-one", Valid: true},
							{ClusterID: "projects/test-proj/locations/europe-central2/clusters/cluster-two", Valid: false, Violations: []string{"violation"}},
						},
					},
				},
				ClusterStats: []*ValidationReportClusterStats{
					{ClusterID: "projects/test-proj/locations/europe-central2/clusters/cluster-one", ValidPoliciesCount: 1},
				},
			}
		},
	}

	collector := &consoleResultCollector{out: out, reportMapper: reportMapperMock}
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
}

func TestSanitizeForTerminal(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "plain text",
			input:    "Cluster is missing private endpoint",
			expected: "Cluster is missing private endpoint",
		},
		{
			name:     "osc-52 clipboard payload with ESC and BEL",
			input:    "\x1b]52;c;Y3VybCBldmlsLnNoIHwgYmFzaA==\x07",
			expected: "]52;c;Y3VybCBldmlsLnNoIHwgYmFzaA==",
		},
		{
			name:     "ansi csi escape sequences neutralized",
			input:    "\x1b[2K\x1b[1Ahello\x1b[0m",
			expected: "[2K[1Ahello[0m",
		},
		{
			name:     "whitespace controls normalized to spaces",
			input:    "line1\r\nline2\twith\ttabs",
			expected: "line1  line2 with tabs",
		},
		{
			name:     "c0 and c1 control codes and DEL",
			input:    "foo\x00\x08\x7f\u0080\u009b\u009dbar",
			expected: "foobar",
		},
		{
			name:     "non-printable unicode filtered (e.g. bidi override, zero-width space)",
			input:    "safe\u202Ereversed\u200Btext",
			expected: "safereversedtext",
		},
		{
			name:     "printable unicode preserved",
			input:    "Cluster GKE-1 测试 🚀",
			expected: "Cluster GKE-1 测试 🚀",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := sanitizeForTerminal(tc.input)
			if got != tc.expected {
				t.Errorf("sanitizeForTerminal(%q) = %q; want %q", tc.input, got, tc.expected)
			}
		})
	}
}

func TestSanitizeExternalURI(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid https url",
			input:    "https://cloud.google.com/kubernetes-engine/docs",
			expected: "https://cloud.google.com/kubernetes-engine/docs",
		},
		{
			name:     "valid http url with whitespace trimmed",
			input:    "  http://example.com/docs/gke  ",
			expected: "http://example.com/docs/gke",
		},
		{
			name:     "javascript uri scheme rejected",
			input:    "javascript:alert(1)",
			expected: "",
		},
		{
			name:     "file uri scheme rejected",
			input:    "file:///etc/passwd",
			expected: "",
		},
		{
			name:     "data uri scheme rejected",
			input:    "data:text/html;base64,PHNjcmlwdD5hbGVydCgxKTwvc2NyaXB0Pg==",
			expected: "",
		},
		{
			name:     "relative path rejected",
			input:    "/docs/user-guide",
			expected: "",
		},
		{
			name:     "injected control characters rejected",
			input:    "https://cloud.google.com\x07\x1b]52;c;evil\x07",
			expected: "",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := sanitizeExternalURI(tc.input)
			if got != tc.expected {
				t.Errorf("sanitizeExternalURI(%q) = %q; want %q", tc.input, got, tc.expected)
			}
		})
	}
}

func TestConsoleResultCollector_SecuritySanitization(t *testing.T) {
	var buff bytes.Buffer
	out := &Output{w: &buff, tabWriter: tabwriter.NewWriter(&buff, 0, 0, 0, '\t', tabwriter.AlignRight)}

	// Malicious payload simulating OSC-52 clipboard hijacking and OSC-8 breakout
	osc52Payload := "\x1b]52;c;Y3VybCBldmlsLnNoIHwgYmFzaA==\x07"
	csiPayload := "\x1b[2K\x1b[1A"

	reportMapperMock := &validationReportMapperMock{
		addResultsFn: func(results []*policy.PolicyEvaluationResult) {},
		getReportFn: func() *ValidationReport {
			return &ValidationReport{
				Policies: []*ValidationReportPolicy{
					{
						PolicyName:  "hostile-policy",
						PolicyGroup: "hostile-group",
						PolicyTitle: "Title with " + csiPayload + " escape",
						Severity:    "high\x1b[31m",
						ExternalURI: "javascript:alert(document.cookie)",
						ClusterEvaluations: []*ValidationReportClusterEvaluation{
							{
								ClusterID:  "projects/test-proj/locations/us-central1/clusters/cluster-one",
								Valid:      false,
								Violations: []string{"Violation: " + osc52Payload + " embedded\ttab\r\nnewline"},
							},
						},
					},
					{
						PolicyName:  "safe-policy",
						PolicyTitle: "Safe Policy",
						Severity:    "medium",
						ExternalURI: "https://cloud.google.com/kubernetes-engine",
						ClusterEvaluations: []*ValidationReportClusterEvaluation{
							{
								ClusterID: "projects/test-proj/locations/us-central1/clusters/cluster-one",
								Valid:     true,
							},
						},
					},
				},
				ClusterStats: []*ValidationReportClusterStats{
					{ClusterID: "projects/test-proj/locations/us-central1/clusters/cluster-one", ViolatedHighCount: 1},
				},
			}
		},
	}

	collector := &consoleResultCollector{out: out, reportMapper: reportMapperMock}
	if err := collector.RegisterResult([]*policy.PolicyEvaluationResult{{}}); err != nil {
		t.Fatalf("RegisterResult err = %v; want nil", err)
	}
	if err := collector.Close(); err != nil {
		t.Fatalf("Close err = %v; want nil", err)
	}

	output := buff.String()

	// Verify hostile payloads are neutralized
	if bytes.Contains([]byte(output), []byte("\x1b]52;")) {
		t.Errorf("output contains unescaped OSC-52 sequence: %q", output)
	}
	if bytes.Contains([]byte(output), []byte("Y3VybCBldmlsLnNoIHwgYmFzaA==\x07")) {
		t.Errorf("output contains untrusted BEL character from payload: %q", output)
	}
	if bytes.Contains([]byte(output), []byte("\x1b[2K")) || bytes.Contains([]byte(output), []byte("\x1b[1A")) {
		t.Errorf("output contains unescaped CSI sequence from policy title: %q", output)
	}
	if bytes.Contains([]byte(output), []byte("javascript:")) {
		t.Errorf("output contains unsanitized javascript: URI: %q", output)
	}

	// Verify the safe policy's documentation link is preserved and formatted as OSC-8
	expectedLink := "\x1b]8;;https://cloud.google.com/kubernetes-engine\x07documentation\x1b]8;;\x07"
	if !bytes.Contains([]byte(output), []byte(expectedLink)) {
		t.Errorf("output missing safe documentation OSC-8 link (%q): %q", expectedLink, output)
	}
}
