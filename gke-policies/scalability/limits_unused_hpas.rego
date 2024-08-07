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
package gke.scalability.unused_hpas

import future.keywords.in
import future.keywords.if
import future.keywords.contains

default valid := false

valid if {
	count(violation) == 0
}

violation contains msg if {
	hpas := {object | object := input.Resources[_]; object.Data.kind == "HorizontalPodAutoscaler"}
	some hpa in hpas
	not hpa.Data.status.lastScaleTime
	msg := sprintf("HPA %s in namespace %s never executed since %s", [hpa.Data.metadata.name, hpa.Data.metadata.namespace, hpa.Data.metadata.creationTimestamp])
}
