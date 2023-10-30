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
# title: Control plane user certificate authentication
# description: >-
#   Disable Client Certificates, which require certificate rotation, for authentication. Instead,
#   use another authentication method like OpenID Connect.
# custom:
#   group: Security
#   severity: High
#   recommendation: >
#     Client certificate authentication cannot be disabled on the existing cluster.
#     The new cluster has to be created with a "Client certificate" option disabled.
#   externalURI: https://cloud.google.com/kubernetes-engine/docs/how-to/api-server-authentication#disabling_authentication_with_a_client_certificate
#   sccCategory: CONTROL_PLANE_CERTIFICATE_AUTH
#   cis:
#     version: "1.4"
#     id: "5.8.2"
#   dataSource: gke

package gke.policy.control_plane_certificate_auth

default valid := false

valid {
	count(violation) == 0
}

violation[msg] {
	input.data.gke.master_auth.client_certificate
	msg := "The GKE cluster authentication should not be configured with a client certificate"
}

violation[msg] {
	input.data.gke.master_auth.client_key
	msg := "The GKE cluster authentication should not be configured with a client key"
}
