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

package gke.policy.control_plane_access

test_authorized_networks_enabled {

    valid with input as {"data": {"gke": {"name":"test-cluster","master_authorized_networks_config": {
        "enabled":true,
        "cidr_blocks":[
            {"display_name":"Test Block","cidr_block":"192.168.0.0./16"}
        ]
    }}}}
}

test_authoized_networks_missing{
    not valid with input as {"data": {"gke": {"name":"test-cluster"}}}
}

test_authorized_networks_disabled{
    not valid with input as {"data": {"gke": {"name":"test-cluster","master_authorized_networks_config": {"enabled":false}}}}
}

test_authorized_networks_no_cidrs_block{
    not valid with input as {"data": {"gke": {"name":"test-cluster","master_authorized_networks_config": {"enabled":true}}}}
}

test_authorized_networks_empty_cidrs_block{
    not valid with input as {"data": {"gke": {"name":"test-cluster","master_authorized_networks_config": {
        "enabled":true,
        "cidr_blocks":[]
    }}}}
}
