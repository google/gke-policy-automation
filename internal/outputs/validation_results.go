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

import "time"

type ValidationResults struct {
	ValidationDate           time.Time                 `json:"validationDate"`
	ClusterValidationResults []ClusterValidationResult `json:"clusters"`
}

type ClusterValidationResult struct {
	ClusterPath       string                   `json:"cluster"`
	ValidationResults []PolicyValidationResult `json:"result"`
	ProcessingErrors  []error                  `json:"processingErrors"`
}

type PolicyValidationResult struct {
	PolicyGroup       string      `json:"policyGroup"`
	PolicyTitle       string      `json:"policyName"`
	PolicyDescription string      `json:"policyDescription"`
	IsValid           bool        `json:"isValid"`
	Violations        []Violation `json:"violations"`
}

type Violation struct {
	ErrorMessage string `json:"errorMessage"`
}
