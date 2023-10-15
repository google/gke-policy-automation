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
# title: GKE private cluster
# description: GKE cluster should be private to ensure network isolation
# custom:
#   group: Security
#   severity: High
#   recommendation: >
#     Once the cluster is created as a public cluster, this cannon be changed.
#     The cluster must be recreated, ensuring that Private cluster mode is enabled.
#   externalURI: https://cloud.google.com/kubernetes-engine/docs/how-to/private-clusters
#   sccCategory: NODES_PUBLIC
#   cis:
#     version: "1.2"
#     id: "5.6.5"
#   dataSource: gke

package gke.policy.private_cluster

default valid := false

valid {
  count(violation) == 0
}

violation[msg] {
  not input.data.gke.private_cluster_config.enable_private_nodes
  msg := "GKE cluster has not enabled private nodes"
}
