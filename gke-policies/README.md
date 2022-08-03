# GKE Policy Automation library

## Policy structure

Please refer to the [Policy Authoring Guide](./AUTHORING.md) for details about structure
of our policy files.

## Available Policies

<!-- BEGIN POLICY-DOC -->
|Group|Title|Description|File|
|-|-|-|-|
|Availability|Control Plane redundancy|GKE cluster should be regional for maximum availability of control plane during upgrades and zonal outages|[gke-policies/policy/control_plane_redundancy.rego](https://github.com/google/gke-policy-automation/blob/main/gke-policies/policy/control_plane_redundancy.rego)|
|Availability|Multi-zone node pools|GKE node pools should be regional (multiple zones) for maximum nodes availability during zonal outages|[gke-policies/policy/node_pool_multi_zone.rego](https://github.com/google/gke-policy-automation/blob/main/gke-policies/policy/node_pool_multi_zone.rego)|
|Availability|Use Node Auto-Repair|GKE node pools should have Node Auto-Repair enabled to configure Kubernetes Engine|[gke-policies/policy/node_pool_autorepair.rego](https://github.com/google/gke-policy-automation/blob/main/gke-policies/policy/node_pool_autorepair.rego)|
|Maintenance|Cloud Monitoring and Logging|GKE cluster should use Cloud Logging and Monitoring|[gke-policies/policy/monitoring_and_logging.rego](https://github.com/google/gke-policy-automation/blob/main/gke-policies/policy/monitoring_and_logging.rego)|
|Management|Enable binary authorization in the cluster|GKE cluster should enable for deploy-time security control that ensures only trusted container images are deployed to gain tighter control over your container environment.|[gke-policies/policy/cluster_binary_authorization.rego](https://github.com/google/gke-policy-automation/blob/main/gke-policies/policy/cluster_binary_authorization.rego)|
|Management|GKE VPC-native cluster|GKE cluster nodepool should be VPC-native as per our best-practices|[gke-policies/policy/vpc_native_cluster.rego](https://github.com/google/gke-policy-automation/blob/main/gke-policies/policy/vpc_native_cluster.rego)|
|Management|Receive updates about new GKE versions|GKE cluster should be proactively receive updates about GKE upgrades and GKE versions|[gke-policies/policy/cluster_receive_updates.rego](https://github.com/google/gke-policy-automation/blob/main/gke-policies/policy/cluster_receive_updates.rego)|
|Management|Schedule maintenance windows and exclusions|GKE cluster should schedule maintenance windows and exclusions to upgrade predictability and to align updates with off-peak business hours.|[gke-policies/policy/cluster_maintenance_window.rego](https://github.com/google/gke-policy-automation/blob/main/gke-policies/policy/cluster_maintenance_window.rego)|
|Scalability|GKE ConfigMaps Limit|GKE ConfigMaps Limit|[gke-policies/scalability/limits_configmaps.rego](https://github.com/google/gke-policy-automation/blob/main/gke-policies/scalability/limits_configmaps.rego)|
|Scalability|GKE HPAs Limit|GKE HPAs Limit|[gke-policies/scalability/limits_hpas.rego](https://github.com/google/gke-policy-automation/blob/main/gke-policies/scalability/limits_hpas.rego)|
|Scalability|GKE L4 ILB Subsetting|GKE cluster should use GKE L4 ILB Subsetting if nodes > 250|[gke-policies/policy/ilb_subsetting.rego](https://github.com/google/gke-policy-automation/blob/main/gke-policies/policy/ilb_subsetting.rego)|
|Scalability|GKE Unused HPAs Limit|GKE Unused HPAs Limit|[gke-policies/scalability/limits_unused_hpas.rego](https://github.com/google/gke-policy-automation/blob/main/gke-policies/scalability/limits_unused_hpas.rego)|
|Scalability|GKE node local DNS cache|GKE cluster should use node local DNS cache|[gke-policies/policy/node_local_dns_cache.rego](https://github.com/google/gke-policy-automation/blob/main/gke-policies/policy/node_local_dns_cache.rego)|
|Scalability|Use node pool autoscaling|GKE node pools should have autoscaling configured to proper resize nodes according to traffic|[gke-policies/policy/node_pool_autoscaling.rego](https://github.com/google/gke-policy-automation/blob/main/gke-policies/policy/node_pool_autoscaling.rego)|
|Security|Control Plane endpoint access|Control Plane endpoint access should be limited to authorized networks only|[gke-policies/policy/control_plane_access.rego](https://github.com/google/gke-policy-automation/blob/main/gke-policies/policy/control_plane_access.rego)|
|Security|Control Plane endpoint visibility|Control Plane endpoint should be locked from external access|[gke-policies/policy/control_plane_endpoint.rego](https://github.com/google/gke-policy-automation/blob/main/gke-policies/policy/control_plane_endpoint.rego)|
|Security|Enrollment in Release Channels|GKE cluster should be enrolled in release channels|[gke-policies/policy/cluster_release_channels.rego](https://github.com/google/gke-policy-automation/blob/main/gke-policies/policy/cluster_release_channels.rego)|
|Security|Forbid default Service Accounts in Node Auto-Provisioning|Node Auto-Provisioning configuration should not allow default Service Accounts|[gke-policies/policy/nap_forbid_default_sa.rego](https://github.com/google/gke-policy-automation/blob/main/gke-policies/policy/nap_forbid_default_sa.rego)|
|Security|Forbid default compute SA on node_pool|GKE node pools should have a dedicated sa with a restricted set of permissions|[gke-policies/policy/node_pool_forbid_default_sa.rego](https://github.com/google/gke-policy-automation/blob/main/gke-policies/policy/node_pool_forbid_default_sa.rego)|
|Security|GKE Network Policies engine|GKE cluster should have Network Policies or Dataplane V2 enabled|[gke-policies/policy/network_policies.rego](https://github.com/google/gke-policy-automation/blob/main/gke-policies/policy/network_policies.rego)|
|Security|GKE RBAC authorization|GKE cluster should use RBAC instead of legacy ABAC authorization|[gke-policies/policy/control_plane_disable_legacy_authorization.rego](https://github.com/google/gke-policy-automation/blob/main/gke-policies/policy/control_plane_disable_legacy_authorization.rego)|
|Security|GKE Shielded Nodes|GKE cluster should use shielded nodes|[gke-policies/policy/shielded_nodes.rego](https://github.com/google/gke-policy-automation/blob/main/gke-policies/policy/shielded_nodes.rego)|
|Security|GKE Workload Identity|GKE cluster should have Workload Identity enabled|[gke-policies/policy/workload_identity.rego](https://github.com/google/gke-policy-automation/blob/main/gke-policies/policy/workload_identity.rego)|
|Security|GKE private cluster|GKE cluster should be private to ensure network isolation|[gke-policies/policy/private_cluster.rego](https://github.com/google/gke-policy-automation/blob/main/gke-policies/policy/private_cluster.rego)|
|Security|Integrity monitoring on the nodes|GKE node pools should have integrity monitoring feature enabled to detect changes in a VM boot measurments|[gke-policies/policy/node_pool_integrity_monitoring.rego](https://github.com/google/gke-policy-automation/blob/main/gke-policies/policy/node_pool_integrity_monitoring.rego)|
|Security|Kubernetes secrets encryption|GKE cluster should use encryption for kubernetes application secrets|[gke-policies/policy/secret_encryption.rego](https://github.com/google/gke-policy-automation/blob/main/gke-policies/policy/secret_encryption.rego)|
|Security|Use Container-Optimized OS|GKE node pools should use Container-Optimized OS which is maintained by Google and optimized for running Docker containers with security and efficiency.|[gke-policies/policy/node_pool_use_cos.rego](https://github.com/google/gke-policy-automation/blob/main/gke-policies/policy/node_pool_use_cos.rego)|
|Security|Use Node Auto-Upgrade|GKE node pools should have Node Auto-Upgrade enabled to configure Kubernetes Engine|[gke-policies/policy/node_pool_autoupgrade.rego](https://github.com/google/gke-policy-automation/blob/main/gke-policies/policy/node_pool_autoupgrade.rego)|
