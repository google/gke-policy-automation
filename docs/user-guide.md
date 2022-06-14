# GKE Policy Automation User Guide

The GKE Policy Automation is a command line tool that validates GKE clusters against set of best practices.

---

## Table of Contents

* [Installation](#installation)
  * [Container image](#container-image)
  * [Binary](#binary)
  * [Source code](#source-code)
* [Authentication](#authentication)
  * [Required IAM roles](#required-iam-roles)
* [Cluster commands](#cluster-commands)
  * [Specifying single cluster](#specifying-single-cluster)
  * [Specifying multiple clusters](#specifying-multiple-clusters)
  * [Cluster discovery](#cluster-discovery)
  * [Reviewing clusters](#reviewing-clusters)
  * [Printing cluster data](#printing-cluster-data)
* [Policy Commands](#policy-commands)
  * [Specifying GIT policy source](#specifying-git-policy-source)
  * [Specifying local policy source](#specifying-local-policy-source)
  * [Validating policies](#validating-policies)
* [Outputs](#outputs)
  * [Local JSON file](#local-json-file)
  * [Cloud Storage bucket](#cloud-storage-bucket)
  * [Pub/Sub topic](#pubsub-topic)
* [Serverless execution](#serverless-execution)
* [Silent mode](#silent-mode)
* [Configuration file](#configuration-file)
* [Debugging](#debugging)

## Installation

### Container image

The container images with GKE Policy Automation tool are hosted on `ghcr.io`. Check the [packages page](https://github.com/google/gke-policy-automation/pkgs/container/gke-policy-automation)
for a list of all tags and versions.

```sh
docker pull ghcr.io/google/gke-policy-automation:latest
docker run --rm ghcr.io/google/gke-policy-automation cluster review \
-project my-project -location europe-west2 -name my-cluster
```

### Binary

Binaries for Linux, Windows and Mac are available as tarballs in the
[release page](https://github.com/google/gke-policy-automation/releases).

### Source code

Go [v1.17](https://go.dev/doc/install) or newer is required. Check the [development guide](../DEVELOPMENT.md)
for more details.

```sh
git clone https://github.com/google/gke-policy-automation.git
cd gke-policy-automation
make build
./gke-policy cluster review \
--project my-project --location europe-west2 --name my-cluster
```

## Authentication

The tool is fetching GKE cluster details using GCP APIs. The [application default credentials](https://cloud.google.com/docs/authentication/production)
are used by default.

* When running the tool in GCP environment, the tool will use the [attached service account](https://cloud.google.com/iam/docs/impersonating-service-accounts#attaching-to-resources)
by default
* When running locally, use `gcloud auth application-default login` command to get application
default credentials
* To use credentials from service account key file pass `--creds` parameter with a path to the file.

### Required IAM roles

* The minimum required IAM role is `roles/container.clusterViewer` on a cluster projects
* For cluster discovery the `cloudasset.assets.searchAllResources` permission is needed on a target
projects, folders or organization. This permission is included, among others, in a `roles/cloudasset.viewer`
role
* For Cloud Storage output, the `roles/storage.objectCreator` role is needed on a target bucket
* For Pub/Sub output, the `roles/pubsub.publisher` role is needed on a target topic

## Cluster commands

The cluster commands perform operations in a context of a GKE clusters.

```sh
USAGE:
   gke-policy cluster command [command options] [arguments...]

COMMANDS:
   print    Print cluster api raw json data
   review   Evaluate policies against given GKE cluster
   help, h  Shows a list of commands or help for one command
```

### Specifying single cluster

The cluster details can be set using command line flags or in a [configuration file](#configuration-file).

* `--project` is a GCP project identifier to which cluster belong
* `--location` is a location of a cluster, either GCP zone or a GCP region
* `--name` is a cluster's name

When using configuration file, it is also possible to reference cluster using `id` attribute
that is combination of the above in a format:
`projects/<project>/locations/<location>/clusters/<name>`

### Specifying multiple clusters

Setting details of a multiple clusters is possible using [configuration file](#configuration-file) only.

The example `config.yaml` file with a three clusters:

```yaml
clusters:
  - name: prod-central
    project: my-project-one
    location: europe-central2
  - id: projects/my-project-two/locations/europe-west2/clusters/prod-west
  - name: prod-north
    project: my-project-three
    location: europe-north1
```

### Cluster discovery

The cluster discovery mechanism is leveraging [Cloud Asset Inventory](https://cloud.google.com/asset-inventory)
API to find GKE clusters in a given GCP projects, folders or in an entire organization. The cluster
discovery can be used in place of a fixed list of cluster identifiers.

Setting cluster discovery is possible using [configuration file](#configuration-file) only.

* cluster discovery can't be configured along with a list of clusters
* cluster discovery projects are referenced by the project identifiers
* cluster discovery folders are referenced by the folder numbers
* cluster discovery organization is referenced by the organization number

The example `config.yaml` file with a cluster discovery enabled on the selected projects and folders:

```yaml
clusterDiscovery:
  enabled: true
  projects:
    - project-one
    - project-two
    - project-three
  folders:
    - "123456789123"
    - "987654321098"
```

The example `config.yaml` file with a cluster discovery enabled on the entire organization:

```yaml
clusterDiscovery:
  enabled: true
  organization: "123456789012"
```

**NOTE**: it might take some time for a GKE clusters to appear in a Cloud Asset Inventory search results.

### Reviewing clusters

Run `./gke-policy cluster review` command to check one or more GKE clusters against the set of REGO policies.

By default, given cluster(s) will be reviewed against default set of GKE best practices fetched
from the GKE Policy Automation repository. The default policy source can be altered
by [specifying custom policy source](#specifying-git-policy-source) details.

Single cluster review example:

```sh
./gke-policy cluster review \
--project my-project --location europe-west2 --name my-cluster
```

Multiple clusters review example:

```sh
./gke-policy cluster review -c config.yaml
```

The `config.yaml` file:

```yaml
clusters:
  - name: prod-central
    project: my-project-one
    location: europe-central2
  - id: projects/my-project-two/locations/europe-west2/clusters/prod-west
```

### Printing cluster data

Run `./gke-policy cluster print` to dump GKE cluster data in a JSON format.

Example:

```sh
./gke-policy cluster print \
--project my-project --location europe-west2 --name my-cluster
```

## Policy commands

The policy commands perform operations in a context of a REGO policies.

```sh
USAGE:
   gke-policy policy command [command options] [arguments...]

COMMANDS:
   check    Validates policy files from defined source
   help, h  Shows a list of commands or help for one command
```

### Specifying GIT policy source

The custom GIT policy source can be specified with a command line flags or in a [configuration file](#configuration-file).

* `git-policy-repo` is a repository URL to clone from
* `git-policy-branch` is a name of a GIT branch to clone
* `git-policy-dir` is a directory within the GIT repository to search for policy files

The GKE Policy Automation tool scans for files with `rego` extension. Refer to the
[policy authoring guide](../gke-policies/README.md) for more details about policies for this tool.

Example of a cluster review command with a custom policy repository:

  ```sh
  ./gke-policy cluster review \
  --project my-project --location europe-west2 --name my-cluster \
  --git-policy-repo "https://github.com/google/gke-policy-automation" \
  --git-policy-branch "main" \
  --git-policy-dir "gke-policies"
  ```

*NOTE*: currently the tool does not support authentication for GIT policy repositories.

### Specifying local policy source

The local policy source directory can be specified with a command line flags or in a [configuration file](#configuration-file).

* `local-policy-dir` is a path to the local policy directory to search for policy files

### Validating policies

Run `./gke-policy policy check` to validate Rego policies from a given policy source.
The policies are validated against the Rego syntax.

Example:

```sh
./gke-policy policy check --local-policy-dir ./gke-policies
```

### Excluding policies

Specific policies or policy groups may be excluded during cluster review. Policy exclusion can only
be configured using a [configuration file](#configuration-file). The below example skips all rego
policies in the `Scalability` group as well as the specific policy
`gke.policy.cluster_binary_authorization`.

```yaml
policyExclusions:
  policies:
    - gke.policy.cluster_binary_authorization
  policyGroups:
    - Scalability
```

## Outputs

The GKE Policy Automation tool produces cluster validation results to the stderr, local JSON file,
file on a GCS bucket and Pub/Sub topic.

### Local JSON file

The validation results can be stored in the local file in a JSON format.
Local file output can be enabled using either command line flag or in a [configuration file](#configuration-file).

Example of enabling local file output in a command line:

```sh
  ./gke-policy cluster review \
  --project my-project --location europe-west2 --name my-cluster \
  --out-file my-cluster-results.json
```

Example of defining local file output using configuration file:

```yaml
clusters:
  - id: projects/my-project-two/locations/europe-west2/clusters/my-cluster
outputs:
  - file: my-cluster-results.json
```

### Cloud Storage bucket

The validation results can be stored in a JSON format as an object in Cloud Storage bucket.
Cloud storage output can be enabled using [configuration file](#configuration-file), example:

```yaml
clusters:
  - id: projects/my-project-two/locations/europe-west2/clusters/my-cluster
outputs:
  - cloudStorage:
      bucket: bucket
      path: path/to/write
```

The Cloud Storage output adds date-time prefix to the given path by default, so the reports from
subsequent checks are not overwritten. This behavior can be disabled by setting `skipDatePrefix`
option to `true`.

### Pub/Sub topic

The validation results can be pushed as a JSON message to the Pub/Sub topic.
Pub/Sub output can be enabled using [configuration file](#configuration-file), example:

```yaml
clusters:
  - id: projects/my-project-two/locations/europe-west2/clusters/my-cluster
outputs:
  - pubsub:
      topic: testTopic
      project: my-pubsub-project
```

## Serverless execution

The GKE Policy Automation tool can be executed in a serverless way to perform automatic evaluations
of a clusters running in your organization. Please check our [Reference Terraform Solution](../terraform/README.md)
that leverages GCP serverless solutions including Cloud Scheduler and Cloud Run.

## Silent mode

The GKE Policy Automation tool produces human readable output to the stderr. You can disable this
behavior by enabling silent mode with `-s` or `--silent` flag.

Using silent mode is useful for automated executions where logs are favoured over human readable output.
Note that
enabling silent mode is not stopping [detailed logging](#debugging) if that is configured.

Example of execution with silent mode and logging enabled:

```sh
GKE_POLICY_LOG=DEBUG ./gke-policy cluster review --silent \
--location europe-central2 --name prod-central --project my-project 
```

## Configuration file

Use `-c <config.yaml>` after the command to use configuration file instead of command line flags. Example:

```sh
./gke-policy cluster review -c config.yaml
```

The below example `config.yaml` shows all available configuration options.

```yaml
silent: true
clusters:
  - name: prod-central
    project: my-project-one
    location: europe-central2
  - id: projects/my-project-two/locations/europe-west2/clusters/prod-west
clusterDiscovery:
  enabled: true
  projects:
    - project-one
    - project-two
    - project-three
  folders:
    - "123456789123" #folder number
    - "987654321098"
  organization: "123456789012" #organization number
policies:
  - repository: https://github.com/google/gke-policy-automation
    branch: main
    directory: gke-policies
  - local: ./my-policies
policyExclusions:
  policies:
    - gke.policy.enable_ilb_subsetting
  policyGroups:
    - Scalability
outputs:
  - file: output-file.json
  - pubsub:
      topic: testTopic
      project: my-pubsub-project
  - cloudStorage:
      bucket: bucket-name
      path: path/to/write
      skipDatePrefix: true

```

## Debugging

Detailed logs can be enabled by setting the `GKE_POLICY_LOG` environment variable to one of supported
log level values. This will cause detailed logs to appear on stderr.

You can set `GKE_POLICY_LOG` to one of the log levels `TRACE`, `DEBUG`, `INFO`, `WARN`, `ERROR` to
change verbosity of the logs.

The file log output can be enabled by setting `GKE_POLICY_LOG_PATH` with a path to the specific file
to where logs will be appended. Note that even when `GKE_POLICY_LOG_PATH` is set, `GKE_POLICY_LOG`
must to set in order for logging to be enabled.

Below is an example of running the application with `DEBUG` logging enabled.

```sh
GKE_POLICY_LOG=DEBUG ./gke-policy cluster review \
--project my-project --location europe-west2 --name my-cluster
```
