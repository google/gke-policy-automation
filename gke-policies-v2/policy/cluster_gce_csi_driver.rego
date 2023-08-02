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
# title: Use Compute Engine persistent disk CSI driver
# description: Automatic deployment and management of the Compute Engine persistent disk CSI driver. The driver provides support for features like customer managed encryption keys or volume snapshots.
# custom:
#   group: Management
#   severity: Medium
#   recommendation: >
#     Navigate to the GKE page in Google Cloud Console and select the name of the cluster.
#     Under Features, in the row for "Compute Engine persistent disk CSI Driver ", click the edit icon.
#     Select the "Enable Compute Engine Persistent Disk CSI Driver " checkbox and click "Save changes".
#   externalURI: https://cloud.google.com/kubernetes-engine/docs/how-to/persistent-volumes/gce-pd-csi-driver
#   sccCategory: GCE_CSI_DRIVER_DISABLED
#   dataSource: gke

package gke.policy.cluster_gce_csi_driver

default valid = false

valid {
  count(violation) == 0
}

violation[msg] {
  not input.data.gke.addons_config.gce_persistent_disk_csi_driver_config.enabled
  msg := "GKE cluster has not configured GCE persistent disk CSI driver"
}
