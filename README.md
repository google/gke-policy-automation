# GKE Policy Automation

This is not an officially supported Google product.

This repository contains the tool and the [policy library](./gke-policies) for validating [GKE](https://cloud.google.com/kubernetes-engine)
clusters against configuration best practices.

[![Build](https://github.com/google/gke-policy-automation/actions/workflows/build.yml/badge.svg)](https://github.com/google/gke-policy-automation/actions/workflows/build.yml)
[![Policy tests](https://github.com/google/gke-policy-automation/actions/workflows/policy-test.yml/badge.svg)](https://github.com/google/gke-policy-automation/actions/workflows/policy-test.yml)
[![Version](https://img.shields.io/github/v/release/google/gke-policy-automation?label=version)](https://img.shields.io/github/v/release/google/gke-policy-automation?label=version)
[![Go Report Card](https://goreportcard.com/badge/github.com/google/gke-policy-automation)](https://goreportcard.com/report/github.com/google/gke-policy-automation)
[![GoDoc](https://godoc.org/github.com/google/gke-policy-automation?status.svg)](https://godoc.org/github.com/google/gke-policy-automation)
![GitHub](https://img.shields.io/github/license/google/gke-policy-automation)

![GKE Policy Automation Demo](./assets/gke-policy-automation-demo.gif)

---

## Table of Contents

* [Installation](#installation)
* [Usage](#usage)
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

Go [v1.17](https://go.dev/doc/install) or newer is required. Check the [development guide](./DEVELOPMENT.md)
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

### Checking the cluster

Check the GKE cluster against the default set of best practices with command line flags.

```sh
./gke-policy check \
--project my-project --location europe-west2 --name my-cluster
```

### Checking multiple clusters

Check multiple GKE clusters against the default set of best practices with a config file.

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

### Discovering and checking multiple clusters

Discover clusters in a selected GCP projects, folders or in the entire organization using
[Cloud Asset Inventory](https://cloud.google.com/asset-inventory) and check them against the default
set of best practices.

```sh
./gke-policy check -c config.yaml
```

The `config.yaml` file:

```yaml
clusterDiscovery:
  enabled: true
  organization: "123456789012"
```

### Defining outputs

Specify the desired outputs for validation results.

* JSON file output with command line flags

  ```sh
  ./gke-policy check \
  --project my-project --location europe-west2 --name my-cluster \
  --out-file output.json
  ```

* JSON file, GCS bucket or PubSub topic with configuration file

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
  ```

### Custom Policy repository

Specify custom repository with the GKE cluster best practices and check the cluster against them.

* Custom policies source with command line flags

  ```sh
  ./gke-policy check \
  --project my-project --location europe-west2 --name my-cluster \
  --git-policy-repo "https://github.com/google/gke-policy-automation" \
  --git-policy-branch "main" \
  --git-policy-dir "gke-policies"
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
    - repository: https://github.com/google/gke-policy-automation
      branch: main
      directory: gke-policies
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
on a cluster projects. Additional roles may be needed, depending on configured [outputs](./docs/user-guide.md#outputs).

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

Please check [GKE Policy authoring guide](./gke-policies/README.md) for guides on authoring REGO rules
for GKE Policy Automation.

## License

[Apache License 2.0](LICENSE)
