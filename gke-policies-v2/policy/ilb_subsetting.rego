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
# title: GKE L4 ILB Subsetting
# description: GKE cluster should use GKE L4 ILB Subsetting if nodes > 250
# custom:
#   group: Scalability
#   severity: High
#   recommendation: >
#     Navigate to the GKE page in Google Cloud Console and select the name of the cluster.
#     Under Networking, in the row for "Subsetting for L4 Internal Load Balancers", click the edit icon.
#     Select the "Enable subsetting for L4 internal load balancers" checkbox and click "Save changes".
#   externalURI: https://cloud.google.com/kubernetes-engine/docs/how-to/internal-load-balancing#subsetting
#   sccCategory: ILB_SUBSETTING_DISABLED
#   dataSource: gke

package gke.policy.enable_ilb_subsetting

default valid := false

valid {
	count(violation) == 0
}

violation[msg] {
	input.data.gke.current_node_count > 250
    not input.data.gke.network_config.enable_l4ilb_subsetting = true

	msg := sprintf("The GKE cluster has %v nodes but is not configured to use L4 ILB Subsetting", [input.data.gke.current_node_count])

}
