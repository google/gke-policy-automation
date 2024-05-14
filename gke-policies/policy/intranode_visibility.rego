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
# title: Enable GKE intranode visibility
# description: GKE cluster should have intranode visibility enabled
# custom:
#   group: Security
#   severity: High
#   recommendation: >
#     Navigate to the GKE page in Google Cloud Console and select the name of the cluster.
#     Under Cluster, click Networking.
#     Select the Enable intranode visibility checkbox and click "Create".
#   externalURI: https://cloud.google.com/kubernetes-engine/docs/how-to/intranode-visibility
#   sccCategory: INTRANODE_VISIBILITY_DISABLED
#   cis:
#     version: "1.4"
#     id: "5.6.1"
#   dataSource: gke
package gke.policy.networkConfig

import future.keywords.if
import future.keywords.contains

default valid := false

valid if {
	count(violation) == 0
}

violation contains msg if {
	not input.networkConfig.enableIntraNodeVisibility = true
	msg := "Cluster is not configured with Intranode Visibility"
}
