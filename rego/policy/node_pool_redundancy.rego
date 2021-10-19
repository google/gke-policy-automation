package gke.policy.node_pool_redundancy

name = "Node pool redundancy"
description = "GKE node pools should be regional for maximum availability of a node pool during zonal outages"
group = "Availability"

default valid = false

valid {
  count(violation) == 0
}

violation[msg] {
  some nodepool
  data.gke.rule.nodepool.location.zonal[nodepool]
  msg := sprintf("invalid locations for GKE node pool %q (not regional)", [nodepool.name])
}
