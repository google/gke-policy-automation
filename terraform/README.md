# GKE Policy Automation serverless deployment

A Terraform code for deploying GKE Policy Automation as an automatic serverless solution
on Google Cloud Platform.

The solution leverages the below GCP components:

* [Cloud Scheduler](https://cloud.google.com/scheduler)
  to trigger execution of a GKE Policy Automation tool in a periodic manner
* [Cloud Run Jobs](https://cloud.google.com/run/docs/create-jobs)
  to run containerized GKE Policy Automation tool
* [Artifact Registry](https://cloud.google.com/artifact-registry)
  to store GKE Policy Automation tool container image locally
* [Cloud Asset Inventory](https://cloud.google.com/asset-inventory/docs/overview)
  to discover GKE clusters in GCP organization, selected folders or projects
* Optionally, [Cloud Storage](https://cloud.google.com/storage),
  [Cloud Pub/Sub](https://cloud.google.com/pubsub), or
  [Security Command Center](https://cloud.google.com/security-command-center)
  as destinations for cluster evaluation results

![GKE Policy Automation infrastructure](../assets/gke-policy-automation-infra.png)

---

## Prerequisites

* Terraform tool, version >=1.13
* `gcloud` command
* Exiting project for GKE Policy Automation resources
* IAM permissions to create resources in the GKE Policy Automation project
* IAM permissions to create new IAM role bindings on projects, folders or organization levels
  *(depending on desired cluster discovery or outputs configuration)*

## Running

Provision infrastructure with Terraform:

1. Set Terraform configuration variables *(check [examples](#example-configurations)
   or [inputs](#inputs) below for details)*.

   Example `tfvars` file:

   ```hcl
   project_id = "gke-policy-123"
   region     = "europe-west2"

   discovery {
     projects = ["gke-project-one", "gke-project-two"]
   }

   output_storage = {
     enabled         = true
     bucket_name     = "gke-validations"
     bucket_location = "EU"
   }
   ```

2. Adjust GKE Policy Automation's `config.yaml` accordingly
   *(check [User Guide](../docs/user-guide.md) for details)*.
3. Run `terraform init`
4. Run `terraform apply -var-file <your-sample-vars-file.tfvars>`

## What happens behind the scenes

The Terraform script within this folder enables all required APIs for you and creates necessary
service accounts and IAM bindings. Depending on configured cluster discovery options, corresponding
IAM bindings for GKE Policy Automation Service Account are created on projects, folders or
organization levels. The code also creates the Artifact Registry remote repository that proxies tool's
docker images from Github Container registry.
It also creates the Secret Manager secret for storing tool's configuration file.

Depending on configured outputs, the code will provision corresponding resources and IAM role
bindings for Cloud Storage, Pub/Sub or Security Command Center.

Lastly, the script creates a Cloud Scheduler running once per day to trigger Cloud Run Job and the
Cloud Run job itself.

## Example configurations

* Cluster discovery on provided projects and Cloud Storage output

  ```hcl
  project_id = "gke-policy-123"
  region     = "europe-west2"

  discovery = {
    projects = [
      "gke-project-01",
      "gke-project-02"
    ]
  }

  output_storage = {
    enabled         = true
    bucket_name     = "gke-validations"
    bucket_location = "EU"
  }
  ```

* Cluster discovery on selected folders, Pub/Sub and Security Command Center outputs

  ```hcl
  project_id = "gke-policy-123"
  region     = "europe-west2"

  discovery = {
    folders = [
      "112316249356",
      "246836235717"
    ]
  }

  output_pubsub = {
    enabled = true
    topic   = "gke-validations"
  }

  output_scc = {
    enabled      = true
    organization = "123456789012"
  }
  ```

* Cluster discovery on the organization with a Security Command Center output

  ```hcl
  project_id = "gke-policy-123"
  region     = "europe-west2"

  discovery = {
    organization = "123456789012"
  }

  output_scc = {
    enabled      = true
    organization = "153963171798"
  }
  ```

## Inputs

| Name | Description | Type | Required | Default |
|---|---|:---:|:---:|:---:|
| [project_id](variables.tf#L17) | Identifier of an existing GCP project for GKE Policy Automation resources. | `string` | ✓ |  |
| [region](variables.tf#L22) | GCP region for GKE Policy Automation resources. | `string` | ✓ |  |
| [discovery](variables.tf#L51) | Configuration of cluster discovery mechanism. Check [discovery attributes](#discovery-attributes).  | `object` | ✓ |  |
| [job_name](variables.tf#L27) | Name of a Cloud Run Job for GKE Policy Automation container. | `string` |  | `gke-policy-automation` |
| [tool_version](variables.tf#L33) | The version of a GKE Policy Automation tool to deploy. | `string` |  | `latest` |
| [config_file_path](variables.tf#L39) | Path to the YAML file with [GKE Policy Automation configuration](../docs/user-guide.md#configuration-file). | `string` |  | `config.yaml` |
| [cron_interval](variables.tf#L45) | CRON interval for triggering the GKE Policy Automation job. | `string` |  | `"0 1 * * *` |
| [output_storage](variables.tf#L64) | Configuration of Cloud Storage output. Check [Cloud Storage attributes](#cloud-storage-attributes). | `object` |  | `{"enabled" = false}` |
| [output_pubsub](variables.tf#L84) | Configuration of Pub/Sub output. Check [Pub/Sub attributes](#pubsub-attributes) | `object` |  | `{"enabled" = false}` |
| [output_scc](variables.tf#L99) | Configuration of Security Command Center output. Check [Security Command Center attributes](#security-command-center-attributes). | `object` |  | `{"enabled" = false}` |

### Discovery attributes

| Name | Description | Type | Required | Default |
|---|---|:---:|:---:|:---:|
| [organization](variables.tf#L53) | The organization number to provision discovery resources for. *One of `organization`, `folders` or `projects` is required.* | `string` |  | `null` |
| [folders](variables.tf#L54) | List of folder numbers to provision discovery resources for. *One of `organization`, `folders` or `projects` is required.* | `list(string)` |  | `[]` |
| [projects](variables.tf#L55) | List of project identifiers to provision discovery resources for. *One of `organization`, `folders` or `projects` is required.* | `list(string)` |  | `[]` |

### Cloud Storage attributes

| Name | Description | Type | Required | Default |
|---|---|:---:|:---:|:---:|
| [enabled](variables.tf#L66) | Indicates if resources for Cloud Storage output will be provisioned. | `bool` | ✓ |  |
| [bucket_name](variables.tf#L67) | The name of a bucket that will be provisioned. | `string` | ✓ |  |
| [bucket_location](variables.tf#L68) | The [location of a bucket](https://cloud.google.com/storage/docs/locations) that will be provisioned. | `string` | ✓ |  |

### Pub/Sub attributes

| Name | Description | Type | Required | Default |
|---|---|:---:|:---:|:---:|
| [enabled](variables.tf#L86) | Indicates if resources for Pub/Sub output will be provisioned. | `bool` | ✓ |  |
| [topic](variables.tf#L87) | The name of a topic that will be provisioned. | `string` | ✓ |  |

### Security Command Center attributes

| Name | Description | Type | Required | Default |
|---|---|:---:|:---:|:---:|
| [enabled](variables.tf#L101) | Indicates if resources for Pub/Sub output will be provisioned. | `bool` | ✓ |  |
| [organization](variables.tf#L102) | The organization number to provision discovery resources for. | `string` | ✓ |  |
| [provision_source](variables.tf#L103) | Indicates weather to provision `roles/securitycenter.sourcesAdmin` for the tool, so it will be able to automatically register itself as a source. If not enabled, then this has to be done [manually beforehand](../docs/user-guide.md#security-command-center). | `bool` |  | `true` |

## Outputs

| name | description | sensitive |
|---|---|:---:|
| [sa_email](outputs.tf#L17) | GKE Policy Automation service account's email address. |  |
| [repository_id](outputs.tf#L22) | Identifier of a GKE Policy Automation repository. |  |
| [config_secret_id](outputs.tf#L27) | Identifier of a GKE Policy Automation configuration secret. |  |
| [env_variables_file](outputs.tf#L32) | File with environmental variables for Artifact Registry and Cloud Run configuration. |  |

## Troubleshooting

If your Cloud Run scheduler shows an error message before you have deployed your Cloud Run Job,
please ignore it. The scheduler cannot reach the job before it has been deployed. If the scheduler
still shows an error after you have deployed the job AND it has been triggered at least once
afterwards, then something is wrong.
