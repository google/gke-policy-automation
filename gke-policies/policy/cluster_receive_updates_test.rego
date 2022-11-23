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

package gke.policy.cluster_receive_updates

test_cluster_with_topic_configured {
    valid with input as {"data": {"gke": {"name": "cluster-not-repairing", "release_channel": {}, "notification_config": { "pubsub": { "enabled": true, "topic": "projects/project-id/topics/cluster-updates-topic"}}}}}
}

test_cluster_without_notification_config {
    not valid with input as {"data": {"gke": {"name": "cluster-not-repairing", "release_channel": {"channel": 2 }, "node_pools": [{"name": "default", "management": {"auto_repair": true, "auto_upgrade": true }}]}}}
}

test_cluster_without_topic_specified {
    not valid with input as {"data": {"gke": {"name": "cluster-not-repairing", "release_channel": {"channel": 2 }, "notification_config": { "pubsub": { "enabled": true }}}}}
}

test_cluster_without_pubsub_enabled {
    not valid with input as {"data": {"gke": {"name": "cluster-not-repairing", "release_channel": {"channel": 2 }, "notification_config": { "pubsub": { "enabled": false, "topic": "projects/project-id/topics/cluster-updates-topic"}}}}}
}