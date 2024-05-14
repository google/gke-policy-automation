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

package gke.policy.node_pool_autorepair_test

import future.keywords.if
import data.gke.policy.node_pool_autorepair

test_autorepair_for_node_pool_enabled if {
    node_pool_autorepair.valid with input as {"data": {"gke": {"name": "cluster-not-repairing", "node_pools": [{"name": "default", "management": {"auto_repair": true, "auto_upgrade": true }}]}}}
}

test_autorepair_for_node_pool_disabled if {
    not node_pool_autorepair.valid with input as {"data": {"gke": {"name": "cluster-not-repairing", "node_pools": [{"name": "default", "management": {"auto_repair": false, "auto_upgrade": true }}]}}}
}

test_autorepair_for_multiple_node_pools_but_only_one_disabled if {
    not node_pool_autorepair.valid with input as {"data": {"gke": {"name": "cluster-not-repairing", "node_pools": [{"name": "default", "management": {"auto_repair": true, "auto_upgrade": true }},{"name": "custom", "management": {"auto_repair": false, "auto_upgrade": true }}]}}}
}

test_autorepair_for_node_pool_empty_managment if {
    not node_pool_autorepair.valid with input as {"data": {"gke": {"name": "cluster-not-repairing", "node_pools": [{"name": "default", "management": {}}]}}}
}

test_autorepair_for_managment_without_auto_repair_field if {
    not node_pool_autorepair.valid with input as {"data": {"gke": {"name": "cluster-not-repairing", "node_pools": [{"name": "default", "management": {"auto_upgrade": true }}]}}}
}