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
# title: Ensure redundancy of Node Auto-provisioning node pools
# description: Node Auto-Provisioning configuration should cover more than one zone
# custom:
#   group: Security
#   severity: High
#   recommendation: >
#     Navigate to the GKE page in Google Cloud Console and select the name of the cluster.
#     Under Automation, in the row for "Node auto-provisioning", click the edit icon.
#     Under the "Node pool location", select multiple zone checkboxes. Click "Save changes" once done.
#   externalURI: https://cloud.google.com/kubernetes-engine/docs/how-to/node-auto-provisioning#auto-provisioning_locations
#   sccCategory: NAP_ZONAL
#   dataSource: gke
package gke.policy.nap_forbid_single_zone

import future.keywords.if
import future.keywords.contains

default valid := false

valid if {
	count(violation) == 0
}

violation contains msg if {
	input.autoscaling.enable_node_autoprovisioning == true
	count(input.autoscaling.autoprovisioning_locations) == 1
	msg := "Cluster is not configured with multiple zones for NAP node pools"
}
