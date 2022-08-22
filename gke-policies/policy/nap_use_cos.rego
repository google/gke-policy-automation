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
# title: Ensure that nodes in Node Auto-Provisioning node pools will use Container-Optimized OS
# description: Nodes in Node Auto-Provisioning should use Container-Optimized OS
# custom:
#   group: Security
#   severity: High
#   sccCategory: NAP_COS_UNCONFIGURED
#   cis:
#     version: "1.2"
#     id: "5.2.1"

package gke.policy.nap_use_cos

import future.keywords.in

default valid = false

valid {
	count(violation) == 0
}

violation[msg] {
	input.autoscaling.enable_node_autoprovisioning == true
	not lower(input.autoscaling.autoprovisioning_node_pool_defaults.image_type) in { "cos", "cos_containerd"}
	
	msg := "GKE cluster Node Auto-Provisioning configuration use Container-Optimized OS"
}
