# METADATA
# title: Control Plane redundancy
# description: GKE cluster should be regional for maximum availability of control plane during upgrades and zonal outages
# custom:
#   group: Availability
package gke.policy.control_plane_redundancy

default valid = false

valid {
  count(violation) == 0
}

violation[msg] {
  some location
  data.gke.rule.location.zonal[location]
  msg := sprintf("invalid GKE Control plane location %q (not regional)", [location])
}
