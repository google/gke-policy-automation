# Copyright 2023 Google LLC
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
# title: Enable Customer-Managed Encryption Keys for persistent disks
# description: >-
#   Use Customer-Managed Encryption Keys (CMEK) to encrypt node boot and
#   dynamically-provisioned attached Google Compute Engine Persistent Disks (PDs) using
#   keys managed within Cloud Key Management Service (Cloud KMS).
# custom:
#   group: Security
#   severity: Medium
#   recommendation: >
#     CMEK cannot be enabled by updating an existing cluster. You must either recreate the desired
#     node pool or create a new cluster.
#   externalURI: https://cloud.google.com/kubernetes-engine/docs/how-to/using-cmek
#   sccCategory: PERSISTENT_DISK_CMEK_DISABLED
#   cis:
#     version: "1.4"
#     id: "5.9.1"
#   dataSource: gke
package gke.policy.node_pool_cmek

import future.keywords.if
import future.keywords.in
import future.keywords.contains

default valid := false

valid if {
  count(violation) == 0
}

violation contains msg if {
  some pool in input.node_pools
  not pool.config.boot_disk_kms_key
  msg := sprintf("Node pool %q is not configured with CMEK for the boot disk", [pool.name])
}
