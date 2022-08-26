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
# title: Use RBAC Google group
# description: GKE cluster should have RBAC security Google group enabled
# custom:
#   group: Security
#   severity: Medium
#   sccCategory: RBAC_SECURITY_GROUP_DISABLED
#   cis:
#     version: "1.2"
#     id: "5.8.3"

package gke.policy.rbac_security_group_enabled

default valid = false

valid {
  count(violation) == 0
}

violation[msg] {  
  not input.authenticator_groups_config.enabled
  msg := sprintf("RBAC security group not enabled for cluster %q", [input.name])
}