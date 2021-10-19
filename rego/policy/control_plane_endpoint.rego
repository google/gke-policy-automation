package gke.policy.control_plane_endpoint

name = "Control Plane endpoint visibility"
description = "Control Plane endpoint should be locked from external access"
group = "Security: isolation"

default valid = false

valid {
  count(violation) == 0
}

violation[msg] {
  not input.private_cluster_config.enable_private_endpoint
  msg := "GKE cluster has not enabled private endpoint" 
}
