# GKE Policy authoring guide

The GKE Policy Automation tool provides ready [set of GKE policies](./) that cover Google PSO best practices
and recommendations.

**The following guide is for GKE policy authors and contributors**. The GKE Policies need to follow
the below rules in order to be successfully compiled and evaluated by the GKE Policy Automation tool.

1. [GKE Policy structure overview](#gke-policy-structure-overview)
2. [GKE Policy metadata](#gke-policy-metadata)
3. [GKE Policy package](#gke-policy-package)
4. [GKE Policy rules](#gke-policy-rules)
5. [GKE Policy tests](#gke-policy-tests)
6. [GKE Policy documentation](#gke-policy-documentation)

---

## Useful links

* [Rego policy language overview](https://www.openpolicyagent.org/docs/latest/policy-language/)
* [Rego policy reference](https://www.openpolicyagent.org/docs/latest/policy-reference/)
* [Rego policy testing](https://www.openpolicyagent.org/docs/latest/policy-testing/)

## GKE Policy structure overview

The GKE policies are ASCII files with policy definitions written in [Rego language](https://www.openpolicyagent.org/docs/latest/policy-language/).

The GKE Policy Automation tool can evaluate policies from local directory or directory
from the remote GIT repository.

### Policy directory tree

GKE policy files can be organized into directories. Although it is not required to use directories
at all, doing so helps to group Rego files of similar purpose. Typically directory structure is
somehow related to the [Rego packages](https://www.openpolicyagent.org/docs/latest/policy-language/#packages)
defined in policy files.

Below is a simple directory structure for GKE policy files:

```sh
/policy_directory
|-- policy
|   |-- first_gke_policy.rego
|   |-- first_gke_policy_test.rego
|   |-- another_gke_policy.rego
|   |-- another_gke_policy_test.rego
|   |-- ...
|-- rule
|   |-- some_rule.rego 
|   |-- another_rule.rego
|   | ...
```

* `policy_directory` -  root directory with all GKE policy Rego files
* `policy` - subdirectory that groups GKE policies (i.e. `gke.policy.xxxx` packages)
* `rule` - subdirectory that groups reusable rules (i.e. `gke.rule.xxxx` packages)

### Policy file structure

GKE policies are written in [Rego language](https://www.openpolicyagent.org/docs/latest/policy-language/).
Each GKE Policy is defined in an individual file and within individual Rego package. The valid
GKE Policy file has also given structure:

* Metadata section on a package level
* Package definition with a name recognized by the tool
* One `valid` and one or more `violation` rules

More details will be covered in a following sections of this document.
Below is an example of a valid GKE Policy file.

```rego
# METADATA
# title: Control Plane endpoint access
# description: Control Plane endpoint access should be limited to authorized networks only
# custom:
#   group: Security
package gke.policy.control_plane_access

default valid = false

valid {
  count(violation) == 0
}

violation[msg] {
  not input.master_authorized_networks_config.enabled
  msg := "GKE cluster has not enabled master authorized networks configuration" 
}

```

## GKE Policy metadata

GKE Policies use [OPA Annotations](https://www.openpolicyagent.org/docs/latest/annotations/#annotations)
to specify policy metadata. The required metadata annotations for GKE policy:

* `title` - human readable name of a policy
* `description` - more detailed description of a policy
* `custom.group` - name of group of a policy for policy grouping / categorization

The annotations should be put on a package scope in a rego file.

## GKE Policy package

Each GKE Policy is defined within individual Rego package.
The package name should start with `gke.policy.` and should match policy name. By Rego definition,
the package name should only contain string operands.

Examples:

* `gke.policy.control_plane_access` for Control Plane access policy
* `gke.policy.private_cluster` for Private Cluster policy

## GKE Policy rules

Each GKE Policy should have following rules:

* `valid` - The rule determines if a given policy is valid or violated. It should generate either
`true` or `false`.
* `violation` - The rule determines violation for a given policy. It should generate string with a
violation description. There can be multiple `violation` rules per one policy if needed.

GKE Policy rules are evaluated against Cluster data returned by Get Cluster gRPC API Call.
Therefore, the `input` document has a protobuf [GKE Cluster model](https://pkg.go.dev/google.golang.org/genproto/googleapis/container/v1#Cluster).

## GKE Policy tests

Each GKE Policy should be covered with unit tests. OPA Rego provides
[testing framework](https://www.openpolicyagent.org/docs/latest/policy-testing/) for that.

* Each GKE Policy should have individual test file
* Test files should be stored in same directory as policies
* Test files should be named same as given policy file and suffixed with `_test.rego`
* Test rules should be in same package as given policy rules

## GKE Policy documentation

The GKE Policy Automation tool can generate markdown documentation
for the policies fetched from a given source. This can be used i.e. to update list of policies
in a [policy library README](README.md).

Example:

```sh
./gke-policy generate policy-docs --local-policy-dir ./gke-policies-v2 -f generated-policy-docs.md
```
