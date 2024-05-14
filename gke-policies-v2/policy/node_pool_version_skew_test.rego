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

package gke.policy.node_pool_version_skew_test

import future.keywords.if
import data.gke.policy.node_pool_version_skew

test_empty_master_version if {
    not node_pool_version_skew.valid with input as {"data": {"gke": {"name":"cluster","node_pools":[{"name":"default-pool","version":"1.22.1-gke.600"}]}}}
}

test_empty_nodepool_version if {
    not node_pool_version_skew.valid with input as {"data": {"gke": {"name":"cluster","current_master_version":"1.22.1-gke.600","node_pools":[{"name":"default-pool"}]}}}
}

test_invalid_master_version if {
    not node_pool_version_skew.valid with input as {"data": {"gke": {"name":"cluster","current_master_version":"1.22.A","node_pools":[{"name":"default-pool","version":"1.22.1-gke.600"}]}}}
}

test_invalid_nodepool_version if {
    not node_pool_version_skew.valid with input as {"data": {"gke": {"name":"cluster","current_master_version":"1.22.1-gke.600","node_pools":[{"name":"default-pool","version":"1.22"}]}}}
}

test_different_major if {
    not node_pool_version_skew.valid with input as {"data": {"gke": {"name":"cluster","current_master_version":"1.22.1-gke.600","node_pools":[{"name":"default-pool","version":"2.22.1-gke.200"}]}}}
}

test_greater_minor if {
    not node_pool_version_skew.valid with input as {"data": {"gke": {"name":"cluster","current_master_version":"1.24.10-gke.200","node_pools":[{"name":"default-pool","version":"1.21.5-gke.200"}]}}}
}

test_good_minor if {
    node_pool_version_skew.valid with input as {"data": {"gke": {"name":"cluster","current_master_version":"1.24.10-gke.200","node_pools":[{"name":"default-pool","version":"1.22.5-gke.200"}]}}}
}
