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
# title: Enrollment in Release Channels
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
#     version: "1.2"
#     id: "5.5.4"

package gke.policy.cluster_release_channels

default valid = false

valid {
  count(violation) == 0
}

violation[msg] {
  not input.release_channel.channel  
  msg := "GKE cluster is not enrolled in release channel"
}
