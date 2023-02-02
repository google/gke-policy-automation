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

package gke.scalability.containers

test_containers_above_warn_limit_std {
	not valid with input as {"data": {"monitoring": {"containers": { "name": "containers", "scalar": 352000}}, "gke": {"name": "test-cluster", "autopilot": {"enabled": false}}}}
}

test_containers_within_warn_limit_std {
	valid with input as {"data": {"monitoring": {"containers": { "name": "containers", "scalar": 121303}}, "gke": {"name": "test-cluster", "autopilot": {"enabled": false}}}}
}

test_containers_above_warn_limit_auto {
	not valid with input as {"data": {"monitoring": {"containers": { "name": "containers", "scalar": 121303}}, "gke": {"name": "test-cluster", "autopilot": {"enabled": true}}}}
}

test_containers_within_warn_limit_auto {
	valid with input as {"data": {"monitoring": {"containers": { "name": "containers", "scalar": 18403}}, "gke": {"name": "test-cluster", "autopilot": {"enabled": true}}}}
}
