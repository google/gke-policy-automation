# GKE Policy Automation

This is not an officially supported Google product.

This repository contains the tool and [policy library](./gke-policies) for validating selected [GKE](https://cloud.google.com/kubernetes-engine)
clusters against configuration best practices.

[![Build](https://github.com/google/gke-policy-automation/actions/workflows/build.yml/badge.svg)](https://github.com/google/gke-policy-automation/actions/workflows/build.yml)
[![Policy tests](https://github.com/google/gke-policy-automation/actions/workflows/policy-test.yml/badge.svg)](https://github.com/google/gke-policy-automation/actions/workflows/policy-test.yml)
[![Version](https://img.shields.io/github/v/release/google/gke-policy-automation?label=version)](https://img.shields.io/github/v/release/google/gke-policy-automation?label=version)
[![Go Report Card](https://goreportcard.com/badge/github.com/google/gke-policy-automation)](https://goreportcard.com/report/github.com/google/gke-policy-automation)
[![GoDoc](https://godoc.org/github.com/google/gke-policy-automation?status.svg)](https://godoc.org/github.com/google/gke-policy-automation)
![GitHub](https://img.shields.io/github/license/google/gke-policy-automation)

---

## Table of Contents

- [Install](#install)
- [Usage](#usage)
- [Test](#test)
- [Contributing](#contributing)
- [License](#license)

## Install

```sh
make
```

or

```sh
make build
```

## Usage

Tool can be used from command line with [gcloud CLI](https://cloud.google.com/sdk/docs/install) installed.
CLI can be previously authenticated with `gcloud auth application-default login` command, or credentials may be passed with `--creds` parameter.

Parameters for GKE cluster review can be provided as command parameters or via configuration .yaml file.

```sh
gke-policy [global options] command [command options] [arguments...]
```

For cluster review with manually provided parameters:

```sh
./gke-policy cluster review -p <GCP_PROJECT_ID> -n <CLUSTER_NAME> -l <CLUSTER_LOCATION>
```

and with .yaml file with format:

```yaml
silent: true
credentialsFile: ./test_credentials.json
clusters:
  - name: my-cluster
    project: my-project
    location: europe-central2
  - name: another
    project: my-project
    location: europe-central2
policies:
  - local: /tmp
outputs:
  - file: /some/file.json
```

Custom policies can be provided via local directory or remote Github repository.
Example for local directory:

```sh
./gke-policy cluster review -p my_project -n my_cluster -l europe-central2-a  --local-policy-dir ./gke-policies/policy
```

and for Github repository:

```sh
./gke-policy cluster review -p my_project -n my_cluster -l europe-central2-a  --git-policy-repo "https://github.com/google/gke-policy-automation" --git-policy-branch main --git-policy-dir gka-policies/policy
```

Policy definition validation can be done with command:

```sh
gke-policy policy check [arguments...]
```

## Test

Testing policy files with [OPA Policy testing framework](https://www.openpolicyagent.org/docs/latest/policy-testing/)

```sh
opa test <POLICY_DIR>
```

for project policy folder:

```sh
opa test gke-policies
```

## Contributing

Please check out [Contributing](./CONTRIBUTING.md) and [Code of Conduct](./docs/code-of-conduct.md) docs before contributing.
See also [README for policies](./gke-policies/README.md)

## License

[Apache License 2.0](LICENSE)

---
