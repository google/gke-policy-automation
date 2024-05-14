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
# title: Enroll cluster in Release Channels
# description: GKE cluster should be enrolled in release channels
# custom:
#   group: Security
#   severity: High
#   sccCategory: RELEASE_CHANNELS_DISABLED
#   recommendation: >
#     Navigate to the GKE page in Google Cloud Console and select the name of the cluster.
#     Under Cluster Basics, in the row for "Release channel	", click the edit icon.
#     Select the "Release channel" option. From the drop-down lists, select the desired release channel and version.
#     Click "Save changes" once done.
#   externalURI: https://cloud.google.com/kubernetes-engine/docs/concepts/release-channels
#   cis:
#     version: "1.4"
#     id: "5.5.4"
#   dataSource: gke
package gke.policy.cluster_release_channels

import future.keywords.if
import future.keywords.contains

default valid := false

valid if {
  count(violation) == 0
}

violation contains msg if {
  not input.release_channel.channel  
  msg := "Cluster is not enrolled in any release channel"
}
