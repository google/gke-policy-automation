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
# title: GKE Workload Identity
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
#     version: "1.2"
#     id: "5.2.2"
#   dataSource: gke

package gke.policy.workload_identity

default valid = false

valid {
	count(violation) == 0
}

violation[msg] {
	not input.Data.gke.workload_identity_config.workload_pool
	msg := "The GKE cluster does not have workload identity enabled"
}
