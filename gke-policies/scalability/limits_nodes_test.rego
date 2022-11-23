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

# METADATA
# title: GKE ConfigMaps Limit
# description: GKE ConfigMaps Limit
# custom:
#   group: Scalability
package gke.scalability.nodes

test_nodes_nbr_not_exceeded_for_private {
	valid with input as {"data": {"monitoring": {"number_of_nodes_by_cluster": { "Name": "number_of_nodes_by_cluster", "Value": "1"}}, "gke": {"name": "test-cluster", "private_cluster_config": {"enable_private_nodes": true}}}}
}

test_nodes_nbr_exceeded_for_private {
	not valid with input as {"data": {"monitoring": {"number_of_nodes_by_cluster": { "Name": "number_of_nodes_by_cluster", "Value": "16000"}}, "gke":  {"name": "test-cluster", "private_cluster_config": {"enable_private_nodes": true}}}}
}

test_nodes_nbr_not_exceeded_for_public {
	valid with input as {"data": {"monitoring": {"number_of_nodes_by_cluster": { "Name": "number_of_nodes_by_cluster", "Value": "1"}}, "gke": {"name": "test-cluster", "private_cluster_config": {"enable_private_nodes": false}}}}
}

test_nodes_nbr_exceeded_for_public {
	not valid with input as {"data": {"monitoring": {"number_of_nodes_by_cluster": { "Name": "number_of_nodes_by_cluster", "Value": "6000"}}, "gke":  {"name": "test-cluster", "private_cluster_config": {"enable_private_nodes": false}}}}
}