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
# title: Cloud Monitoring and Logging
# description: GKE cluster should use Cloud Logging and Monitoring
# custom:
#   group: Maintenance
#   severity: Medium
#   recommendation: >
#     Navigate to the GKE page in Google Cloud Console and select the name of the cluster.
#     Under Features, in the row for "Cloud Logging", click the edit icon.
#     Select the "Enable Cloud Logging" checkbox. From the drop-down list, select System and Workloads.
#     Click "Save changes" once done.
#     Under Features, in the row for "Cloud Monitoring", click the edit icon.
#     Select the "Enable Cloud Monitoring" checkbox. From the drop-down list, select System.
#     Click "Save changes" once done.
#   externalURI: https://cloud.google.com/stackdriver/docs/solutions/gke/installing
#   sccCategory: LOGGING_OR_MONITORING_DISABLED
#   cis:
#     version: "1.2"
#     id: "5.7.1"
#   dataSource: gke

package gke.policy.logging_and_monitoring

default valid = false

valid {
	count(violation) == 0
}

violation[msg] {
	not input.data.gke.logging_config.component_config.enable_components

	msg := "The GKE cluster does not have Cloud Logging enabled"
}

violation[msg] {
	not input.data.gke.monitoring_config.component_config.enable_components

	msg := "The GKE cluster does not have Cloud Monitoring enabled"
}