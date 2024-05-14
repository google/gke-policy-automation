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

package gke.policy.nap_use_cos_test

import future.keywords.if
import data.gke.policy.nap_use_cos

test_cluster_not_enabled_nap if {
    nap_use_cos.valid with input as {"name": "cluster-without-nap", "autoscaling": {"enable_node_autoprovisioning": false}}
}

test_cluster_enabled_nap_without_cos if {
    not nap_use_cos.valid with input as {"name": "cluster-with-nap", "autoscaling": {"enable_node_autoprovisioning": true, "autoprovisioning_node_pool_defaults": {"image_type": "ANOTHER"}}}
}

test_cluster_enabled_nap_with_cos_containerd if {
    nap_use_cos.valid with input as {"name": "cluster-with-nap", "autoscaling": {"enable_node_autoprovisioning": true, "autoprovisioning_node_pool_defaults": {"image_type": "COS_CONTAINERD"}} }
}

test_cluster_enabled_nap_with_cos if {
    nap_use_cos.valid with input as {"name": "cluster-with-nap", "autoscaling": {"enable_node_autoprovisioning": true, "autoprovisioning_node_pool_defaults": {"image_type": "COS"}} }
}
