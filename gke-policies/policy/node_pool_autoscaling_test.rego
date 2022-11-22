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

package gke.policy.node_pool_autoscaling

test_node_pool_autoscaling_enabled {
    valid with input as {"Data": {"gke": {"name": "cluster", "node_pools": [{"name": "default", "autoscaling": {"enabled": true}}]}}}
}

test_node_pool_autoscaling_disabled {
    not valid with input as {"Data": {"gke": {"name": "cluster", "node_pools": [{"name": "default", "autoscaling": {"enabled": false}}]}}}
}

test_multiple_node_pool_autoscaling_but_only_one_enabled {
    not valid with input as {"Data": {"gke": {"name": "cluster", "node_pools": [{"name": "default", "autoscaling": {"enabled": true}}, {"name": "custom", "autoscaling": {"enabled": false}}]}}}
}

test_multiple_node_pool_autoscaling_enabled {
    valid with input as {"Data": {"gke": {"name": "cluster", "node_pools": [{"name": "default", "autoscaling": {"enabled": true}}, {"name": "custom", "autoscaling": {"enabled": true}}]}}}
}

test_node_pool_without_autoscaling_field {
    not valid with input as {"name": "cluster", "node_pools": [{"name": "default"}]}
}