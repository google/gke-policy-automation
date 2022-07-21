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
# title: Enable binary authorization in the cluster
# description: GKE cluster should enable for deploy-time security control that ensures only trusted container images are deployed to gain tighter control over your container environment.
# custom:
#   group: Management
#   severity: Low
#   sccCategory: BINARY_AUTHORIZATION_DISABLED
#   cis:
#     version: "1.2"
#     id: "5.10.5"

package gke.policy.cluster_binary_authorization

default valid = false

valid {
  count(violation) == 0
}

violation[msg] {
  not input.binary_authorization.enabled
  msg := "GKE cluster has not configured binary authorization policies"
}
