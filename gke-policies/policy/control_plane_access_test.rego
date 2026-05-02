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

package gke.policy.control_plane_access_test

import data.gke.policy.control_plane_access
import future.keywords.if

test_authorized_networks_enabled if {
	control_plane_access.valid with input as {
		"name": "test-cluster",
		"master_authorized_networks_config": {
			"enabled": true,
			"cidr_blocks": [{
				"display_name": "Test Block",
				"cidr_block": "192.168.0.0/16",
			}],
		},
	}
}

test_authorized_networks_missing if {
	not control_plane_access.valid with input as {"name": "test-cluster"}
}

test_authorized_networks_disabled if {
	not control_plane_access.valid with input as {
		"name": "test-cluster",
		"master_authorized_networks_config": {"enabled": false},
	}
}

test_authorized_networks_no_cidrs_block if {
	not control_plane_access.valid with input as {
		"name": "test-cluster",
		"master_authorized_networks_config": {"enabled": true},
	}
}

test_authorized_networks_empty_cidrs_block if {
	not control_plane_access.valid with input as {
		"name": "test-cluster",
		"master_authorized_networks_config": {
			"enabled": true,
			"cidr_blocks": [],
		},
	}
}

test_private_endpoint_only_without_authorized_networks if {
	control_plane_access.valid with input as {
		"name": "test-cluster",
		"private_cluster_config": {"enable_private_endpoint": true},
	}
}

test_public_endpoint_disabled_without_authorized_networks if {
	control_plane_access.valid with input as {
		"name": "test-cluster",
		"control_plane_endpoints_config": {"ip_endpoints_config": {"enable_public_endpoint": false}},
	}
}

test_ip_endpoints_disabled_without_authorized_networks if {
	control_plane_access.valid with input as {
		"name": "test-cluster",
		"control_plane_endpoints_config": {"ip_endpoints_config": {"enabled": false}},
	}
}

test_authorized_networks_enabled_on_control_plane_endpoints_config if {
	control_plane_access.valid with input as {
		"name": "test-cluster",
		"control_plane_endpoints_config": {"ip_endpoints_config": {"authorized_networks_config": {
			"enabled": true,
			"cidr_blocks": [{
				"display_name": "Test Block",
				"cidr_block": "192.168.0.0/16",
			}],
		}}},
	}
}

test_authorized_networks_on_control_plane_endpoints_config_no_cidrs_block if {
	not control_plane_access.valid with input as {
		"name": "test-cluster",
		"control_plane_endpoints_config": {"ip_endpoints_config": {"authorized_networks_config": {"enabled": true}}},
	}
}
