# Copyright 2023 Google LLC
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
# title: Number of namespaces in a cluster
# description: The total number of namespaces in a cluster
# custom:
#   group: Scalability
#   severity: High
#   recommendation: >
#     Please concider distributing workloads among more than one cluster when the total number of required namespaces
#     is above the limit of a supported number of namespaces for a single cluster.
#   externalURI: https://github.com/kubernetes/community/blob/master/sig-scalability/configs-and-limits/thresholds.md
#   sccCategory: NAMESPACES_LIMIT
#   dataSource: monitoring

package gke.scalability.namespaces

default valid := false
default limit := 10000
default threshold := 80

valid {
	count(violation) == 0
}

violation[msg] {
	warn_limit := round(limit * threshold * 0.01)
    input.data.monitoring.namespaces.scalar > warn_limit
	msg := sprintf("Total number of namespaces %d has reached warning level %d (limit is %d)", [input.data.monitoring.namespaces.scalar, warn_limit, limit])
}
