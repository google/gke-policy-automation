<!-- markdownlint-disable MD041 -->
<img src="assets/gke-policy-automation-logo.png" alt="GKE Policy Automation logo"
title="GKE Policy Automation" align="left" height="70" />
<!-- markdownlint-enable MD041 -->

# GKE Policy Automation

This repository contains the tool and the [policy library](./gke-policies-v2) for validating [GKE](https://cloud.google.com/kubernetes-engine)
clusters against configuration [best practices](#checking-best-practices)
and [scalability limits](#checking-scalability-limits).

[![Build](https://github.com/google/gke-policy-automation/actions/workflows/build.yml/badge.svg)](https://github.com/google/gke-policy-automation/actions/workflows/build.yml)
[![Policy tests](https://github.com/google/gke-policy-automation/actions/workflows/policy-test.yml/badge.svg)](https://github.com/google/gke-policy-automation/actions/workflows/policy-test.yml)
[![Version](https://img.shields.io/github/v/release/google/gke-policy-automation?label=version)](https://img.shields.io/github/v/release/google/gke-policy-automation?label=version)
[![Go Report Card](https://goreportcard.com/badge/github.com/google/gke-policy-automation)](https://goreportcard.com/report/github.com/google/gke-policy-automation)
[![GoDoc](https://godoc.org/github.com/google/gke-policy-automation?status.svg)](https://godoc.org/github.com/google/gke-policy-automation)
![GitHub](https://img.shields.io/github/license/google/gke-policy-automation)

![GKE Policy Automation Demo](./assets/gke-policy-automation-demo.gif)

Note: this is not an officially supported Google product.

---

## Table of Contents

* [Installation](#installation)
* [Usage](#usage)
  * [Checking best practices](#checking-best-practices)
  * [Checking scalability limits](#checking-scalability-limits) (**New feature!**)
  * [Common check options](#common-check-options)
  * [Defining inputs](#defining-inputs)
  * [Defining outputs](#defining-outputs)
  * [Authentication](#authentication)
  * [Serverless execution](#serverless-execution)
* [Contributing](#contributing)
* [License](#license)

## Installation

### Container image

The container images with GKE Policy Automation tool are hosted on `ghcr.io`. Check the [packages page](https://github.com/google/gke-policy-automation/pkgs/container/gke-policy-automation)
for a list of all tags and versions.

```sh
docker pull ghcr.io/google/gke-policy-automation:latest
docker run --rm ghcr.io/google/gke-policy-automation check \
-project my-project -location europe-west2 -name my-cluster
```

### Binary

Binaries for Linux, Windows and Mac are available as tarballs in the
[release page](https://github.com/google/gke-policy-automation/releases).

### Source code

Go [v1.20](https://go.dev/doc/install) or newer is required. Check the [development guide](./DEVELOPMENT.md)
for more details.

```sh
git clone https://github.com/google/gke-policy-automation.git
cd gke-policy-automation
make build
./gke-policy check \
--project my-project --location europe-west2 --name my-cluster
```

## Usage

**Full user guide**: [GKE Policy Automation User Guide](./docs/user-guide.md).

### Checking best practices

The configuration best practices check validates GKE clusters against the set of
GKE configuration policies.

```sh
./gke-policy check \
--project my-project --location europe-west2 --name my-cluster
```

### Checking scalability limits

The scalability limits check validates GKE clusters against the GKE quotas and limits.
The tool will report violations when the current values will cross the certain thresholds.

```sh
./gke-policy check scalability \
--project my-project --location europe-west2 --name my-cluster
```

**NOTE**: you need to run `kube-state-metrics` to export cluster metrics to use cluster scalability
limits check. Refer to the [kube-state-metrics installation & configuration guide](./docs/kube-state-metrics.md)
for more details.

The tool assumes that metrics are available in Cloud Monitoring, i.e. in a result of
[Google Cloud Managed Service for Prometheus](https://cloud.google.com/stackdriver/docs/managed-prometheus)
based metrics collection. If self managed Prometheus collection is used, be sure to:

* Configure Prometheus scraping for `kube-state-metrics` using `PodMonitor` / `ServiceMonitor` and
 corresponding annotations, i.e. `prometheus.io/scrape`
* Configure custom Prometheus API server address in a tool

  * Prepare `config.yaml`:

     ```yaml
     inputs:
       metricsAPI:
         enabled: true
         address: http://my-prometheus-svc:8080 # Prometheus server API endpoint
         username: user   # username for basic authentication (optional)
         password: secret # password for basic authentication (optional)
     ```

  * Run `./gke-policy check scalability -c config.yaml`

### Common check options

The common options apply to all types of check commands.

#### Selecting multiple clusters

Check multiple GKE clusters using the config file.

```sh
./gke-policy check -c config.yaml
```

The `config.yaml` file:

```yaml
clusters:
  - name: prod-central
    project: my-project-one
    location: europe-central2
  - id: projects/my-project-two/locations/europe-west2/clusters/prod-west
```

#### Using cluster discovery

Check multiple clusters by discovering them in a selected GCP projects, folders or in the entire organization
using [Cloud Asset Inventory](https://cloud.google.com/asset-inventory) and configuration file.

```sh
./gke-policy check -c config.yaml
```

The `config.yaml` file:

```yaml
clusterDiscovery:
  enabled: true
  organization: "123456789012"
```

It is possible to use cluster discovery on a given project using command line flags only:

```sh
./gke-policy check --discovery -p my-project-id
```

### Defining inputs

Data for cluster validation can be retrieved from multiple data sources,
eg. GKE API, Cloud Monitoring API or local JSON file exported from GKE API.
For best practices checks GKE API is enabled by default,
and for scalability checks, metrics API is enabled as well.
Check [Inputs user guide](./docs/user-guide.md#inputs) for more details.

Example:

* Metrics API input from Cloud Monitoring configured in dedicated project
and other values set with defaults for scalability check

```yaml
inputs:
  gkeAPI:
    enabled: true
  gkeLocal:
    enabled: false
    file:
  metricsAPI:
    enabled: true
    project: sample-project
    metrics:
```

### Defining outputs

The cluster validation results can be published to multiple outputs, including JSON file, Pub/Sub topic,
Cloud Storage bucket or Security Command Center. Check [Outputs user guide](./docs/user-guide.md#outputs)
for more details.

Examples:

* JSON file output with command line flags

  ```sh
  ./gke-policy check \
  --project my-project --location europe-west2 --name my-cluster \
  --out-file output.json
  ```

* All outputs enabled in a configuration file

  ```yaml
  clusters:
    - name: my-cluster
      project: my-project
      location: europe-west2
  outputs:
    - file: output.json
    - pubsub:
        topic: Test
        project: my-pubsub-project
    - cloudStorage:
        bucket: bucket-name
        path: path/to/write
    - securityCommandCenter:
        organization: "153963171798"
  ```

#### Custom Policy repository

Specify custom repository with the GKE cluster best practices and check the cluster against them.

* Custom policies source with command line flags

  ```sh
  ./gke-policy check \
  --project my-project --location europe-west2 --name my-cluster \
  --git-policy-repo "https://github.com/google/gke-policy-automation" \
  --git-policy-branch "main" \
  --git-policy-dir "gke-policies-v2"
  ```

* Custom policies source with configuration file

  ```sh
  ./gke-policy check -c config.yaml
  ```

  The `config.yaml` file:

  ```yaml
  clusters:
    - name: my-cluster
      project: my-project
      location: europe-west2
  policies:
    - repository: https://domain.com/your/custom/repository
      branch: main
      directory: gke-policies-v2
  ```

### Authentication

The tool is fetching GKE cluster details using GCP APIs. The [application default credentials](https://cloud.google.com/docs/authentication/production)
are used by default.

* When running the tool in GCP environment, the tool will use the [attached service account](https://cloud.google.com/iam/docs/impersonating-service-accounts#attaching-to-resources)
by default
* When running locally, use `gcloud auth application-default login` command to get application
default credentials
* To use credentials from service account key file pass `--creds` parameter with a path to the file.

The minimum required IAM role is `roles/container.clusterViewer`
on a cluster projects. Additional roles may be needed, depending on configured outputs
\- check [authentication section](./docs/user-guide.md#authentication) in the user guide.

### Serverless execution

The GKE Policy Automation tool can be executed in a serverless way to perform automatic evaluations
of a clusters running in your organization. Please check our [reference Terraform Solution](./terraform/README.md)
that leverages GCP serverless solutions including Cloud Scheduler and Cloud Run.

## Contributing

Please check out [Contributing](./CONTRIBUTING.md) and [Code of Conduct](./docs/code-of-conduct.md)
docs before contributing.

### Development

Please check [GKE Policy Automation development](./DEVELOPMENT.md) for guides on building and developing
the application.

### Policy authoring

Please check [GKE Policy authoring guide](./gke-policies-v2/README.md) for guides on authoring REGO rules
for GKE Policy Automation.

## License

[Apache License 2.0](LICENSE)
