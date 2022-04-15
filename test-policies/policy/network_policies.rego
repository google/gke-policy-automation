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
# title: GKE Network Policies engine
# description: GKE cluster should have Network Policies or Dataplane V2 enabled
# custom:
#   group: Security
package gke.policy.network_policies_engine

default valid = false

valid {
   count(violation) == 0
}


dataplane_v1_netpol_config {
  not input.addons_config.network_policy_config.disabled
}

dataplane_v1_netpol_enabled {
  not input.network_policy.enabled
}

dataplane_v1 {
  dataplane_v1_netpol_config
  dataplane_v1_netpol_enabled
}

dataplane_v2 {
  not input.addons_config.network_policy_config.disabled
  not input.network_config.datapath_provider == 2
}

violation[msg] {
  dataplane_v1
  dataplane_v2
  msg := "GKE cluster has no active Network Policies engines"
}
