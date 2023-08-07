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

package gke.policy.vpc_native_cluster

test_vpc_native_cluster_with_pods_range {
    valid with input as {"name": "cluster-not-repairing", "ip_allocation_policy": {"use_ip_aliases": true}, "node_pools": [{"name": "default", "network_config": { "pod_range": "gke-cluster-1-vpc-pods-273c12cd", "pod_ipv4_cidr_block": "10.48.0.0/14" }, "management": {"auto_repair": true, "auto_upgrade": true }}]}
}

test_vpc_native_cluster_without_pods_range {
    not valid with input as {"name": "cluster-not-repairing", "node_pools": [{"name": "default", "management": {"auto_repair": true, "auto_upgrade": true }}]}
}

test_vpc_native_cluster_using_ip_aliases {
    valid with input as {"name": "cluster-not-repairing", "ip_allocation_policy": {"use_ip_aliases": true}}
}

test_vpc_native_cluster_not_using_ip_aliases {
    not valid with input as {"name": "cluster-not-repairing", "ip_allocation_policy": {"use_ip_aliases": false}}
}