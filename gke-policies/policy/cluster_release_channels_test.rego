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

package gke.policy.cluster_release_channels

test_cluster_not_enrolled_to_release_channels {
    not valid with input as {"name": "cluster-not-repairing", "release_channel": {}, "node_pools": [{"name": "default", "management": {"auto_repair": true, "auto_upgrade": true }}]}
}

test_cluster_enrolled_to_release_channels {
    valid with input as {"name": "cluster-not-repairing", "release_channel": {"channel": 2 }, "node_pools": [{"name": "default", "management": {"auto_repair": true, "auto_upgrade": true }}]}
}