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

resource "google_artifact_registry_repository" "mirror_repo" {
  provider      = google-beta
  location      = var.region
  repository_id = "gke-policy-automation-mirror"
  description   = "Repository for mirroring GKE policy automation image"
  format        = "docker"
  project       = var.project_id

  depends_on = [
    google_project_service.project
  ]
}

resource "google_service_account" "service_account_cr" {
  account_id   = "sa-gke-policy-au"
  display_name = "Service Account for GKE Policy Automation Cloud Run Service"
  project      = var.project_id
}

resource "google_project_iam_member" "run_invoker" {
  project = var.project_id
  role    = "roles/run.invoker"
  member = "serviceAccount:${google_service_account.service_account_cr.email}"
}

resource "google_project_iam_member" "cluster_viewer" {
  project = var.project_id
  role    = "roles/container.clusterViewer"
  member = "serviceAccount:${google_service_account.service_account_cr.email}"
}

resource "google_project_iam_member" "gcs_writer" {
  project = var.project_id
  role    = "roles/storage.admin"
  member  = "serviceAccount:${google_service_account.service_account_cr.email}"
}

resource "google_project_iam_member" "asset_inventory_search" {
  project = var.project_id
  role    = "roles/cloudasset.viewer"
  member  = "serviceAccount:${google_service_account.service_account_cr.email}"
}
