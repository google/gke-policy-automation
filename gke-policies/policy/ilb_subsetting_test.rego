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
# title: GKE L4 ILB Subsetting
# description: GKE cluster should use GKE L4 ILB Subsetting is nodes > 250
# custom:
#   group: Scalability
package gke.policy.enable_ilb_subsetting

test_enabled_ilb_subsetting_high_nodes {
	valid with input as {"name": "test-cluster", "current_node_count": 251, "network_config": { "enable_l4ilb_subsetting": true }}
}

test_disabled_ilb_subsetting_low_nodes {
	valid with input as {"name": "test-cluster", "current_node_count": 3, "network_config": {}}
}

test_disabled_ilb_subsetting_high_nodes {
	not valid with input as {"name": "test-cluster", "current_node_count": 251, "network_config": {}}
}
