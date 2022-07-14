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
# title: Integrity monitoring on the nodes
# description: GKE node pools should have integrity monitoring feature enabled to detect changes in a VM boot measurments
# custom:
#   group: Security
#   severity: Critical
#   sccCategory: NODEPOOL_INTEGRITY_MONITORING_DISABLED

package gke.policy.node_pool_integrity_monitoring

default valid = false

valid {
  count(violation) == 0
}

violation[msg] {  
  not input.node_pools[pool].config.shielded_instance_config.enable_integrity_monitoring
  msg := sprintf("Node pool %q has disabled integrity monitoring feature.", [input.node_pools[pool].name])
} 
