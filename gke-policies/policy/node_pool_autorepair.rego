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
# title: Enable node auto-repair
# description: GKE node pools should have Node Auto-Repair enabled to configure Kubernetes Engine
# custom:
#   group: Availability
#   severity: High
#   recommendation: >
#     Navigate to the GKE page in Google Cloud Console and select the name of the cluster.
#     Select "Nodes" tab and click on the name of the target node pool. Within the node pool
#     details pane, click EDIT. Under the "Management" heading, select the "Enable auto-repair"
#     checkbox. Slick "Save" button once done.
#   externalURI: https://cloud.google.com/kubernetes-engine/docs/how-to/node-auto-repair
#   sccCategory: NODEPOOL_AUTOREPAIR_DISABLED
#   cis:
#     version: "1.4"
#     id: "5.5.2"
#   dataSource: gke
package gke.policy.node_pool_autorepair

import future.keywords.if
import future.keywords.in
import future.keywords.contains

default valid := false

valid if {
  count(violation) == 0
}

violation contains msg if {  
  some pool in input.node_pools
  not pool.management.auto_repair
  msg := sprintf("Node pool %q is not configured with auto-repair", [pool.name])
} 
