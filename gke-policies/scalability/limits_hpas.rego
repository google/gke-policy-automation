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
# title: GKE HPAs Limit
# description: GKE HPAs Limit
# custom:
#   group: Scalability
#   severity: High
#   sccCategory: HPAS_LIMIT
#   dataSource: k8s

package gke.scalability.hpas

default valid = false

default hpas_limit = 2 #the value is ONLY for demo purpose, does not reflect a real limit

valid {
	count(violation) == 0
}

violation[msg] {
	hpas := {object | object := input.Data.k8s.Resources[_]; object.Data.kind == "HorizontalPodAutoscaler"}
	count(hpas) > hpas_limit
	msg := sprintf("HPAs found: %d higher than the limit: %d", [count(hpas), hpas_limit])
	print(msg)
}
