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
# title: Ensure that nodes in Node Auto-Provisioning node pools will use integrity monitoring
# description: Nodes in Node Auto-Provisioning should use integrity monitoring
# custom:
#   group: Security
#   severity: Critical
#   sccCategory: NAP_INTEGRITY_MONITORING_DISABLED
#   recommendation: >
#     The Integrity Monitoring can be enabled for Node Autoprovisioning using the configuration file only.
#     Prepare the YAML configuration file with Integrity Monitoring enabled:
#       shieldedInstanceConfig:
#         enableSecureBoot: true
#         enableIntegrityMonitoring: false
#     Next, run the following gcloud command:
#     gcloud container clusters update CLUSTER_NAME --enable-autoprovisioning --autoprovisioning-config-file FILE_NAME
#     Refer to the official documentation for more details.
#   externalURI: https://cloud.google.com/kubernetes-engine/docs/how-to/node-auto-provisioning#node_integrity
#   cis:
#     version: "1.2"
#     id: "5.2.1"
#   dataSource: gke

package gke.policy.nap_integrity_monitoring

default valid = false

valid {
	count(violation) == 0
}

violation[msg] {
	input.Data.gke.autoscaling.enable_node_autoprovisioning == true
	input.Data.gke.autoscaling.autoprovisioning_node_pool_defaults.shielded_instance_config.enable_integrity_monitoring == false
	
	msg := "GKE cluster Node Auto-Provisioning configuration use integrity monitoring"
}
