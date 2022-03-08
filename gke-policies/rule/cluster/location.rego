package gke.rule.cluster.location

regional[location] {
    location := input.location
    regex.match("^[^-]+-[^-]+$", location)
}

zonal[location] {
    location := input.location
    regex.match("^[^-]+-[^-]+-[^-]+$", location)
}