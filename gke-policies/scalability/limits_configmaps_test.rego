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
# description: GKE ConfigMaps Limit
# custom:
#   group: Scalability
package gke.scalability.configmaps

test_configmap_underusage {
	valid with input as {"data": {"k8s": {"Resources": [{"Type": {"Group": "", "Version": "v1", "Name": "configmaps", "Namespaced": true}, "data": {"apiVersion": "v1", "kind": "ConfigMap", "metadata": {"annotations": {"control-plane.alpha.kubernetes.io/leader": ""}, "creationTimestamp": "2022-06-21T10:10:31Z", "managedFields": [{"apiVersion": "v1", "fieldsType": "FieldsV1", "fieldsV1": {"f:metadata": {"f:annotations": {".": {}, "f:control-plane.alpha.kubernetes.io/leader": {}}}}, "manager": "manager", "operation": "Update", "time": "2022-06-21T10:10:31Z"}], "name": "", "namespace": "asm-system", "resourceVersion": "", "uid": ""}}}]}}}
}

test_configmap_overusage {
	not valid with input as {"data": {"k8s": {"Resources": [{"Type": {"Group": "", "Version": "v1", "Name": "configmaps", "Namespaced": true}, "data": {"apiVersion": "v1", "kind": "ConfigMap", "metadata": {"annotations": {"control-plane.alpha.kubernetes.io/leader": ""}, "creationTimestamp": "2022-06-21T10:10:31Z", "managedFields": [{"apiVersion": "v1", "fieldsType": "FieldsV1", "fieldsV1": {"f:metadata": {"f:annotations": {".": {}, "f:control-plane.alpha.kubernetes.io/leader": {}}}}, "manager": "manager", "operation": "Update", "time": "2022-06-21T10:10:31Z"}], "name": "test1", "namespace": "asm-system", "resourceVersion": "", "uid": ""}}}, {"Type": {"Group": "", "Version": "v1", "Name": "configmaps", "Namespaced": true}, "data": {"apiVersion": "v1", "kind": "ConfigMap", "metadata": {"annotations": {"control-plane.alpha.kubernetes.io/leader": ""}, "creationTimestamp": "2022-06-21T10:10:31Z", "managedFields": [{"apiVersion": "v1", "fieldsType": "FieldsV1", "fieldsV1": {"f:metadata": {"f:annotations": {".": {}, "f:control-plane.alpha.kubernetes.io/leader": {}}}}, "manager": "manager", "operation": "Update", "time": "2022-06-21T10:10:31Z"}], "name": "test2", "namespace": "asm-system", "resourceVersion": "", "uid": ""}}}, {"Type": {"Group": "", "Version": "v1", "Name": "configmaps", "Namespaced": true}, "data": {"apiVersion": "v1", "kind": "ConfigMap", "metadata": {"annotations": {"control-plane.alpha.kubernetes.io/leader": ""}, "creationTimestamp": "2022-06-21T10:10:31Z", "managedFields": [{"apiVersion": "v1", "fieldsType": "FieldsV1", "fieldsV1": {"f:metadata": {"f:annotations": {".": {}, "f:control-plane.alpha.kubernetes.io/leader": {}}}}, "manager": "manager", "operation": "Update", "time": "2022-06-21T10:10:31Z"}], "name": "test3", "namespace": "asm-system", "resourceVersion": "", "uid": ""}}}]}}}
}
