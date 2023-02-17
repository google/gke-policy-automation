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

package gke.scalability.nodes_per_pool_zone

test_nodes_per_pool_zone_above_warn_limit {
	not valid with input as {"data": {"monitoring": {"nodes_per_pool_zone": { "name": "nodes_per_pool_zone", "vector": {"default-pool": {"europe-central2-a": 642, "europe-central2-b": 734,"europe-central2-a": 821}}}}}}
}

test_nodes_per_pool_zone_below_warn_limit {
    valid with input as {"data": {"monitoring": {"nodes_per_pool_zone": { "name": "nodes_per_pool_zone", "vector": {"default-pool": {"europe-central2-a": 642, "europe-central2-b": 734,"europe-central2-a": 690}}}}}}
}

test_nodes_per_pool_zone_autopilot {
	valid with input as {"data": {"monitoring": {"nodes_per_pool_zone": { "name": "nodes_per_pool_zone", "vector": {"default-pool": {"europe-central2-a": 642, "europe-central2-b": 734,"europe-central2-a": 821}}}}, "gke": {"name": "test-cluster", "autopilot": {"enabled": true}}}}
}
