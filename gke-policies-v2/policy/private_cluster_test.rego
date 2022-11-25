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

package gke.policy.private_cluster

test_private_nodes_enabled {
    valid with input as {"data": {"gke": {"name": "test-cluster", "private_cluster_config": {"enable_private_nodes": true}}}}
}

test_private_nodes_disabled {
    not valid with input as {"data": {"gke": {"name": "test-cluster", "private_cluster_config": {"enable_private_nodes": false}}}}
}

test_private_cluster_config_missing {
    not valid with input as {"data": {"gke": {"name": "test-cluster"}}}
}