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
# title: Number of services in a cluster
# description: The total number of services running in a cluster
# custom:
#   group: Scalability
#   severity: Medium
#   recommendation: >
#     The performance of iptables used by kube-proxy degrades if there are too many services or
#     the number of backends behind a Service is high. We recommend keeping the number of services in the cluster
#     below the limit.
#   externalURI: https://cloud.google.com/kubernetes-engine/docs/concepts/planning-large-clusters#limits-best-practices-large-scale-clusters
#   sccCategory: SERVICES_LIMIT
#   dataSource: monitoring
package gke.scalability.services

import future.keywords.if
import future.keywords.contains

default valid := false
default limit := 10000
default threshold := 80

valid if {
	count(violation) == 0
}

violation contains msg if {
	warn_limit := round(limit * threshold * 0.01)
    input.data.monitoring.services.scalar > warn_limit
	msg := sprintf("Total number of services %d has reached warning level %d (limit is %d)", [input.data.monitoring.services.scalar, warn_limit, limit])
}
