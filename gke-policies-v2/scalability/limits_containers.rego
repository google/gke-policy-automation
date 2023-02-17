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
# title: Number of containers in a cluster
# description: The total number of containers running in a cluster
# custom:
#   group: Scalability
#   severity: High
#   recommendation: >
#     Please concider distributing workloads among more than one cluster when the total number of required containers
#     is above the limit of a supported number of containers for a single cluster.
#   externalURI: https://cloud.google.com/kubernetes-engine/quotas
#   sccCategory: CONTAINERS_LIMIT
#   dataSource: monitoring, gke

package gke.scalability.containers

default valid = false
default limit_standard = 400000
default limit_autopilot = 24000
default threshold = 80

valid {
	count(violation) == 0
}

violation[msg] {
	warn_limit = round(limit_standard * threshold * 0.01)
	not input.data.gke.autopilot.enabled
    input.data.monitoring.containers.scalar > warn_limit
	msg := sprintf("Total number of containers %d has reached warning level %d (limit is %d for standard clusters)", [input.data.monitoring.containers.scalar, warn_limit, limit_standard])
}

violation[msg] {
	warn_limit = round(limit_autopilot * threshold * 0.01)
	input.data.gke.autopilot.enabled
    input.data.monitoring.containers.scalar > warn_limit
	msg := sprintf("Total number of containers %d has reached warning level %d (limit is %d for autopilot clusters)", [input.data.monitoring.containers.scalar, warn_limit, limit_autopilot])
}
