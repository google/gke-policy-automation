# METADATA
# title: Control Plane endpoint visibility
# description: Control Plane endpoint should be locked from external access
# custom:
#   group: Security
package gke.policy.control_plane_endpoint

default valid = false

valid {
  count(violation) == 0
}

violation[msg] {
  not input.private_cluster_config.enable_private_endpoint
  msg := "GKE cluster has not enabled private endpoint" 
}
