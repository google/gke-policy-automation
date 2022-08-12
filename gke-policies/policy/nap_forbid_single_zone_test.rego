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

package gke.policy.nap_forbid_single_zone

test_cluster_not_enabled_nap {
    valid with input as {"name": "cluster-without-nap", "autoscaling": {"enable_node_autoprovisioning": false}}
}

test_cluster_enabled_nap_without_enabled_autoprovisioning_locations {
    valid with input as {"name": "cluster-with-nap", "autoscaling": {"enable_node_autoprovisioning": true}}
}

test_cluster_enabled_nap_with_enabled_autoprovisioning_locations_single {
    not valid with input as {"name": "cluster-with-nap", "autoscaling": {"enable_node_autoprovisioning": true, "autoprovisioning_locations": "europe-central2-a" }}
}

test_cluster_enabled_nap_with_enabled_autoprovisioning_locations_multiple {
    valid with input as {"name": "cluster-with-nap", "autoscaling": {"enable_node_autoprovisioning": true, "autoprovisioning_locations": "europe-central2-a",  "autoprovisioning_locations": "europe-central2-b" }}
}