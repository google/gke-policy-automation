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
	"github.com/google/gke-policy-automation/internal/gke"
	"github.com/google/gke-policy-automation/internal/policy"
)

type ValidationResultCollector interface {
	RegisterResult(results []*policy.PolicyEvaluationResult) error
	Close() error
	Name() string
}

type ClusterDumpCollector interface {
	RegisterCluster(cluster *gke.Cluster)
	Close() error
}
