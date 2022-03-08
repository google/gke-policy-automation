# METADATA
# title: Test of invalid policy
# description: Test of invalid policy
# custom:
#   group: Test
package gke.policy.errored_test_policy

valid {
  count(violation) == 0
}

violation[msg] {
  msg := "GKE cluster has not enabled private endpoint" 
}
