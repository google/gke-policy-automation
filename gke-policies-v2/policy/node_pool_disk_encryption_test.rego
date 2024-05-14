# Copyright 2023 Google LLC
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

package gke.policy.node_pool_cmek_test

import future.keywords.if
import data.gke.policy.node_pool_cmek

test_cluster_node_pool_with_cmek if {
    node_pool_cmek.valid with input as {"data": {"gke": {
        "name": "cluster-test", 
        "node_pools": [
            {
                "name": "default",
                "config": {
                    "boot_disk_kms_key": "projects/test/locations/europe-central2/keyRings/test/cryptoKeys/cluster-test"
                },
            },
        ]
    }}}
}

test_cluster_node_pool_without_cmek if {
    not node_pool_cmek.valid with input as {"data": {"gke": {
        "name": "cluster-test", 
        "node_pools": [
            {
                "name": "default",
            },
        ]
    }}}
}
