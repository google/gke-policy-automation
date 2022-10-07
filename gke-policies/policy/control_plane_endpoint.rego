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
# title: Control Plane endpoint visibility
# description: Control Plane endpoint should be locked from external access
# custom:
#   group: Security
#   severity: High
#   recommendation: >
#     Once the cluster is created without enabling private control plane address only, this cannon be changed.
#     The cluster must be recreated, ensuring that Private cluster mode is enabled and
#     Public endpoint access is disabled.
#   externalURI: https://cloud.google.com/kubernetes-engine/docs/concepts/private-cluster-concept#endpoints_in_private_clusters
#   sccCategory: CONTROL_PLANE_ENDPOINT_PUBLIC
#   cis:
#     version: "1.2"
#     id: "5.6.4"

package gke.policy.control_plane_endpoint

default valid = false

valid {
  count(violation) == 0
}

violation[msg] {
  not input.private_cluster_config.enable_private_endpoint
  msg := "GKE cluster has not enabled private endpoint" 
}
