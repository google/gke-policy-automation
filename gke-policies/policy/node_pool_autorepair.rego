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
# title: Use Node Auto-Repair
# description: GKE node pools should have Node Auto-Repair enabled to configure Kubernetes Engine
# custom:
#   group: Availability
#   severity: High
#   sccCategory: NODEPOOL_AUTOREPAIR_DISABLED

package gke.policy.node_pool_autorepair

default valid = false

valid {
  count(violation) == 0
}

violation[msg] {  
  not input.node_pools[pool].management.auto_repair
  msg := sprintf("autorepair not set for GKE node pool %q", [input.node_pools[pool].name])
} 


