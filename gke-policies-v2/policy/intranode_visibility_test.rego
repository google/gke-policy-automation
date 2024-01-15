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

package gke.policy.networkConfig

test_enabled_intranode_visibility {
	valid with input as {"data": {"gke": {"name": "test-cluster", "networkConfig": { "enableIntraNodeVisibility": true }}}}
}

test_disabled_intranode_visibility {
	not valid with input as {"data": {"gke": {"name": "test-cluster", "networkConfig": { "enableIntraNodeVisibility": false }}}}
}

