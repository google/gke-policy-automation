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
# title: Secure boot on the nodes
# description: Secure Boot helps ensure that the system only runs authentic software by verifying the digital signature of all boot components, and halting the boot process if signature verification fails
# custom:
#   group: Security
#   severity: Medium
#   recommendation: >
#     Once the node pool is provisioned, it cannot be updated to enable Secure Boot.
#     It is required to create new node pool in the cluster with Secure Boot feature
#     enabled.
#   externalURI: https://cloud.google.com/kubernetes-engine/docs/how-to/shielded-gke-nodes#secure_boot
#   sccCategory: SECURE_BOOT_DISABLED
#   cis:
#     version: "1.2"
#     id: "5.5.7"

package gke.policy.node_pool_secure_boot

default valid = false

valid {
  count(violation) == 0
}

violation[msg] {  
  not input.node_pools[pool].config.shielded_instance_config.enable_secure_boot
  msg := sprintf("Node pool %q has disabled secure boot.", [input.node_pools[pool].name])
} 
