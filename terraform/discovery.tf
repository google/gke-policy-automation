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

locals {
  discovery_apis = var.discovery.organization != null || length(var.discovery.projects) > 0 || length(var.discovery.projects) > 0 ? ["cloudasset.googleapis.com"] : []
}

resource "google_project_service" "discovery" {
  for_each           = toset(local.discovery_apis)
  project            = data.google_project.project.project_id
  service            = each.key
  disable_on_destroy = false
}

resource "google_project_iam_member" "discovery" {
  for_each = toset(var.discovery.projects)
  project  = each.key
  role     = "roles/cloudasset.viewer"
  member   = "serviceAccount:${google_service_account.sa.email}"
}

resource "google_project_iam_member" "cluster-viewer" {
  for_each = toset(var.discovery.projects)
  project  = each.key
  role     = "roles/container.clusterViewer"
  member   = "serviceAccount:${google_service_account.sa.email}"
}

resource "google_folder_iam_member" "discovery" {
  for_each = toset(var.discovery.folders)
  folder   = "folders/${each.key}"
  role     = "roles/cloudasset.viewer"
  member   = "serviceAccount:${google_service_account.sa.email}"
}

resource "google_folder_iam_member" "cluster-viewer" {
  for_each = toset(var.discovery.folders)
  folder   = "folders/${each.key}"
  role     = "roles/container.clusterViewer"
  member   = "serviceAccount:${google_service_account.sa.email}"
}

resource "google_organization_iam_member" "discovery" {
  count  = var.discovery.organization != null ? 1 : 0
  org_id = var.discovery.organization
  role   = "roles/cloudasset.viewer"
  member = "serviceAccount:${google_service_account.sa.email}"
}

resource "google_organization_iam_member" "cluster-viewer" {
  count  = var.discovery.organization != null ? 1 : 0
  org_id = var.discovery.organization
  role   = "roles/container.clusterViewer"
  member = "serviceAccount:${google_service_account.sa.email}"
}
