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
# title: Enable maintenance windows
# description: GKE cluster should use maintenance windows and exclusions to upgrade predictability and to align updates with off-peak business hours.
# custom:
#   group: Management
#   severity: Medium
#   recommendation: >
#     Navigate to the GKE page in Google Cloud Console and select the name of the cluster.
#     Under Automation, in the row for "Maintenance Window", click the edit icon.
#     Select the "Enable Maintenance Window" checkbox. Select the start time and length, then select the days of the week on which the maintenance window occurs.
#     To edit the recurrence rule specification (RRule) directly, select Custom editor.
#     Click "Save changes" once done.
#   externalURI: https://cloud.google.com/kubernetes-engine/docs/concepts/maintenance-windows-and-exclusions
#   sccCategory: MAINTENANCE_WINDOWS_DISABLED
#   dataSource: gke
package gke.policy.cluster_maintenance_window

import future.keywords.if
import future.keywords.contains

default valid := false

valid if {
  count(violation) == 0
}

violation contains msg if {
  not input.data.gke.maintenance_policy.window.Policy
  msg := "GKE cluster is not configured with maintenance window"
}
