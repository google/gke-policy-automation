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
# title: Enable GKE node local DNS cache
# description: GKE cluster should use node local DNS cache
# custom:
#   group: Scalability
#   severity: Medium
#   recommendation: >
#     Navigate to the GKE page in Google Cloud Console and select the name of the cluster.
#     Under Networking, in the row for "NodeLocal DNSCache", click the edit icon.
#     Select the "Enable NodeLocal DNSCache" checkbox and click "Save changes" button.
#   externalURI: https://cloud.google.com/kubernetes-engine/docs/how-to/nodelocal-dns-cache
#   sccCategory: DNS_CACHE_DISABLED
#   dataSource: gke

package gke.policy.node_local_dns_cache

default valid := false

valid {
	count(violation) == 0
}

violation[msg] {
    not input.data.gke.addons_config.dns_cache_config.enabled = true
    msg := "Cluster is not configured with node local DNS cache"
}
