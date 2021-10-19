package gke.rule.nodepool.location

regional[nodepool] {
    nodepool := input.node_pools[_]
    count(nodepool.locations) > 1
}

zonal[nodepool] {
    nodepool := input.node_pools[_]
    count(nodepool.locations) < 2
}