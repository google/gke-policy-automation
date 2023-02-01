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
# title: Number of PODs per node
# description: The total number of PODs running on a single node
# custom:
#   group: Scalability
#   severity: Low
#   sccCategory: PODS_PER_NODE_LIMIT
#   dataSource: monitoring, gke

package gke.scalability.pods_per_node

default valid = false
default threshold = 80

valid {
	count(violation) == 0
}

violation[msg] {
	some nodepool, node
	pods_cnt := input.data.monitoring.pods_per_node.vector[nodepool][node]
	pooldata := [object | object := input.data.gke.node_pools[_]; object.name == nodepool]
	limit := pooldata[0].max_pods_constraint.max_pods_per_node
    warn_limit := round(limit * threshold * 0.01)
	pods_cnt > warn_limit
	msg := sprintf("Node %s is running %d PODs and reached warning level of %d (limit is %d)", [node, pods_cnt, warn_limit, limit])
}
