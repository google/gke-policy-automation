# Copyright 2022 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# METADATA
# title: Number of secrets with KMS encryption
# description: The total number of secrets when KMS secret encryption is enabled
# custom:
#   group: Scalability
#   severity: High
#   recommendation: >
#     A cluster must decrypt all Secrets during cluster startup when application-layer secrets encryption is enabled.
#     If the number of secrets you store is above the limit, your cluster might become unstable during startup or upgrades, causing workload outages.
#   externalURI: https://cloud.google.com/kubernetes-engine/docs/concepts/planning-large-clusters#limits-best-practices-large-scale-clusters
#   sccCategory: SECRETS_WITH_ENCRYPTION_LIMIT
#   dataSource: monitoring, gke

package gke.scalability.secrets_with_enc

default valid := false
default limit := 30000
default threshold := 80

valid {
	count(violation) == 0
}

violation[msg] {
	warn_limit := round(limit * threshold * 0.01)
    secrets_cnt := input.data.monitoring.secrets.scalar
	input.data.gke.database_encryption.state == 1
    secrets_cnt> warn_limit
	msg := sprintf("Total number of secrets with encryption %d has reached warning level %d (limit is %d)", [secrets_cnt, warn_limit, limit])
}
