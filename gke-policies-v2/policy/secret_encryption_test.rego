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

package gke.policy.secret_encryption

test_enabled_encryption {
	valid with input as {"data": {"gke": {"name": "cluster-1", "database_encryption": {"state": 1}}}}
}

test_disabled_encryption {
	not valid with input as {"data": {"gke": {"name": "cluster-1", "database_encryption": {"state": 2}}}}
}
