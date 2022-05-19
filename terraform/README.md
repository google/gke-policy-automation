# Setting up an automatic review job

## Introduction

To receive regular reports on the status of your GKE cluster, you can deploy the GKE Policy Review t
ool as a serverless CRON job on GCP.

[Cloud Run Jobs](https://cloud.google.com/run/docs/triggering/using-scheduler) is the perfect tool
to run such a job. Cloud Run Jobs can be triggered by a Cloud Scheduler on any CRON schedule and
only consume resources during the time your job is actually actively running.

Unfortunately, this service is currently only available in one region and not supported by Terraform
yet. While most of this setup can easily be deployed by running this Terraform script, Cloud Run
Jobs have to be deployed manually with the gcloud CLI until the Terraform provider is available.

## Before you begin

Before running the script and the additional gcloud commands, please set the following environment
variables:

`export TF_VAR_job_name="gke-policy-automation-job"`

`export TF_VAR_project_id="YOUR GCP PROJECT ID"`

`export TF_VAR_region="YOUR GCP REGION, e.g. europe-west1"`

Change the config.yaml file to match your GKE cluster. Replace the following properties:
```yaml
clusters:
    - name: YOUR_CLUSTER_NAME
      project: YOUR_PROJECT
      location: YOUR_CLUSTER_LOCATION
```
<pre>

</pre>

Please do **NOT** modify the ((BUCKET_NAME)) placeholder as this will be automatically added by
Terraform before uploading the file to Secret Manager.

## What happens behind the scenes
The Terraform script within this folder enables all required APIs for you and creates necessary
service accounts and IAM bindings. It also creates the Artifact Registry required by Cloud Run,
the GCS bucket for storing the reports and a Secret Manager. The Secret Manager is used to provide
the config.yaml file to the Cloud Run Job, as Cloud Run does not easily support persistent file
storage.

Additionally, the script creates a Cloud Scheduler running every 15 minutes. That scheduler will
ultimately trigger the CLoud Run Job. The 15 minute interval is most likely too frequent for actual
use, but is ideal for demo purposes.

## Deploying required resources via Terraform

Create all required resources with `terraform apply`

## Push the container image to Artifact Registry

Only images stored in Artifact Registry (and the deprecated Container Registry) can be deployed to
Cloud Run, other registries are not supported. Therefore you have to copy the existing container
image from the GitHub Container Registry to Artifact Registry.

Pull the container image from the GitHub Container Registry:

`docker pull ghcr.io/google/gke-policy-automation:latest`

Authenticate against your newly created Artifact Registry:

`gcloud auth configure-docker ${TF_VAR_region}-docker.pkg.dev`

```bash
gcloud auth print-access-token | \
docker login -u oauth2accesstoken --password-stdin https://${TF_VAR_region}-docker.pkg.dev
```

Tag the image for the new location:

`sudo docker tag ghcr.io/google/gke-policy-automation:latest${TF_VAR_region}-docker.pkg.dev/${TF_VAR_project_id}/gke-policy-automation-mirror/gke-policy-automation:1.0`

Push the image to Artifact Registry:

`docker push ${TF_VAR_region}-docker.pkg.dev/${TF_VAR_project_id}/gke-policy-automation-mirror/gke-policy-automation:1.0`

## Create a Cloud Run Job

Create a Cloud Run Job with the image from Artifact Registry:

```bash
gcloud beta run jobs create ${TF_VAR_job_name} \
    --image ${TF_VAR_region}-docker.pkg.dev/${TF_VAR_project_id}/gke-policy-automation-mirror/gke-policy-automation:1.0 \
    --command=/gke-policy,cluster,review \
    --args=-c,/etc/secrets/config.yaml \
    --set-secrets /etc/secrets/config.yaml=gke-policy-review-config:latest \
    --service-account=sa-gke-policy-au@${TF_VAR_project_id}.iam.gserviceaccount.com \
    --region=europe-west9
```

## Test the job

Either wait until the job is triggered automatically by Cloud Scheduler, or head to the Google Cloud
Console to trigger the Cloud Scheduler manually.

## Troubleshooting

If your Cloud Run scheduler shows an error message before you have deployed your Cloud Run Job,
please ignore it. The scheduler cannot reach the job before it has been deployed. If the scheduler
still shows an error after you have deployed the job AND it has been triggered at least once
afterwards, then something is wrong.
