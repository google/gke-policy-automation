# Changelog

## 0.2.0 (Unreleased)

## 0.1.0 (April 15, 2022)

FEATURES:

* Cluster JSON data print command ([#37](https://github.com/google/gke-policy-automation/pull/37))
* Policy check command ([#7](https://github.com/google/gke-policy-automation/issues/7))

IMPROVEMENTS:

* Mandatory params check and color fix ([#26](https://github.com/google/gke-policy-automation/pull/26))

BUG FIXES:

* Specifying multiple clusters in `config.yaml` causes panic ([#27](https://github.com/google/gke-policy-automation/issues/27))
* Specifying --local-policy-dir CLI flag is not stopping from reading default GIT repo bug ([#21](https://github.com/google/gke-policy-automation/issues/21))
* Missing configuration parameters should cause tool to fail fast ([#20](https://github.com/google/gke-policy-automation/issues/20))

## 0.0.1 (March 30, 2022)

NOTES:

* initial version of the `gke-policy` tool after migration from PoC project

FEATURES:

* `gke-policy cluster review` command validates GKE clusters against best practices described
with REGO policies
