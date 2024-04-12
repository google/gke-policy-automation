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
# title: Enable node pool auto-scaling
# description: GKE node pools should have autoscaling configured to proper resize nodes according to traffic
# custom:
#   group: Scalability
#   severity: Medium
#   recommendation: >
#     Navigate to the GKE page in Google Cloud Console and select the name of the cluster.
#     Select "Nodes" tab and click on the name of the target node pool. Within the node pool
#     details pane, click EDIT. Under the "Size" heading, select the "Enable cluster autoscaler"
#     checkbox. Adjust the values in the "Minimum number of nodes" and "Maximum number of nodes"
#     fields. Slick "Save" button once done.
#   externalURI: https://cloud.google.com/kubernetes-engine/docs/concepts/cluster-autoscaler
#   sccCategory: NODEPOOL_AUTOSCALING_DISABLED
#   dataSource: gke

package gke.policy.node_pool_autoscaling

default valid := false

valid {
  count(violation) == 0
}

violation[msg] {
  some pool
  not input.data.gke.node_pools[pool].autoscaling.enabled
  msg := sprintf("Node pool %q is not configured with autoscaling", [input.data.gke.node_pools[pool].name])
}
