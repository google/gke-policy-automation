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

package gke.policy.node_pool_multi_zone_test

import future.keywords.if
import data.gke.policy.node_pool_multi_zone

test_node_pool_one_zone if {
    not node_pool_multi_zone.valid with input as {"data": {"gke": {"name": "cluster", "node_pools": [{"name": "default", "locations": ["us-central1-a"]}]}}}
}

test_node_pool_two_zones if {
    node_pool_multi_zone.valid with input as {"data": {"gke": {"name": "cluster", "node_pools": [{"name": "default", "locations": ["us-central1-a", "us-central1-b"]}]}}}
}

test_node_pool_three_zones if {
    node_pool_multi_zone.valid with input as {"data": {"gke": {"name": "cluster", "node_pools": [{"name": "default", "locations": ["us-central1-a", "us-central1-b", "us-central1-c"]}]}}}
}