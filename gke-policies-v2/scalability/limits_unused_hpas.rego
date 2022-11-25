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
# title: GKE Unused HPAs Limit
# description: GKE Unused HPAs Limit
# custom:
#   group: Scalability
#   severity: Low
#   sccCategory: HPAS_UNUSED
#   dataSource: k8s

package gke.scalability.unused_hpas

default valid = false

valid {
	print(violation)
	count(violation) == 0
}

violation[msg] {
	hpas := {object | object := input.data.k8s.Resources[_]; object.data.kind == "HorizontalPodAutoscaler"}
	some i
	not hpas[i].data.status.lastScaleTime
	msg := sprintf("HPA %s in namespace %s never executed since %s", [hpas[i].data.metadata.name, hpas[i].data.metadata.namespace, hpas[i].data.metadata.creationTimestamp])
}
