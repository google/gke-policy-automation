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

package gke.policy.node_pool_use_cos

test_node_pool_using_cos {
    valid with input as {"Data": {"gke": {"name": "cluster-cos", "node_pools": [{"name": "default", "config": {"image_type": "cos"}}]}}}
}

test_node_pool_using_cos_containerd {
    valid with input as {"Data": {"gke": {"name": "cluster-cos", "node_pools": [{"name": "default", "config": {"image_type": "cos_containerd"}}]}}}
}

test_node_pool_using_cos_uppercase {
    valid with input as {"Data": {"gke": {"name": "cluster-cos", "node_pools": [{"name": "default", "config": {"image_type": "COS"}}]}}}
}

test_node_pool_using_cos_containerd_uppercase {
    valid with input as {"Data": {"gke": {"name": "cluster-cos", "node_pools": [{"name": "default", "config": {"image_type": "COS_CONTAINERD"}}]}}}
}

test_node_pool_not_using_cos {
    not valid with input as {"Data": {"gke": {"name": "cluster-not-cos", "node_pools": [{"name": "default", "config": {"image_type": "another_image"}}]}}}
}

test_multiple_node_pool_using_cos_but_only_one {
    not valid with input as {"Data": {"gke": {"name": "cluster-not-cos", "node_pools": [{"name": "default", "config": {"image_type": "cos"}},{"name": "custom", "config": {"image_type": "other"}}]}}}
}

test_multiple_node_pool_using_cos {
    valid with input as {"Data": {"gke": {"name": "cluster-cos", "node_pools": [{"name": "default", "config": {"image_type": "cos"}},{"name": "custom", "config": {"image_type": "cos_containerd"}}]}}}
}