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

# METADATA
# title: Enable GKE upgrade notifications
# description: GKE cluster should be proactively receive updates about GKE upgrades and GKE versions
# custom:
#   group: Management
#   severity: Low
#   recommendation: >
#     Navigate to the GKE page in Google Cloud Console and select the name of the cluster.
#     Under Automation, in the row for "Notifications", click the edit icon.
#     Select the "Enable Notifications" checkbox. From the drop-down list, select the Pub/Sub topic where you want to send update notifications.
#     To filter notifications, select the Filter notification types checkbox, and then select the notification types you want to receive.
#     Click "Save changes" once done.
#   externalURI: https://cloud.google.com/kubernetes-engine/docs/concepts/cluster-notifications
#   sccCategory: UPDATE_NOTIFICATIONS_DISABLED
#   dataSource: gke

package gke.policy.cluster_receive_updates

default valid := false

valid {
  count(violation) == 0
}

violation[msg] {
  not input.data.gke.notification_config.pubsub.enabled
  msg := "Cluster is not configured with upgrade notifications"
}

violation[msg] {
  not input.data.gke.notification_config.pubsub.topic
  msg := "Cluster is not configured with upgrade notofications topic"
}
