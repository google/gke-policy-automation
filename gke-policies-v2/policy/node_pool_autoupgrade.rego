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
# title: Enable node auto-upgrade
# description: GKE node pools should have Node Auto-Upgrade enabled to configure Kubernetes Engine
# custom:
#   group: Security
#   severity: High
#   recommendation: >
#     Navigate to the GKE page in Google Cloud Console and select the name of the cluster.
#     Select "Nodes" tab and click on the name of the target node pool. Within the node pool
#     details pane, click EDIT. Under the "Management" heading, select the "Enable auto-upagde"
#     checkbox. Slick "Save" button once done.
#   externalURI: https://cloud.google.com/kubernetes-engine/docs/how-to/node-auto-upgrades
#   sccCategory: NODEPOOL_AUTOUPGRADE_DISABLED
#   cis:
#     version: "1.4"
#     id: "5.5.3"
#   dataSource: gke

package gke.policy.node_pool_autoupgrade

default valid := false

valid {
  count(violation) == 0
}

violation[msg] {
  some pool
  not input.data.gke.node_pools[pool].management.auto_upgrade
  msg := sprintf("Node pool %q is not configured with auto-upgrade", [input.data.gke.node_pools[pool].name])
}
