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
# title: GKE Nodes Limit
# description: GKE Nodes Limit
# custom:
#   group: Scalability
#   severity: High
#   recommendation: >
#     For GKE Standard clusters, adjust Compute Engine machine type used for nodes to
#     accomodate more PODs on a single node. Increase number of PODs per node setting on a nodepools when posssible.
#     Note that number of PODs per node should be aligned with the size of the POD's IP range.
#     When the above is not applicable, concider running your workloads on additional cluster(s).
#   externalURI: https://cloud.google.com/kubernetes-engine/quotas
#   sccCategory: NODES_LIMIT
#   dataSource: monitoring, gke

package gke.scalability.nodes

default valid = false

default private_nodes_limit = 15000
default public_nodes_limit = 5000
default autopilot_nodes_limit = 1000
default threshold = 80

valid {
	count(violation) == 0
}

violation[msg] {
	warn_limit = round(private_nodes_limit * threshold * 0.01)
	nodes := input.data.monitoring.nodes.scalar 
	is_private := input.data.gke.private_cluster_config.enable_private_nodes
	is_private = true 
	nodes > warn_limit
	msg := sprintf("nodes found: %d higher than the limit for private clusters: %d", [nodes, warn_limit])
}

violation[msg] {
	warn_limit = round(public_nodes_limit * threshold * 0.01)
	nodes := input.data.monitoring.nodes.scalar 
	is_private := input.data.gke.private_cluster_config.enable_private_nodes
	is_private = false 
	nodes > warn_limit
	msg := sprintf("nodes found: %d higher than the limit for non private clusters: %d", [nodes, warn_limit])
}

violation[msg] {
	warn_limit = round(autopilot_nodes_limit * threshold * 0.01)
	nodes := input.data.monitoring.nodes.scalar 
	input.data.gke.autopilot.enabled
	nodes > warn_limit
	msg := sprintf("nodes found: %d higher than the warn limit for autopilot clusters: %d", [nodes, warn_limit])
}
