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
# title: Enable Google Groups for RBAC
# description: GKE cluster should have RBAC security Google group enabled
# custom:
#   group: Security
#   severity: Medium
#   recommendation: >
#     This recommendation requires Google Group with a name gke-security-groups to be
#     created in your domain as a prerequsite.
#     Navigate to the GKE page in Google Cloud Console and select the name of the cluster.
#     Under Security, in the row for "Google Groups for RBAC", click the edit icon.
#     Select the "Enable Google Groups for RBAC" checkbox. Fill the name of your group
#     in the text field below. Click "Save changes" once done.
#   externalURI: https://cloud.google.com/kubernetes-engine/docs/how-to/google-groups-rbac
#   sccCategory: RBAC_SECURITY_GROUP_DISABLED
#   cis:
#     version: "1.4"
#     id: "5.8.3"
#   dataSource: gke
package gke.policy.rbac_security_group_enabled

import future.keywords.if
import future.keywords.contains

default valid := false

valid if {
  count(violation) == 0
}

violation contains msg if {
  not input.data.gke.authenticator_groups_config.enabled
  msg := "Cluster is not configured with Google Groups for RBAC"
}
