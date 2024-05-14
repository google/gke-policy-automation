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
# title: Number of nodes in a nodepool zone
# description: The total number of nodes running in a single node pool zone
# custom:
#   group: Scalability
#   severity: Low
#   recommendation: >
#     The limit applies when container-native load balancing with NEGs is not used with GKE Ingress controller.
#     We recommend using container-native load balancing with NEGs (Network Endpoint Groups) to improve the
#     load balancing performance and avoid this limit.
#   externalURI: https://cloud.google.com/kubernetes-engine/quotas
#   sccCategory: NODES_PER_POOL_ZONE_LIMIT
#   dataSource: monitoring, gke
package gke.scalability.nodes_per_pool_zone

import future.keywords.if
import future.keywords.contains

default valid := false
default limit := 1000
default threshold := 80

valid if {
	count(violation) == 0
}

violation contains msg if {
	warn_limit := round(limit * threshold * 0.01)
	some nodepool, zone
	not input.data.gke.autopilot.enabled
    nodes_cnt := input.data.monitoring.nodes_per_pool_zone.vector[nodepool][zone]
	nodes_cnt > warn_limit
	msg := sprintf("Total number of nodes %d in a nodepool %s in a zone %s has reached warning level %d (limit is %d)", [nodes_cnt, nodepool, zone, warn_limit, limit])
}
