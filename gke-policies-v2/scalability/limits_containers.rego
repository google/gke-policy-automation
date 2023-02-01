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
# title: Number of containers in a cluster
# description: The total number of containers running in a cluster
# custom:
#   group: Scalability
#   severity: High
#   sccCategory: CONTAINERS_LIMIT
#   dataSource: monitoring

package gke.scalability.pods

default valid = false
default limit = 200000
default threshold = 80

valid {
	count(violation) == 0
}

violation[msg] {
	warn_limit = round(limit * threshold * 0.01)
    input.data.monitoring.containers.scalar > warn_limit
	msg := sprintf("Total number of containers %d has reached warning level %d (limit is %d)", [input.data.monitoring.pods.scalar, warn_limit, limit])
}
