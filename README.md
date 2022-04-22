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

---

## Table of Contents

* [Installation](#installation)
* [Usage](#usage)
* [Contributing](#contributing)
* [License](#license)

## Installation

### Container image

```sh
docker pull ghcr.io/google/gke-policy-automation:latest
docker run --rm ghcr.io/google/gke-policy-automation cluster review \
-project my-project -location europe-west2 -name my-cluster
```

### Binary

Binaries for Linux, Windows and Mac are available as tarballs in the
[release page](https://github.com/google/gke-policy-automation/releases).

### Source code

Go [v1.17](https://go.dev/doc/install) or newer is required. Check [GKE Policy Automation development](./DEVELOPMENT.md)
for more guides on building and developing the application.

```sh
git clone https://github.com/google/gke-policy-automation.git
cd gke-policy-automation
make build
./gke-policy cluster review \
-project my-project -location europe-west2 -name my-cluster
```

## Usage

**Full user guide**: please refer to [GKE Policy Automation user guide](./docs/user-guide.md).

### Checking the cluster

### Checking multiple clusters

### Custom Policy repository

### Authentication

The tool is using [application default credentials](https://cloud.google.com/docs/authentication/production)
by default.

* When running the tool in GCP environment, the tool will use the [attached service account](https://cloud.google.com/iam/docs/impersonating-service-accounts#attaching-to-resources)
by default
* When running locally, use `gcloud auth application-default login` command to obtain application
default credentials

It is also possible to use credentials from service account key file by passing `--creds` parameter
with a path to the file.

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
