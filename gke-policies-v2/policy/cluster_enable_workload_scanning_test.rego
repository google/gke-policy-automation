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

package gke.policy.cluster_workload_scanning_test

import future.keywords.if
import data.gke.policy.cluster_workload_scanning

test_cluster_enabled_workload_scanning if {
    cluster_workload_scanning.valid with input as {"data": {"gke": {
        "name": "cluster-test", 
        "security_posture_config": {
           "mode": 2,
           "vulnerability_mode": 2
        }
    }}}
}

test_cluster_disabled_workload_scanning if {
    not cluster_workload_scanning.valid with input as {"data": {"gke": {
        "name": "cluster-test", 
        "security_posture_config": {
           "mode": 1,
           "vulnerability_mode": 1
        }
    }}}
}

test_cluster_unknown_workload_scanning if {
    not cluster_workload_scanning.valid with input as {"data": {"gke": {
        "name": "cluster-test", 
        "security_posture_config": {
           "mode": 1,
           "vulnerability_mode": 0
        }
    }}}
}

test_cluster_missing_security_posture if {
    not cluster_workload_scanning.valid with input as {"data": {"gke": {
        "name": "cluster-test"
    }}}
}
