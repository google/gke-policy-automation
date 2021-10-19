package gke.policy.control_plane_access

name = "Control Plane endpoint access"
description = "Control Plane endpoint access should be limited to authorized networks only"
group = "Security: isolation"

default valid = false

valid {
  count(violation) == 0
}

violation[msg] {
  not input.master_authorized_networks_config.enabled
  msg := "GKE cluster has not enabled master authorized networks configuration" 
}
