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
# title: GKE ConfigMaps Limit
# description: GKE ConfigMap Limit
# custom:
#   group: Scalability
package gke.scalability.configmaps

default configmaps_limit = 30 #value is ONLY for demo purpose, does not reflect a real limit

valid {
	objects := {keep | keep := input.Resources[_]; keep.Data.kind == "ConfigMap"}
	print("configmaps found: ", count(objects))
	count(objects) <= configmaps_limit
}
