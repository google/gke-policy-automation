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
# title: Ensure acceptable version skew in a cluster
# description: Difference between cluster control plane version and node pools version should be no more than 2 minor versions.
# custom:
#   group: Management
#   severity: Critical
#   recommendation: >
#     Navigate to the GKE page in Google Cloud Console and select the name of the cluster.
#     Select "Nodes" tab and click on the name of the target node pool. Within the node pool
#     details pane, click EDIT. Under the "Node version heading, click "Change" button.
#     Select the desired node version from the list. The difference between target nodes version
#     and current control plane version should be no more than 2 minor versions.
#     Click "Upgrade" button once done.
#   externalURI: https://cloud.google.com/kubernetes-engine/docs/how-to/upgrading-a-cluster#upgrading-nodes
#   sccCategory: NODEPOOL_VERSION_SKEW_UNSUPPORTED
#   dataSource: gke
package gke.policy.node_pool_version_skew

import future.keywords.if
import future.keywords.in
import future.keywords.contains

default valid := false

expr := `^([0-9]+)\.([0-9]+)\.([0-9]+)(-.+)*$`

valid if {
  count(violation) == 0
}

violation contains msg if {
  not input.current_master_version
  msg := "control plane version is undefined"
}

violation contains msg if {
  some pool in input.node_pools
  not pool.version
  msg := sprintf("Node pool %q version is undefined", [pool.name])
}

violation contains msg if {
  master_ver := regex.find_all_string_submatch_n(expr, input.current_master_version, 1)
  count(master_ver) == 0
  msg := sprintf("Control plane version %q does not match version regex", [input.current_master_version])
}

violation contains msg if {
  some pool in input.node_pools
  node_pool_ver := regex.find_all_string_submatch_n(expr, pool.version, 1)
  count(node_pool_ver) == 0
  msg := sprintf("Node pool %q version %q does not match version regex", [pool.name, pool.version])
}

violation contains msg if {
  master_ver := regex.find_all_string_submatch_n(expr, input.current_master_version, 1)
  some pool in input.node_pools
  node_pool_ver := regex.find_all_string_submatch_n(expr, pool.version, 1)
  master_ver[0][1] != node_pool_ver[0][1]
  msg := sprintf("Node pool %q and control plane major versions differ", [pool.name])
}

violation contains msg if {
  master_ver := regex.find_all_string_submatch_n(expr, input.current_master_version, 1)
  some pool in input.node_pools
  node_pool_ver := regex.find_all_string_submatch_n(expr, pool.version, 1)
  minor_diff := to_number(master_ver[0][2]) - to_number(node_pool_ver[0][2])
  abs(minor_diff) > 2
  msg := sprintf("Node pool %q and control plane minor versions difference is greater than 2", [pool.name])
}
