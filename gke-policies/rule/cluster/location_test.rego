package gke.rule.cluster.location

test_regional {
    location := "europe-central2"
    regional(location)
    not zonal(location)
}

test_zonal {
    location := "europe-central2-a"
    zonal(location)
    not regional(location)
}

test_not_regional_nor_zonal {
    location := "test"
    not regional(location)
    not zonal(location)
}
