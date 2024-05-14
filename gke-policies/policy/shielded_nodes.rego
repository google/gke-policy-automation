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
# title: Enable Shielded Nodes
# description: GKE cluster should use shielded nodes
# custom:
#   group: Security
#   severity: High
#   recommendation: >
#     Navigate to the GKE page in Google Cloud Console and select the name of the cluster.
#     Under Security, in the row for "Shielded GKE nodes", click the edit icon.
#     Select the "Enable Shielded GKE Nodes" checkbox and click "Save changes".
#   externalURI: https://cloud.google.com/kubernetes-engine/docs/how-to/shielded-gke-nodes
#   sccCategory: SHIELDED_NODES_DISABLED
#   cis:
#     version: "1.4"
#     id: "5.5.5"
#   dataSource: gke
package gke.policy.shielded_nodes

import future.keywords.if
import future.keywords.contains

default valid := false

valid if {
	count(violation) == 0
}

violation contains msg if {
	not input.shielded_nodes.enabled = true
	msg := "Cluster is not configured with shielded nodes"
}
