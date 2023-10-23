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
# title: GKE VPC-native cluster
# description: GKE cluster nodepool should be VPC-native as per our best-practices
# custom:
#   group: Management
#   severity: CRITICAL
#   recommendation: >
#     Once the cluster is created as a route-based cluster, this cannon be changed.
#     The cluster must be recreated, ensuring that VPN-native networking
#     (with an alias IP ranges) is configured.
#   externalURI: https://cloud.google.com/kubernetes-engine/docs/concepts/alias-ips
#   sccCategory: VPC_NATIVE_ROUTING_DISABLED
#   cis:
#     version: "1.4"
#     id: "5.6.2"

package gke.policy.vpc_native_cluster

default valid = false

valid {
  count(violation) == 0
}

violation[msg] {
  not input.node_pools[pool].network_config.pod_ipv4_cidr_block
  msg := sprintf("Nodepool %q of the GKE cluster is not configured to use VPC-native routing", [input.node_pools[pool].name])
}

violation[msg] {
  not input.ip_allocation_policy.use_ip_aliases
  msg := "the GKE cluster is not configured to use VPC-native routing"
}