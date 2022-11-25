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

package gke.policy.rbac_security_group_enabled

test_rbac_group_enabled {
    valid with input as {"name": "cluster1", "authenticator_groups_config": {"enabled": true}}
}

test_rbac_group_disabled {
    not valid with input as {"name": "cluster1", "authenticator_groups_config": {"enabled": false}}
}

test_rbac_group_without_authenticator_group {
    not valid with input as {"name": "cluster1"}
}