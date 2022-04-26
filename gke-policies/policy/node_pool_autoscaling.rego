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
# title: Use node pool autoscaler
# description: GKE node pools should have autoscaler configured to proper resize nodes according to traffic
# custom:
#   group: Availability
package gke.policy.node_pool_autoscaling

default valid = false

valid {
  count(violation) == 0
}

violation[msg] {
  not input.node_pools[pool].autoscaling.enabled
  msg := sprintf("Node pool %q does not have autoscaling configured.", [input.node_pools[pool].name])
}