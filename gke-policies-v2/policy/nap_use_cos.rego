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
# title: Configure Container-Optimized OS for Node Auto-Provisioning node pools
# description: Nodes in Node Auto-Provisioning should use Container-Optimized OS
# custom:
#   group: Security
#   severity: High
#   recommendation: >
#     The node image type for node created by Node Autoprovisioning can be changed from the command
#     line only. To change it to the recommended COS with containerid, run the following command:
#     gcloud container clusters update CLUSTER_NAME --enable-autoprovisioning --autoprovisioning-image-type cos_containerd
#   externalURI: https://cloud.google.com/kubernetes-engine/docs/how-to/node-auto-provisioning#default-image-type
#   sccCategory: NAP_COS_UNCONFIGURED
#   cis:
#     version: "1.4"
#     id: "5.5.1"
#   dataSource: gke
package gke.policy.nap_use_cos

import future.keywords.in
import future.keywords.if
import future.keywords.contains

default valid := false

valid if {
	count(violation) == 0
}

violation contains msg if {
	input.data.gke.autoscaling.enable_node_autoprovisioning == true
	not lower(input.data.gke.autoscaling.autoprovisioning_node_pool_defaults.image_type) in { "cos", "cos_containerd"}
	msg := "Cluster is not configured with COS for NAP node pools"
}
