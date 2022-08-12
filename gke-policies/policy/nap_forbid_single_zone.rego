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
# title: Ensure that node pool locations within Node Auto-Provisioning are covering one than more zone (or not enforced at all)
# description: Node Auto-Provisioning configuration should cover more than one zone
# custom:
#   group: Security
#   severity: Critical
#   sccCategory: NAP_MULTIPLE_ZONES_CONFIGURED
#   cis:
#     version: "1.2"
#     id: "5.2.1"

package gke.policy.nap_forbid_single_zone

default valid = false

valid {
	count(violation) == 0
}

violation[msg] {
	input.autoscaling.enable_node_autoprovisioning == true
	count(input.autoscaling.autoprovisioning_locations) == 1
	msg := "GKE cluster Node Auto-Provisioning configuration should cover more than one zone"
}
