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
# title: Number of HPAs in a cluster
# description: The optimal number of Horizontal Pod Autoscalers in a cluster
# custom:
#   group: Scalability
#   severity: Medium
#   recommendation: >
#     Horizontal Pod Autoscaler doesn't have a hard limit on the supported number of HPA objects.
#     However, above a certain number of HPA objects, the period between HPA recalculations may become longer than the standard 15 seconds.
#   externalURI: https://cloud.google.com/kubernetes-engine/docs/concepts/horizontalpodautoscaler#scalability
#   sccCategory: HPAS_OPTIMAL_LIMIT
#   dataSource: monitoring

package gke.scalability.hpas

default valid = false
default limit = 300
default threshold = 80

valid {
	count(violation) == 0
}

violation[msg] {
	warn_limit = round(limit * threshold * 0.01)
	hpas := input.data.monitoring.hpas.scalar 
	hpas > warn_limit
	msg := sprintf("Total number of HPAs %d has reached warning level %d (limit is %d)", [hpas, warn_limit, limit])
}
