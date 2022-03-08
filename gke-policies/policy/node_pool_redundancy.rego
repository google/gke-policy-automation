# METADATA
# title: Control Plane redundancy
# description: GKE node pools should be regional for maximum availability of a node pool during zonal outages
# custom:
#   group: Availability
package gke.policy.node_pool_redundancy

default valid = false

valid {
  count(violation) == 0
}

violation[msg] {
  some nodepool
  data.gke.rule.nodepool.location.zonal[nodepool]
  msg := sprintf("invalid locations for GKE node pool %q (not regional)", [nodepool.name])
}
