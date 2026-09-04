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
	"net/url"
	"strings"
	"unicode"

	"github.com/fatih/color"
	"github.com/google/gke-policy-automation/internal/gke"
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
	p.out.InitTabs(0, 4)
	for i, policy := range report.Policies {
		severityf := severitySprintfFunc(policy.Severity)
		ruleTitleF := color.New(color.Bold, color.FgHiWhite).SprintfFunc()
		p.out.Printf("%s #%d %s %s",
			IconMagnifier,
			i+1,
			severityf("%s", strings.ToUpper(sanitizeForTerminal(policy.Severity))),
			ruleTitleF("%s", sanitizeForTerminal(policy.PolicyTitle)),
		)
		if safeURI := sanitizeExternalURI(policy.ExternalURI); safeURI != "" {
			extURI := fmt.Sprintf("(\x1b]8;;%s\x07%s\x1b]8;;\x07)", safeURI, "documentation")
			p.out.Printf(" %s\n", extURI)
		} else {
			p.out.Printf("\n")
		}
		for _, evaluation := range policy.ClusterEvaluations {
			statusf := evalStatusSprintfFunc(*evaluation)
			clusterDataf := color.New(color.FgCyan).Sprintf
			project, location, cluster := gke.MustSliceClusterID(evaluation.ClusterID)

			p.out.TabPrintf("   - projects/%s/locations/%s/clusters/%s\t[%s]\n",
				clusterDataf("%s", project),
				clusterDataf("%s", location),
				clusterDataf("%s", cluster),
				statusf("%s", evalStatusString(*evaluation)),
			)

			if !evaluation.Valid {
				violationF := color.New(color.Italic, color.FgRed).Sprintf
				for _, violation := range evaluation.Violations {
					p.out.TabPrintf("      %s\t\n",
						violationF("%s %s", IconMiddleDot, sanitizeForTerminal(violation)),
					)
				}
			}
			log.Infof("Policy: %s, Cluster: %s, Valid: %v", policy.PolicyName, evaluation.ClusterID, evaluation.Valid)
		}
		p.out.TabFlush()
		p.out.Printf("\n")
	}
	summaryf := color.New(color.Bold, color.FgHiWhite).Sprintf
	p.out.Printf("%s %s",
		IconInfo,
		summaryf("Evaluated %d policies on %d clusters\n", len(report.Policies), len(report.ClusterStats)),
	)
	p.out.InitTabs(0, 2)
	for _, stat := range report.ClusterStats {
		clusterDataf := color.New(color.FgCyan).Sprintf
		criticalf := color.New(color.FgHiRed).Sprintf
		highf := color.New(color.FgRed).Sprintf
		mediumf := color.New(color.FgYellow).Sprintf
		lowf := color.New(color.FgHiWhite).Sprintf
		project, location, cluster := gke.MustSliceClusterID(stat.ClusterID)

		p.out.TabPrintf("  - projects/%s/locations/%s/clusters/%s\t: %s, %s, %s, %s\n",
			clusterDataf("%s", project),
			clusterDataf("%s", location),
			clusterDataf("%s", cluster),
			criticalf("%d Critical", stat.ViolatedCriticalCount),
			highf("%d High", stat.ViolatedHighCount),
			mediumf("%d Medium", stat.ViolatedMediumCount),
			lowf("%d Low", stat.ViolatedLowCount),
		)
	}
	p.out.TabFlush()
	p.out.Printf("\n")
	return nil
}

func (p *consoleResultCollector) Name() string {
	return "console"
}

// sanitizeForTerminal neutralizes control characters and escape sequences
// (C0, C1, ANSI CSI, OSC) from untrusted policy metadata and messages.
// Whitespace control characters (\t, \r, \n) are converted to spaces to avoid
// breaking tabwriter columns or joining words together.
func sanitizeForTerminal(s string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r == '\t' || r == '\n' || r == '\r':
			return ' '
		case r < 0x20 || r == 0x7f || (r >= 0x80 && r <= 0x9f):
			return -1
		case !unicode.IsPrint(r):
			return -1
		}
		return r
	}, s)
}

// sanitizeExternalURI ensures the external URI is a clean, absolute HTTP/HTTPS URL
// before embedding it in an OSC-8 terminal hyperlink. If the raw string contains
// control characters or invalid schemes, it is rejected.
func sanitizeExternalURI(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || sanitizeForTerminal(trimmed) != trimmed {
		return ""
	}
	u, err := url.ParseRequestURI(trimmed)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return ""
	}
	return u.String()
}

type sprintfFunc func(format string, a ...interface{}) string

func severitySprintfFunc(severity string) sprintfFunc {
	var sevColor []color.Attribute
	switch strings.ToLower(severity) {
	case "critical":
		sevColor = []color.Attribute{color.Bold, color.FgHiRed}
	case "high":
		sevColor = []color.Attribute{color.FgHiRed}
	case "medium":
		sevColor = []color.Attribute{color.Bold, color.FgHiYellow}
	default:
		sevColor = []color.Attribute{color.FgHiWhite}
	}
	return color.New(sevColor...).SprintfFunc()
}

func evalStatusSprintfFunc(e ValidationReportClusterEvaluation) sprintfFunc {
	if e.Errored {
		return color.New(color.Bold, color.FgHiYellow).Sprintf
	}
	if e.Valid {
		return color.New(color.Bold, color.FgHiGreen).Sprintf
	}
	return color.New(color.Bold, color.FgHiRed).Sprintf
}

func evalStatusString(e ValidationReportClusterEvaluation) string {
	if e.Errored {
		return " ERROR "
	}
	if e.Valid {
		return " VALID "
	}
	return "INVALID"
}
