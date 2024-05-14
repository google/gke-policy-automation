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

package gke.policy.nap_integrity_monitoring_test

import future.keywords.if
import data.gke.policy.nap_integrity_monitoring

test_cluster_not_enabled_nap if {
    nap_integrity_monitoring.valid with input as {"data": {"gke": {"name": "cluster-without-nap", "autoscaling": {"enable_node_autoprovisioning": false}}}}
}

test_cluster_enabled_nap_with_integrity_monitoring_enabled if {
    nap_integrity_monitoring.valid with input as {"data": {"gke": {
        "name": "cluster-with-nap", 
        "autoscaling": {
            "enable_node_autoprovisioning": true, 
            "autoprovisioning_node_pool_defaults": {
                "shielded_instance_config":{"enable_integrity_monitoring": true}
            }
        }
    }}}
}

test_cluster_enabled_nap_without_integrity_monitoring_enabled if {
    not nap_integrity_monitoring.valid with input as {"data": {"gke": {
        "name": "cluster-with-nap", 
        "autoscaling": {
            "enable_node_autoprovisioning": true, 
            "autoprovisioning_node_pool_defaults": {
                "shielded_instance_config":{"enable_integrity_monitoring": false}
            },
        }
    }}}
}