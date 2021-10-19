package gke.policy.private_cluster

name = "GKE private cluster"
description = "GKE cluster should be private to ensure network isolation"
group = "Security: isolation"

default valid = false

valid {
  count(violation) == 0
}

violation[msg] {
  not input.private_cluster_config.enable_private_nodes
  msg := "GKE cluster has not enabled private nodes"
}
