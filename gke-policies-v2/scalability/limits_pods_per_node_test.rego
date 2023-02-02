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

package gke.scalability.pods_per_node

test_pods_per_node_above_warn_limit {
	not valid with input as {"data": {"monitoring": {"pods_per_node": { "name": "pods_per_node", "vector": {"default-pool": {"gke-cluster-demo-default-pool-0767d05a-lkkp": 46, "gke-cluster-demo-default-pool-0f74dd4f-3zsv": 97}}}}, "gke":{"node_pools":[{"name": "default-pool", "max_pods_constraint":{"max_pods_per_node":110}}]}}}
}

test_pods_per_node_within_warn_limit {
	valid with input as {"data": {"monitoring": {"pods_per_node": { "name": "pods_per_node", "vector": {"default-pool": {"gke-cluster-demo-default-pool-0767d05a-lkkp": 46, "gke-cluster-demo-default-pool-0f74dd4f-3zsv": 32}}}}, "gke":{"node_pools":[{"name": "default-pool", "max_pods_constraint":{"max_pods_per_node":64}}]}}}
}
