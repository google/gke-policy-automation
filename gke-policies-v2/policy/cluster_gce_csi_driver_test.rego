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

package gke.policy.cluster_gce_csi_driver

test_gce_csi_driver_addon_empty {
    not valid with input as {"data": {"gke": {"name":"cluster-demo","addons_config":{"gce_persistent_disk_csi_driver_config":{}}}}}
}

test_gce_csi_driver_addon_disabled {
    not valid with input as {"data": {"gke": {"name":"cluster-demo","addons_config":{"gce_persistent_disk_csi_driver_config":{"enabled":false}}}}}
}

test_gce_csi_driver_addon_enabled {
   valid with input as {"data": {"gke": {"name":"cluster-demo","addons_config":{"gce_persistent_disk_csi_driver_config":{"enabled":true}}}}}
}