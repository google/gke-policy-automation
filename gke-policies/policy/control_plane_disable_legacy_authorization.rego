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
# title: GKE RBAC authorization
# description: GKE cluster should use RBAC instead of legacy ABAC authorization
# custom:
#   group: Security
#   severity: Critical
#   recommendation: >
#     Navigate to the GKE page in Google Cloud Console and select the name of the cluster.
#     Under Security, in the row for "Legacy authorization", click the edit icon.
#     Unselect the "Enable legacy authorization" checkbox and click "Save changes".
#   externalURI: https://cloud.google.com/kubernetes-engine/docs/how-to/api-server-authentication#legacy-auth
#   sccCategory: RBAC_DISABLED
#   cis:
#     version: "1.2"
#     id: "5.8.4"
#   dataSource: gke

package gke.policy.disable_legacy_authorization

default valid = false

valid {
	count(violation) == 0
}

violation[msg] {
	input.Data.gke.legacy_abac.enabled
	msg := "The GKE cluster is configured to use legacy ABAC authorization mechanism"
}
