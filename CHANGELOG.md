# Changelog

## 1.2.0 (Jul 22, 2022)

FEATURES:

* Security Command Center output [#100](https://github.com/google/gke-policy-automation/pull/100)

IMPROVEMENTS:

* Cluster discovery triggered from CLI [#92](https://github.com/google/gke-policy-automation/pull/92)
* New console output, cluster evaluations are now policy oriented [#90](https://github.com/google/gke-policy-automation/pull/90)
* Tool can generate markdown documentation from policies [#86](https://github.com/google/gke-policy-automation/pull/86)

BUG FIXES:

* Cluster discovery skipped zonal clusters due to name pattern mismatch[#91](https://github.com/google/gke-policy-automation/pull/91)

## 1.1.0 (Jul 1, 2022)

FEATURES:

* Introduced check commands and multiple packages handling [#89](https://github.com/google/gke-policy-automation/pull/89)
* Use of K8S resources data in REGO policies [#61](https://github.com/google/gke-policy-automation/issues/61)
* Policy filtering logic with policy names and groups [#69](https://github.com/google/gke-policy-automation/issues/69)

## 1.0.1 (Jun 2, 2022)

BUG FIXES:

* Bumped dependency versions, including yaml.3 [#84](https://github.com/google/gke-policy-automation/pull/84)
* Bump github.com/open-policy-agent/opa from 0.38.1 to 0.40.0 [#83](https://github.com/google/gke-policy-automation/pull/83)

## 1.0.0 (May 25, 2022)

FEATURES:

* Terraform serverless solution [#75](https://github.com/google/gke-policy-automation/pull/75)
* Cluster discovery mechanism [#59](https://github.com/google/gke-policy-automation/pull/59)
* Cluster review with cluster data from a file [#50](https://github.com/google/gke-policy-automation/pull/50)
* Command that prints raw cluster data [#37](https://github.com/google/gke-policy-automation/pull/37)
* Policy Evaluation result JSON output to Cloud Storage [#34](https://github.com/google/gke-policy-automation/issues/34)
* Policy Evaluation result JSON output to Pub/Sub [#33](https://github.com/google/gke-policy-automation/issues/33)
* Policy Evaluation result JSON output to local file [#5](https://github.com/google/gke-policy-automation/issues/5)

IMPROVEMENTS:

* Adjusted exit code on errors and improved logging[#81](https://github.com/google/gke-policy-automation/pull/81)
* Custom user-agent in GCP API calls [#78](https://github.com/google/gke-policy-automation/pull/78)
* Default GIT policy source params are set in consistent way [#43](https://github.com/google/gke-policy-automation/pull/43)

BUG FIXES:

* Tool fetches cluster details even if there are no policies [#45](https://github.com/google/gke-policy-automation/issues/45)

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
