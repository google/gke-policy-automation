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
# title: Enable Kubernetes secrets encryption
# description: GKE cluster should use encryption for kubernetes application secrets
# custom:
#   group: Security
#   severity: Medium
#   recommendation: >
#     This recommendation requires KMS key to be provisioned as a prerequsite.
#     Navigate to the GKE page in Google Cloud Console and select the name of the cluster.
#     Under Security, in the row for "Application-layer secrets encryption", click the edit icon.
#     Select the "Encrypt secrets at the application layer" checkbox. Select your KMS key from
#     the list or provide it's resource name. Click "Save changes" once done.
#   externalURI: https://cloud.google.com/kubernetes-engine/docs/how-to/encrypting-secrets
#   sccCategory: SECRETS_ENCRYPTION_DISABLED
#   cis:
#     version: "1.4"
#     id: "5.3.1"
#   dataSource: gke
package gke.policy.secret_encryption

import future.keywords.if
import future.keywords.contains

default valid := false

valid if {
	count(violation) == 0
}

violation contains msg if {
	input.data.gke.database_encryption.state != 1
	msg := "Cluster is not configured with kubernetes secrets encryption"
}
