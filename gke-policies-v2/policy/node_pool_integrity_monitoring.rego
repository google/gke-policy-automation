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
# title: Enable integrity monitoring for node pools
# description: GKE node pools should have integrity monitoring feature enabled to detect changes in a VM boot measurements
# custom:
#   group: Security
#   severity: Critical
#   recommendation: >
#    Once the node pool is provisioned, it cannot be updated to enable Integrity Monitoring.
#    It is required to create new node pool in the cluster with Integrity Monitoring feature
#    enabled.
#   externalURI: https://cloud.google.com/kubernetes-engine/docs/how-to/shielded-gke-nodes#node_integrity
#   sccCategory: NODEPOOL_INTEGRITY_MONITORING_DISABLED
#   cis:
#     version: "1.4"
#     id: "5.5.6"
#   dataSource: gke
package gke.policy.node_pool_integrity_monitoring

import future.keywords.if
import future.keywords.in
import future.keywords.contains

default valid := false

valid if {
  count(violation) == 0
}

violation contains msg if {
  some pool in input.data.gke.node_pools
  not pool.config.shielded_instance_config.enable_integrity_monitoring
  msg := sprintf("Node pool %q is not configured with integrity monitoring", [pool.name])
}
