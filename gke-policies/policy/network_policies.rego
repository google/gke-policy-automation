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
# title: Enable Kubernetes Network Policies
# description: GKE cluster should have Network Policies or Dataplane V2 enabled
# custom:
#   group: Security
#   severity: High
#   recommendation: >
#     Navigate to the GKE page in Google Cloud Console and select the name of the cluster.
#     Under Networking, in the row for "Network policy", click the edit icon.
#     Select the "Enable Network Policy for nodes" checkbox and click "Save changes" button.
#   externalURI: https://cloud.google.com/kubernetes-engine/docs/tutorials/network-policy
#   sccCategory: NETWORK_POLICIES_DISABLED
#   cis:
#     version: "1.4"
#     id: "5.6.7"
#   dataSource: gke
package gke.policy.network_policies_engine

import future.keywords.if
import future.keywords.contains

default valid := false

valid if {
	count(violation) == 0
}

violation contains msg if {
	input.addons_config.network_policy_config.disabled
	not input.network_policy
	not input.network_config.datapath_provider == 2
	msg := "Cluster is not configured with Kubneretes Network Policies"
}

violation contains msg if {
	count(input.addons_config.network_policy_config) == 0
	not input.network_policy.enabled
	not input.network_config.datapath_provider == 2
	msg := "Cluster is configured with Kubneretes Network Policies without configuration"
}

violation contains msg if {
	input.addons_config.network_policy_config.disabled
	count(input.network_policy) == 0
	not input.network_config.datapath_provider == 2
	msg := "Cluster is not DPv2 and has not configured Kubneretes Network Policies"
}
