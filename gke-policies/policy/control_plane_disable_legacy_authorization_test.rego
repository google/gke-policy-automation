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

package gke.policy.disable_legacy_authorization

test_enabled_legacy_authorization {
	not valid with input as {"name": "cluster-1", "legacy_abac": {"enabled": true}, "node_pools": [{"name": "default", "management": {"auto_repair": true, "auto_upgrade": true}}]}
}

test_disabled_legacy_authorization {
	valid with input as {"name": "cluster-1", "legacy_abac": {}, "node_pools": [{"name": "default", "management": {"auto_repair": true, "auto_upgrade": true}}]}
}
