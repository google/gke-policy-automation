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

package gke.scalability.secrets_with_enc

test_secrets_with_enc_above_warn_limit {
	not valid with input as {"data": {"monitoring": {"secrets": { "name": "secrets", "scalar": 28000}}, "gke": {"name": "cluster-1", "database_encryption": {"state": 1}}}}
}

test_secrets_with_enc_below_warn_limit {
	valid with input as {"data": {"monitoring": {"secrets": { "name": "secrets", "scalar": 307}}, "gke": {"name": "cluster-1", "database_encryption": {"state": 1}}}}
}

test_secrets_no_enc_above_warn_limit {
	valid with input as {"data": {"monitoring": {"secrets": { "name": "secrets", "scalar": 28000}}, "gke": {"name": "cluster-1", "database_encryption": {"state": 2}}}}
}


