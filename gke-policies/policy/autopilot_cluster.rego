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
# title: Use GKE Autopilot mode
# description: GKE Autopilot mode is the recommended way to operate a GKE cluster
# custom:
#   group: Management
#   severity: Medium
#   recommendation: >
#     Autopilot mode (recommended): GKE manages the underlying infrastructure such as node configuration,
#     autoscaling, auto-upgrades, baseline security configurations, and baseline networking configuration.
#   externalURI: https://cloud.google.com/kubernetes-engine/docs/concepts/choose-cluster-mode
#   sccCategory: AUTOPILOT_DISABLED
#   dataSource: gke
package gke.policy.autopilot

import future.keywords.if
import future.keywords.contains

default valid := false

valid if {
	count(violation) == 0
}

violation contains msg if {
	not input.autopilot.enabled
	msg := "Cluster is not using Autopilot mode"
}
