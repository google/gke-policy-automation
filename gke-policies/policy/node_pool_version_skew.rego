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
# title: Version skew between node pools and control plane
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

default valid = false

expr := `^([0-9]+)\.([0-9]+)\.([0-9]+)(-.+)*$`

valid {
  count(violation) == 0
}

violation[msg] {
  not input.data.gke.current_master_version
  msg := "control plane version is undefined"
}

violation[msg] {
  some node_pool
  not input.data.gke.node_pools[node_pool].version
  msg := sprintf("node pool %q control plane version is undefined", [input.data.gke.node_pools[node_pool].name])
}

violation[msg] {
  master_ver := regex.find_all_string_submatch_n(expr, input.data.gke.current_master_version, 1)
  count(master_ver) == 0
  msg := sprintf("control plane version %q does not match version regex", [input.data.gke.current_master_version])
}

violation[msg] {
  some node_pool
  node_pool_ver := regex.find_all_string_submatch_n(expr, input.data.gke.node_pools[node_pool].version, 1)
  count(node_pool_ver) == 0
  msg := sprintf("node pool %q version %q does not match version regex", [input.data.gke.node_pools[node_pool].name, input.data.gke.node_pools[node_pool].version])
}

violation[msg] {
  master_ver := regex.find_all_string_submatch_n(expr, input.data.gke.current_master_version, 1)
  some node_pool
  node_pool_ver := regex.find_all_string_submatch_n(expr, input.data.gke.node_pools[node_pool].version, 1)
  master_ver[0][1] != node_pool_ver[0][1]
  msg := sprintf("node pool %q and control plane major versions differ", [input.data.gke.node_pools[node_pool].name])
}

violation[msg] {
  master_ver := regex.find_all_string_submatch_n(expr, input.data.gke.current_master_version, 1)
  some node_pool
  node_pool_ver := regex.find_all_string_submatch_n(expr, input.data.gke.node_pools[node_pool].version, 1)
  minor_diff := to_number(master_ver[0][2]) - to_number(node_pool_ver[0][2])
  abs(minor_diff) > 2
  msg := sprintf("node pool %q and control plane minor versions difference is greater than 2", [input.data.gke.node_pools[node_pool].name])
}
