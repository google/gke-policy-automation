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

package gke.scalability.services_per_ns

test_services_per_ns_above_warn_limit {
	not valid with input as {"data": {"monitoring": {"services_per_ns": { "name": "services_per_ns", "vector": {"default": 1, "kube-system": 4, "demo-test":4523}}}}}
}

test_services_per_ns_within_warn_limit {
	valid with input as {"data": {"monitoring": {"services_per_ns": { "name": "services_per_ns", "vector": {"default": 1, "kube-system": 4, "demo-test":453}}}}}
}
