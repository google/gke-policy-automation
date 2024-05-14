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
# title: Use GKE Workload Identity
# description: GKE cluster should have Workload Identity enabled
# custom:
#   group: Security
#   severity: CRITICAL
#   recommendation: >
#     Navigate to the GKE page in Google Cloud Console and select the name of the cluster.
#     Under Security, in the row for "Workload Identity", click the edit icon.
#     Select the "Enable Workload Identity" checkbox. Leave the default workload identity pool
#     unchanged, as the default one is the only one that is currently supported.
#     Click "Save changes" once done.
#   externalURI: https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity
#   sccCategory: WORKLOAD_IDENTITY_DISABLED
#   cis:
#     version: "1.4"
#     id: "5.2.2"
#   dataSource: gke
package gke.policy.workload_identity

import future.keywords.if
import future.keywords.contains

default valid := false

valid if {
	count(violation) == 0
}

violation contains msg if {
	not input.workload_identity_config.workload_pool
	msg := "Cluster is not configured with Workload Identity"
}
