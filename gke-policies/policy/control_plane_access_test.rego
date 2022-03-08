package gke.policy.control_plane_access

test_authorized_networks_enabled {
    valid with input as {"master_authorized_networks_config": {"enabled":true}}
}

test_authorized_networks_disabled{
    not valid with input as {"master_authorized_networks_config": {"enabled":false}}
}