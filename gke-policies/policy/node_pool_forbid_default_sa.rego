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
# title: Forbid default compute SA on node_pool
# description: GKE node pools should have a dedicated sa with a restricted set of permissions
# custom:
#   group: Security
#   severity: Critical
#   sccCategory: DEFAULT_SA_CONFIGURED
#   cis:
#     version: "1.2"
#     id: "5.2.1"

package gke.policy.node_pool_forbid_default_sa

default valid = false

valid {
	count(violation) == 0
}

violation[msg] {
	input.node_pools[pool].config.service_account == "default"
	msg := sprintf("GKE cluster node_pool %q should have a dedicated SA", [input.node_pools[pool].name])
}
