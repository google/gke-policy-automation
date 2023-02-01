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
#   sccCategory: NODES_LIMIT
#   dataSource: monitoring, gke

package gke.scalability.nodes

default valid = false

default private_nodes_limit = 15000
default public_nodes_limit = 5000

valid {
	count(violation) == 0
}

violation[msg] {
	nodes := input.data.monitoring.nodes.scalar 
	is_private := input.data.gke.private_cluster_config.enable_private_nodes
	is_private = true 
	nodes > private_nodes_limit
	msg := sprintf("nodes found: %d higher than the limit for private clusters: %d", [nodes, private_nodes_limit])
	print(msg)
}

violation[msg] {
	nodes := input.data.monitoring.nodes.scalar 
	is_private := input.data.gke.private_cluster_config.enable_private_nodes
	is_private = false 
	nodes > public_nodes_limit
	msg := sprintf("nodes found: %d higher than the limit for non private clusters: %d", [nodes, public_nodes_limit])
	print(msg)
}
