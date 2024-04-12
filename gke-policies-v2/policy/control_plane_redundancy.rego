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
# title: Ensure redundancy of the Control Plane
# description: GKE cluster should be regional for maximum availability of control plane during upgrades and zonal outages
# custom:
#   group: Availability
#   severity: High
#   recommendation: >
#     Once the cluster is created as a zonal cluster, it is not possible to change it's type to regional.
#     The cluster must be recreated, ensuring that regional location type is choosen.
#   externalURI: https://cloud.google.com/kubernetes-engine/docs/concepts/regional-clusters
#   sccCategory: CONTROL_PLANE_ZONAL
#   dataSource: gke

package gke.policy.control_plane_redundancy

import data.gke.rule.cluster.location

default valid := false

valid {
  count(violation) == 0
}

violation[msg] {
  not input.data.gke.location
  msg := "Cluster location infromation is missing"
}

violation[msg] {
  not location.regional(input.data.gke.location)
  msg := sprintf("Cluster location %q is not regional", [input.data.gke.location])
}
