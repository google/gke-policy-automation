# Copyright 2023 Google LLC
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

package gke.policy.control_plane_basic_auth

test_cluster_without_basic_auth {
    valid with input as {"data": {"gke": {
        "name": "cluster-test", 
        "master_auth": {
           "cluster_ca_certificate": "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUVMVENDQXBXZ0F3SUJBZ0lSQUpIeTI1V..."
        }
    }}}
}

test_cluster_with_basic_auth {
    not valid with input as {"data": {"gke": {
        "name": "cluster-test", 
        "master_auth": {
           "cluster_ca_certificate": "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUVMVENDQXBXZ0F3SUJBZ0lSQUpIeTI1V...",
           "username": "user",
           "password": "aabbccddeeffgghh"
        }
    }}}
}
