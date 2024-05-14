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

package gke.policy.control_plane_endpoint_test

import future.keywords.if
import data.gke.policy.control_plane_endpoint

test_private_endpoint_enabled if {
    control_plane_endpoint.valid with input as {"data": {"gke": {"name": "test-cluster", "private_cluster_config": {"enable_private_endpoint": true}}}}
}

test_private_endpoint_disabled if {
    not control_plane_endpoint.valid with input as {"data": {"gke": {"name": "test-cluster", "private_cluster_config": {"enable_private_endpoint": false}}}}
}

test_private_cluster_config_missing if {
    not control_plane_endpoint.valid with input as {"data": {"gke": {"name": "test-cluster"}}}
}
