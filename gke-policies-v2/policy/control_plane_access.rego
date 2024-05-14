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
# title: Limit Control Plane endpoint access
# description: Control Plane endpoint access should be limited to authorized networks only
# custom:
#   group: Security
#   severity: Critical
#   recommendation: >
#     Navigate to the GKE page in Google Cloud Console and select the name of the cluster.
#     Under Networking, in the row for "Control plane authorized networks", click the edit icon.
#     Select the "Enable control plane authorized networks" checkbox. Click "Add Authorized network" and fill name and network fields.
#     Click Done. Add additional aiuthorized networks as needed.
#     Click "Save changes" once done.
#   externalURI: https://cloud.google.com/kubernetes-engine/docs/how-to/authorized-networks
#   sccCategory: CONTROL_PLANE_ACCESS_UNRESTRICTED
#   cis:
#     version: "1.4"
#     id: "5.6.3"
#   dataSource: gke
package gke.policy.control_plane_access

import future.keywords.if
import future.keywords.contains

default valid := false

valid if {
  count(violation) == 0
}

violation contains msg if {
  not input.data.gke.master_authorized_networks_config.enabled
  msg := "Cluster is not configured with master authorized networks"
}

violation contains msg if {
  not input.data.gke.master_authorized_networks_config.cidr_blocks
  msg := "Cluster is not configured with master authorized networks CIDRs"
}

violation contains msg if {
  count(input.data.gke.master_authorized_networks_config.cidr_blocks) < 1
  msg := "Cluster is not configured with master authorized networks CIDRs"
}
