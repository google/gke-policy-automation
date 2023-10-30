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
# title: Use Container-Optimized OS
# description: GKE node pools should use Container-Optimized OS which is maintained by Google and optimized for running Docker containers with security and efficiency.
# custom:
#   group: Security
#   severity: High
#   recommendation: >
#     Navigate to the GKE page in Google Cloud Console and select the name of the cluster.
#     Select "Nodes" tab and click on the name of the target node pool. Within the node pool
#     details pane, click EDIT. Under the "Image type" heading, click "Change" button.
#     Select the Container-Optimized OS with containerd image type from the list.
#     Click "Change" button once done.
#   externalURI: https://cloud.google.com/kubernetes-engine/docs/concepts/node-images
#   sccCategory: NODEPOOL_COS_UNUSED
#   cis:
#     version: "1.2"
#     id: "5.5.1"
#   dataSource: gke

package gke.policy.node_pool_use_cos

import future.keywords.in

default valid := false

valid {
  count(violation) == 0
}

violation[msg] {
  some pool
  not lower(input.data.gke.node_pools[pool].config.image_type) in {"cos", "cos_containerd"}
  not startswith(lower(input.data.gke.node_pools[pool].config.image_type), "windows")
  msg := sprintf("Node pool %q does not use Container-Optimized OS.", [input.data.gke.node_pools[pool].name])
}
