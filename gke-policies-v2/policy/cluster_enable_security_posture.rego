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

# METADATA
# title: Enable Security Posture dashboard
# description: >-
#   The Security Posture feature enables scanning of clusters and running workloads against standards and industry best practices.
#   The dashboard displays the scan results and provides actionable recommendations for concerns. 
# custom:
#   group: Security
#   severity: Medium
#   recommendation: >
#     Enable Container Security API on the cluster project.
#     Next, navigate to the GKE page in Google Cloud Console and select the name of the cluster. Under Security, in the row for
#     "Security posture", click the edit icon. Select the "Enable security posture" checkbox and click "Save changes".
#   externalURI: https://cloud.google.com/kubernetes-engine/docs/concepts/about-security-posture-dashboard
#   sccCategory: SECURITY_POSTURE_DISABLED
#   dataSource: gke

package gke.policy.cluster_security_posture

default valid := false

valid {
  count(violation) == 0
}

violation[msg] {
  not input.data.gke.security_posture_config.mode == 2
  msg := "GKE cluster has not disabled Security Posture"
}
