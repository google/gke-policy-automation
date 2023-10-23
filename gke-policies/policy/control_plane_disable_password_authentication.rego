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
# title: Control plane user basic authentication
# description: >-
#   Disable Basic Authentication (basic auth) for API server authentication as it uses static
#   passwords which need to be rotated.
# custom:
#   group: Security
#   severity: Critical
#   recommendation: >
#     Navigate to the GKE page in Google Cloud Console and select the name of the cluster. Under Security, 
#     in the row for "Basic authentication", click the edit icon. Unselect the "Enable basic authentication"
#     checkbox and click "Save changes".
#   externalURI: https://cloud.google.com/kubernetes-engine/docs/how-to/api-server-authentication#disabling_authentication_with_a_static_password
#   sccCategory: CONTROL_PLANE_BASIC_AUTH
#   cis:
#     version: "1.4"
#     id: "5.8.1"
#   dataSource: gke

package gke.policy.control_plane_basic_auth

default valid := false

valid {
	count(violation) == 0
}

violation[msg] {
	input.master_auth.password
	msg := "The GKE cluster authentication is configured with a client password"
}

violation[msg] {
	input.master_auth.username
	msg := "The GKE cluster authentication is configured with a client username"
}
