// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

func getScalabilityMetricsDefaults() []ConfigMetric {
	return []ConfigMetric{
		{
			MetricName: "pods",
			Query:      "sum (kube_pod_info{cluster=$CLUSTER_NAME,location=$CLUSTER_LOCATION,project_id=$CLUSTER_PROJECT})",
		},
		{
			MetricName: "pods_per_node",
			Query:      "sum by (label_cloud_google_com_gke_nodepool,node) ((kube_pod_info{cluster=$CLUSTER_NAME,location=$CLUSTER_LOCATION,project_id=$CLUSTER_PROJECT}) + on(node) group_left(label_cloud_google_com_gke_nodepool) (0 * kube_node_labels{cluster=$CLUSTER_NAME,location=$CLUSTER_LOCATION,project_id=$CLUSTER_PROJECT}))",
		},
		{
			MetricName: "containers",
			Query:      "sum (kube_pod_container_info{cluster=$CLUSTER_NAME,location=$CLUSTER_LOCATION,project_id=$CLUSTER_PROJECT})",
		},
		{
			MetricName: "nodes",
			Query:      "sum (kube_node_info{cluster=$CLUSTER_NAME,location=$CLUSTER_LOCATION,project_id=$CLUSTER_PROJECT})",
		},
		{
			MetricName: "nodes_per_pool_zone",
			Query:      "sum by (label_cloud_google_com_gke_nodepool,label_topology_kubernetes_io_zone) (kube_node_labels{cluster=$CLUSTER_NAME,location=$CLUSTER_LOCATION,project_id=$CLUSTER_PROJECT})",
		},
		{
			MetricName: "services",
			Query:      "sum (kube_service_info{cluster=$CLUSTER_NAME,location=$CLUSTER_LOCATION,project_id=$CLUSTER_PROJECT})",
		},
		{
			MetricName: "services_per_ns",
			Query:      "sum by (exported_namespace) (kube_service_info{cluster=$CLUSTER_NAME,location=$CLUSTER_LOCATION,project_id=$CLUSTER_PROJECT})",
		},
		{
			MetricName: "hpas",
			Query:      "sum (kube_horizontalpodautoscaler_info{cluster=$CLUSTER_NAME,location=$CLUSTER_LOCATION,project_id=$CLUSTER_PROJECT})",
		},
		{
			MetricName: "secrets",
			Query:      "sum (kube_secret_info{cluster=$CLUSTER_NAME,location=$CLUSTER_LOCATION,project_id=$CLUSTER_PROJECT})",
		},
		{
			MetricName: "namespaces",
			Query:      "count (kube_namespace_status_phase{phase=\"Active\", cluster=$CLUSTER_NAME,location=$CLUSTER_LOCATION,project_id=$CLUSTER_PROJECT})",
		},
	}
}
