# Changelog

## 1.4.4 (Dec 13, 2024)

IMPROVEMENTS:

* Upgraded direct and transitive dependencies [#218](https://github.com/google/gke-policy-automation/pull/218)
[#219](https://github.com/google/gke-policy-automation/pull/219)

## 1.4.3 (Sep 20, 2024)

IMPROVEMENTS:

* Upgraded direct and transitive dependencies [#217](https://github.com/google/gke-policy-automation/pull/217)

## 1.4.2 (Sep 20, 2024)

IMPROVEMENTS:

* Upgraded direct and transitive dependencies [#216](https://github.com/google/gke-policy-automation/pull/216)

## 1.4.1 (Aug 8, 2024)

IMPROVEMENTS:

* Terraform with cloud run job and remote repo [#214](https://github.com/google/gke-policy-automation/pull/214)
* Upgraded direct and transitive dependencies [#213](https://github.com/google/gke-policy-automation/pull/213)

## 1.4.0 (May 6, 2024)

IMPROVEMENTS:

* Unified policies metadata and new console output [#208](https://github.com/google/gke-policy-automation/pull/208)

BUG FIXES:

* Policy recommendations not present in SCC finding summary [#206](https://github.com/google/gke-policy-automation/pull/206)

## 1.3.4 (Dec 29, 2023)

IMPROVEMENTS:

* Upgraded all direct and transitive dependencies

## 1.3.3 (Nov 8, 2023)

FEATURES:

* Krew based installation [#105](https://github.com/google/gke-policy-automation/issues/105)

NEW POLICIES:

* GKE intranode visibility [#196](https://github.com/google/gke-policy-automation/pull/196)
* Control plane user basic authentication [#197](https://github.com/google/gke-policy-automation/pull/197)
* Control plane user certificate authentication [#197](https://github.com/google/gke-policy-automation/pull/197)
* Customer-Managed Encryption Keys for persistent disks [#197](https://github.com/google/gke-policy-automation/pull/197)
* Enable Security Posture dashboard [#197](https://github.com/google/gke-policy-automation/pull/197)
* Enable Workload vulnerability scanning [#197](https://github.com/google/gke-policy-automation/pull/197)

IMPROVEMENTS:

* Upgraded direct and indirect dependencies [#195](https://github.com/google/gke-policy-automation/pull/195)
* Adjusted all policies to GKE CIS version 1.4 benchmark [#197](https://github.com/google/gke-policy-automation/pull/197)
* Added Regal for linting Rego [#194](https://github.com/google/gke-policy-automation/pull/194)

BUG FIXES:

* Policy `node_pool_use_cos` should not fail on windows node pools [#198](https://github.com/google/gke-policy-automation/pull/198)

## 1.3.2 (Aug 10, 2023)

IMPROVEMENTS:

* Upgraded direct and indirect dependencies [#192](https://github.com/google/gke-policy-automation/pull/192)
* New layout of generated policy documentation [#191](https://github.com/google/gke-policy-automation/pull/191)

BUG FIXES:

* Added anchors to cluster asset regex for security [#190](https://github.com/google/gke-policy-automation/pull/190)

## 1.3.1 (Jan 1, 2023)

IMPROVEMENTS:

* Upgraded Go to 1.20
* Upgraded all direct and indirect dependencies

BUG FIXES:

* Upgraded CIRCL indirect dependency to v1.3.3 to fix security issues with error-handling
on rand readers (CVE-2023-1732)

## 1.3.0 (Mar 14, 2023)

FEATURES:

* GKE Scalability checks based on metrics from kube-state-metrics [#179](https://github.com/google/gke-policy-automation/pull/179)
* Introduced external URI and recommendations to the policy model and outputs [#131](https://github.com/google/gke-policy-automation/pull/111),
  [#141](https://github.com/google/gke-policy-automation/pull/141)

IMPROVEMENTS:

* Introduced modularized inputs concept [#127](https://github.com/google/gke-policy-automation/issues/127)
* Added PromQL integration with a Cloud Monitoring and self hosted Prometheus for metrics ingestion [#132](https://github.com/google/gke-policy-automation/pull/132),
  [#178](https://github.com/google/gke-policy-automation/pull/178)
* Security Command Center output performance improvements [#151](https://github.com/google/gke-policy-automation/pull/151)
* Logs from logger can be stored in a files and in JSON format [#155](https://github.com/google/gke-policy-automation/pull/155)
* Adding -json flag to output results to stdout in JSON format [#147](https://github.com/google/gke-policy-automation/pull/147)

BUG FIXES:

* Fixed variable types in Terraform code [#150](https://github.com/google/gke-policy-automation/pull/150)

## 1.2.2 (Nov 8, 2022)

IMPROVEMENTS:

* Add support for JSON output to stdout [#129](https://github.com/google/gke-policy-automation/issues/129)

## 1.2.1 (Aug 17, 2022)

IMPROVEMENTS:

* Improved efficiency of K8S resources fetching [#107](https://github.com/google/gke-policy-automation/pull/107)
* Updated policy docs generator [#109](https://github.com/google/gke-policy-automation/pull/109)

BUG FIXES:

* Tool should not fail on a discovered cluster that does not exist [#113](https://github.com/google/gke-policy-automation/issues/113)
* Failed cluster discovery was not returning an error [#104](https://github.com/google/gke-policy-automation/pull/104)

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
