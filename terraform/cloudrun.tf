/**
 * Copyright 2022 Google LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

resource "google_service_account" "sa" {
  project      = data.google_project.project.project_id
  account_id   = "gke-policy-automation"
  display_name = "Service Account for GKE Policy Automation Cloud Run Service"
}

resource "google_artifact_registry_repository" "mirror" {
  project       = data.google_project.project.project_id
  location      = var.region
  repository_id = "ghcr"
  description   = "Proxy repository for Github container registry"
  format        = "docker"
  mode          = "REMOTE_REPOSITORY"

  remote_repository_config {
    description                 = "Proxy for Github container registry"
    disable_upstream_validation = false
    docker_repository {
      custom_repository {
        uri = "https://ghcr.io"
      }
    }
  }

  depends_on = [
    google_project_service.project
  ]
}

resource "google_cloud_run_v2_job" "gke_policy_automation" {
  project  = data.google_project.project.project_id
  name     = var.job_name
  location = var.region

  template {
    template {
      service_account = google_service_account.sa.email
      volumes {
        name = "config"
        secret {
          secret = google_secret_manager_secret.config.id
          items {
            version = "latest"
            path    = "config.yaml"
          }
        }
      }
      containers {
        image   = "${var.region}-docker.pkg.dev/${google_artifact_registry_repository.mirror.project}/${google_artifact_registry_repository.mirror.name}/google/gke-policy-automation:${var.tool_version}"
        command = ["/gke-policy", "check"]
        args    = ["-c", "/etc/secrets/config.yaml"]
        env {
          name  = "GKE_POLICY_LOG"
          value = "INFO"
        }
        volume_mounts {
          name       = "config"
          mount_path = "/etc/secrets"
        }
      }
    }
  }
  depends_on = [
    google_project_service.project
  ]
}

resource "google_project_iam_member" "run_invoker" {
  project = data.google_project.project.project_id
  role    = "roles/run.invoker"
  member  = "serviceAccount:${google_service_account.sa.email}"
}
