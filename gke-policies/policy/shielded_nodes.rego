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
# title: GKE Shielded Nodes
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
#     version: "1.2"
#     id: "5.5.5"
#   dataSource: gke

package gke.policy.shielded_nodes

default valid = false

valid {
	count(violation) == 0
}

violation[msg] {
	not input.data.gke.shielded_nodes.enabled = true

	msg := "The GKE cluster does not have shielded nodes enabled"
}
