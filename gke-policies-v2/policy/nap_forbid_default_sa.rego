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
# title: Change default Service Accounts in Node Auto-Provisioning
# description: Node Auto-Provisioning configuration should not allow default Service Accounts
# custom:
#   group: Security
#   severity: Critical
#   recommendation: >
#     Navigate to the GKE page in Google Cloud Console and select the name of the cluster.
#     Under Automation, in the row for "Node auto-provisioning", click the edit icon.
#     Expand the "Service account" drop-down list and select dedicated, non-default service
#     account. Click "Save changes" once done.
#   externalURI: https://cloud.google.com/kubernetes-engine/docs/how-to/node-auto-provisioning#identity
#   sccCategory: NAP_DEFAULT_SA_CONFIGURED
#   cis:
#     version: "1.4"
#     id: "5.2.1"
#   dataSource: gke
package gke.policy.nap_forbid_default_sa

import future.keywords.if
import future.keywords.contains

default valid := false

valid if {
	count(violation) == 0
}

violation contains msg if {
	not input.data.gke.autopilot.enabled
	input.data.gke.autoscaling.enable_node_autoprovisioning == true
	input.data.gke.autoscaling.autoprovisioning_node_pool_defaults.service_account == "default"
	msg := "Cluster is configured with default service account for Node Auto-Provisioning"
}
