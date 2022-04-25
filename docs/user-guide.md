# GKE Policy Automation User Guide

The GKE Policy Automation is a command line tool that validates GKE clusters against set of best practices.

---

## Table of Contents

* [Authentication](#authentication)
* [Cluster commands](#cluster-commands)
  * [Checking clusters](#checking-clusters)
  * [Dumping cluster data](#dumping-cluster-data)
* [Policy Commands](**)
  * [Validating policies](#contributing)
* [Outputs](#outputs)
* [Configuration file](#configuration-file)
* [Debugging](#debugging)

## Authentication

## Cluster commands

### Checking clusters

### Dumping cluster data

## Policy commands

### Validating policies

## Outputs

The GKE Policy Automation tool produces output to the `stdout`.

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
policies:
  - repository: https://github.com/google/gke-policy-automation
    branch: main
    directory: gke-policiese
  - local: ./my-policies
```

## Debugging

TBD

## OLD

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
./gke-policy cluster review -p my_project -n my_cluster -l europe-central2-a \
--local-policy-dir ./gke-policies/policy
```

and for Github repository:

```sh
./gke-policy cluster review -p my_project -n my_cluster -l europe-central2-a \
--git-policy-repo "https://github.com/google/gke-policy-automation" \
--git-policy-branch main \
--git-policy-dir gka-policies
```
