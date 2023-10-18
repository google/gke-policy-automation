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
# title: Number of services per namespace
# description: The total number of services running in single namespace
# custom:
#   group: Scalability
#   severity: Medium
#   recommendation: >
#     The number of environment variables generated for Services might outgrow shell limits.
#     This might cause Pods to crash on startup. We recommend keeping the number of services per namespace
#     below the limit.
#   externalURI: https://cloud.google.com/kubernetes-engine/docs/concepts/planning-large-clusters#limits-best-practices-large-scale-clusters
#   sccCategory: SERVICES_PER_NS_LIMIT
#   dataSource: monitoring

package gke.scalability.services_per_ns

default valid := false
default limit := 5000
default threshold := 80

valid {
	count(violation) == 0
}

violation[msg] {
	warn_limit := round(limit * threshold * 0.01)
	some namespace
	srv_cnt := input.data.monitoring.services_per_ns.vector[namespace]
    srv_cnt > warn_limit
	msg := sprintf("Total number of services %d in a namespace %s has reached warning level %d (limit is %d)", [srv_cnt, namespace, warn_limit, limit])
}
