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
# title: Cloud Monitoring and Logging
# description: GKE cluster should use Cloud Logging and Monitoring
# custom:
#   group: Maintenance
package gke.policy.logging_and_monitoring

test_enabled_logging_and_monitoring {
	valid with input as {"Data": {"gke": {
	  "name": "test-cluster",
	  "logging_config": {
		"component_config": {
			"enable_components": "SYSTEM_COMPONENTS", 
			"enable_components": "WORKLOADS"
        }
	  },
	  "monitoring_config": {
		"component_config": {
			"enable_components": "SYSTEM_COMPONENTS"
		}
      }
	}}}
}

test_disabled_logging {
	not valid with input as {"Data": {"gke": {
	  "name": "test-cluster",
	  "logging_config": {"component_config": {}},
	  "monitoring_config": {
		"component_config": {
			"enable_components": "SYSTEM_COMPONENTS"
		  }
      }
	}}}
}

test_disabled_monitoring {
	not valid with input as {"Data": {"gke": {
	  "name": "test-cluster",
	  "logging_config": {
		"component_config": {
			"enable_components": "SYSTEM_COMPONENTS", 
			"enable_components": "WORKLOADS"
        }
	  },
	  "monitoring_config": {"component_config": {}}
	}}}
}
