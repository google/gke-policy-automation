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

package gke.policy.node_local_dns_cache

test_enabled_node_local_dns_cache {
	valid with input as {"name": "test-cluster", "addons_config": { "dns_cache_config": { "enabled": true }}}
}

test_absent_dns_cache_config {
	not valid with input as {"name": "test-cluster", "addons_config": {}}
}

test_disabled_node_local_dns_cache {
	not valid with input as {"name": "test-cluster", "addons_config": { "dns_cache_config": { "enabled": false }}}
}
