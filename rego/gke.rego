package gke

policies[policy_name] {
    policy := data.gke.policy[policy_name]
}

violations[policy_name] {
    policy := data.gke.policy[policy_name]
    policy.valid != true
}

policies_data[policy_data] {
    some policy_name
    policy_data = {"name": policy_name, "data": data.gke.policy[policy_name]} 
}