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

package gke.policy.node_pool_integrity_monitoring

test_empty_shielded_instance_config {
    not valid with input as {"Data": {"gke": {"name": "cluster", "node_pools": [{"name": "default-pool", "config": {"machine_type": "e2-medium", "shielded_instance_config":{}}}]}}}
}

test_disabled_integrity_monitoring {
    not valid with input as {"Data": {"gke": {"name": "cluster", "node_pools": [{"name": "default-pool", "config": {"machine_type": "e2-medium", "shielded_instance_config":{"enable_integrity_monitoring": false}}}]}}}
}

test_enabled_integrity_monitoring {
    valid with input as {"Data": {"gke": {"name": "cluster", "node_pools": [{"name": "default-pool", "config": {"machine_type": "e2-medium", "shielded_instance_config":{"enable_integrity_monitoring": true}}}]}}}
}