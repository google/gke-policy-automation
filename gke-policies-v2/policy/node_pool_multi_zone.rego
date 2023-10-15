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
# title: Multi-zone node pools
# description: GKE node pools should be regional (multiple zones) for maximum nodes availability during zonal outages
# custom:
#   group: Availability
#   severity: High
#   recommendation: >
#     Navigate to the GKE page in Google Cloud Console and select the name of the cluster.
#     Select "Nodes" tab and click on the name of the target node pool. Within the node pool
#     details pane, click EDIT. Under the "Zones" heading, select the cheboxes for multiple
#     zones. Slick "Save" button once done.
#   externalURI: https://cloud.google.com/kubernetes-engine/docs/concepts/node-pools#multiple-zones
#   sccCategory: NODEPOOL_ZONAL
#   dataSource: gke

package gke.policy.node_pool_multi_zone

default valid := false

valid {
  count(violation) == 0
}

violation[msg] {
  some pool
  count(input.data.gke.node_pools[pool].locations) < 2
  msg := sprintf("Node pool %q is not on multiple zones.", [input.data.gke.node_pools[pool].name])
}