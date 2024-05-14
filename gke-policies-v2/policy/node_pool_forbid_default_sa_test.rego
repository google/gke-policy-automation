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

package gke.policy.node_pool_forbid_default_sa_test

import future.keywords.if
import data.gke.policy.node_pool_forbid_default_sa

test_cluster_with_2_np_and_mixed_sas if {
	not node_pool_forbid_default_sa.valid with input as {"data": {"gke": {"name": "cluster-1", "legacy_abac": {"enabled": false}, "node_pools": [{"name": "default", "config": {"machine_type": "e2-standard-4", "disk_size_gb": 100, "service_account": "default", "image_type": "COS_CONTAINERD", "disk_type": "pd-standard", "workload_metadata_config": {"mode": 2}, "shielded_instance_config": {"enable_integrity_monitoring": true}}, "management": {"auto_repair": true, "auto_upgrade": true}}, {"name": "pool-1", "config": {"machine_type": "e2-standard-2", "disk_size_gb": 100, "oauth_scopes": ["https://www.googleapis.com/auth/cloud-platform"], "service_account": "gke-sa@prj.iam.gserviceaccount.com", "metadata": {"disable-legacy-endpoints": "true"}, "image_type": "COS_CONTAINERD", "disk_type": "pd-standard", "workload_metadata_config": {"mode": 2}, "shielded_instance_config": {"enable_integrity_monitoring": true}}}]}}}
}

test_cluster_with_2_np_and_dedicated_sas if {
	node_pool_forbid_default_sa.valid with input as {"data": {"gke": {"name": "cluster-1", "legacy_abac": {"enabled": false}, "node_pools": [{"name": "default", "config": {"machine_type": "e2-standard-4", "disk_size_gb": 100, "service_account": "gke-sa@prj.iam.gserviceaccount.com", "image_type": "COS_CONTAINERD", "disk_type": "pd-standard", "workload_metadata_config": {"mode": 2}, "shielded_instance_config": {"enable_integrity_monitoring": true}}, "management": {"auto_repair": true, "auto_upgrade": true}}, {"name": "pool-1", "config": {"machine_type": "e2-standard-2", "disk_size_gb": 100, "oauth_scopes": ["https://www.googleapis.com/auth/cloud-platform"], "service_account": "gke-sa@prj.iam.gserviceaccount.com", "metadata": {"disable-legacy-endpoints": "true"}, "image_type": "COS_CONTAINERD", "disk_type": "pd-standard", "workload_metadata_config": {"mode": 2}, "shielded_instance_config": {"enable_integrity_monitoring": true}}}]}}}
}

test_cluster_with_1_np_and_default_sa if {
	not node_pool_forbid_default_sa.valid with input as {"data": {"gke": {"name": "cluster-1", "legacy_abac": {"enabled": false}, "node_pools": [{"name": "default", "config": {"machine_type": "e2-standard-4", "disk_size_gb": 100, "service_account": "default", "image_type": "COS_CONTAINERD", "disk_type": "pd-standard", "workload_metadata_config": {"mode": 2}, "shielded_instance_config": {"enable_integrity_monitoring": true}}, "management": {"auto_repair": true, "auto_upgrade": true}}]}}}
}

test_cluster_with_1_np_and_dedicated_sa if {
	node_pool_forbid_default_sa.valid with input as {"data": {"gke": {"name": "cluster-1", "legacy_abac": {"enabled": false}, "node_pools": [{"name": "pool-1", "config": {"machine_type": "e2-standard-4", "disk_size_gb": 100, "service_account": "gke-sa@prj.iam.gserviceaccount.com", "image_type": "COS_CONTAINERD", "disk_type": "pd-standard", "workload_metadata_config": {"mode": 2}, "shielded_instance_config": {"enable_integrity_monitoring": true}}, "management": {"auto_repair": true, "auto_upgrade": true}}]}}}
}

test_autopilot_with_default if {
	node_pool_forbid_default_sa.valid with input as {"data": {"gke": {"name": "cluster-1", "autopilot": {"enabled": true}, "node_pools": [{"name": "pool-1", "config": {"machine_type": "e2-standard-4", "disk_size_gb": 100, "service_account": "default", "image_type": "COS_CONTAINERD"}}]}}}
}