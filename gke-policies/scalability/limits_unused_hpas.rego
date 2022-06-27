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
# title: GKE unused HPAs 
# description: GKE unused HPAs 
# custom:
#   group: Scalability
package gke.scalability.unused_hpas

default valid = false

valid {
	violation
}

violation[msg] {
	hpas := {object | object := input.Resources[_]; object.Data.kind == "HorizontalPodAutoscaler"}

	some i
	not hpas[i].Data.status.lastScaleTime
	msg := sprintf("HPA never executed %s namespace %s", [hpas[i].Data.metadata.name, hpas[i].Data.metadata.namespace])
}
