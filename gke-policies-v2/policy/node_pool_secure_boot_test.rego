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

package gke.policy.node_pool_secure_boot_test

import future.keywords.if
import data.gke.policy.node_pool_secure_boot

test_empty_shielded_instance_config if {
    not node_pool_secure_boot.valid with input as {"data": {"gke": {"name": "cluster", "node_pools": [{"name": "default-pool", "config": {"machine_type": "e2-medium", "shielded_instance_config":{}}}]}}}
}

test_disabled_secure_boot if {
    not node_pool_secure_boot.valid with input as {"data": {"gke": {"name": "cluster", "node_pools": [{"name": "default-pool", "config": {"machine_type": "e2-medium", "shielded_instance_config":{"enable_secure_boot": false}}}]}}}
}

test_enabled_secure_boot if {
    node_pool_secure_boot.valid with input as {"data": {"gke": {"name": "cluster", "node_pools": [{"name": "default-pool", "config": {"machine_type": "e2-medium", "shielded_instance_config":{"enable_secure_boot": true}}}]}}}
}