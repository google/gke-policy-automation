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
# title: Change default Service Accounts in node pools
# description: GKE node pools should have a dedicated sa with a restricted set of permissions
# custom:
#   group: Security
#   severity: Critical
#   recommendation: >
#     Navigate to the GKE page in Google Cloud Console and select the name of the cluster.
#     Select "Nodes" tab and click on the name of the target node pool. Within the node pool
#     details pane, click EDIT. Under the "Management" heading, select the "Enable auto-upagde"
#     checkbox. Click "Save" button once done.
#   externalURI: https://cloud.google.com/kubernetes-engine/docs/how-to/hardening-your-cluster#use_least_privilege_sa
#   sccCategory: DEFAULT_SA_CONFIGURED
#   cis:
#     version: "1.4"
#     id: "5.2.1"
#   dataSource: gke
package gke.policy.node_pool_forbid_default_sa

import future.keywords.if
import future.keywords.in
import future.keywords.contains

default valid := false

valid if {
	count(violation) == 0
}

violation contains msg if {
	not input.data.gke.autopilot.enabled
	some pool in input.data.gke.node_pools
	pool.config.service_account == "default"
	msg := sprintf("Node pool %q is configured with default SA", [pool.name])
}
