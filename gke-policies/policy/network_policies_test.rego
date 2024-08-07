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

package gke.policy.network_policies_engine_test

import future.keywords.if
import data.gke.policy.network_policies_engine

test_dataplane_v1_without_netpol if {
	not network_policies_engine.valid with input as {"name": "test-cluster", "addons_config": {"network_policy_config": {"disabled": true}}, "private_cluster_config": {"enable_private_nodes": true}, "network_config": {"datapath_provider": 1}}
}

test_dataplane_v1_with_netpol_disabled if {
	not network_policies_engine.valid with input as {"name": "test-cluster", "addons_config": {"network_policy_config": {}}, "private_cluster_config": {"enable_private_nodes": false}, "network_policy": {"provider": 1, "enabled": false}, "network_config": {"datapath_provider": 1}}
}

test_dataplane_v1_without_netpol_conf if {
	not network_policies_engine.valid with input as {"name": "test-cluster", "addons_config": {"network_policy_config": {}}, "private_cluster_config": {"enable_private_nodes": false}, "network_policy": {}}
}

test_dataplane_v1_with_netpol if {
	network_policies_engine.valid with input as {"name": "test-cluster", "addons_config": {"network_policy_config": {}}, "private_cluster_config": {"enable_private_nodes": false}, "network_policy": {"provider": 1, "enabled": true}}
}

test_dataplane_v2_with_netpol if {
	network_policies_engine.valid with input as {"name": "test-cluster", "addons_config": {"network_policy_config": {"disabled": true}}, "network_config": {"datapath_provider": 2}}
}
