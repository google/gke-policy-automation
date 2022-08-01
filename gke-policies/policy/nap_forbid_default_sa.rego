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
# title: Forbid default SA in NAP
# description: NAP configuration should not allow default Service Accounts
# custom:
#   group: Security
#   severity: Critical
#   sccCategory: NAP_DEFAULT_SA_CONFIGURED
#   cis:
#     version: "1.2"
#     id: "5.2.1"

package gke.policy.nap_forbid_default_sa

default valid = false

valid {
	count(violation) == 0
}

violation[msg] {
	input.autoscaling.enable_node_autoprovisioning == true
	input.autoscaling.autoprovisioning_node_pool_defaults.service_account == "default"
	msg := "GKE cluster node autoprovisioning should have a dedicated SA configured"
}
