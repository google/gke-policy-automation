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
# title: Control Plane redundancy
# description: GKE cluster should be regional for maximum availability of control plane during upgrades and zonal outages
# custom:
#   group: Availability
package gke.policy.control_plane_redundancy

default valid = false

valid {
  count(violation) == 0
}

violation[msg] {
  some location
  data.gke.rule.location.zonal[location]
  msg := sprintf("invalid GKE Control plane location %q (not regional)", [location])
}
