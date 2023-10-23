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
# title: Enable Workload vulnerability scanning
# description: >-
#   The Workload vulnerability scanning is a set of capabilities in the security posture dashboard that automatically 
#   scans for known vulnerabilities in your container images and in specific language packages during the runtime
#   phase of software delivery lifecycle.
# custom:
#   group: Security
#   severity: Medium
#   recommendation: >
#     Enable Container Security API on the cluster project.
#     Next, navigate to the GKE page in Google Cloud Console and select the name of the cluster. Under Security, 
#     in the row for "Workload vulnerability scanning", click the edit icon. Select the
#     "Enable workload vulnerability scanning" checkbox and click "Save changes".
#   externalURI: https://cloud.google.com/kubernetes-engine/docs/concepts/about-workload-vulnerability-scanning
#   sccCategory: WORKLOAD_SCANNING_DISABLED
#   dataSource: gke

package gke.policy.cluster_workload_scanning

default valid := false

valid {
  count(violation) == 0
}

violation[msg] {
  not input.data.gke.security_posture_config.vulnerability_mode == 2
  msg := "GKE cluster has not configured workload vulnerability scanning"
}
