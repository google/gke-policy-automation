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
  storage_apis = var.output_storage.enabled ? ["storage.googleapis.com"] : []
}

resource "google_project_service" "storage-out" {
  for_each           = toset(local.storage_apis)
  project            = data.google_project.project.project_id
  service            = each.key
  disable_on_destroy = false
}

resource "random_id" "storage-out" {
  count       = var.output_storage.enabled ? 1 : 0
  byte_length = 8
}

resource "google_storage_bucket" "storage-out" {
  count                       = var.output_storage.enabled ? 1 : 0
  project                     = data.google_project.project.project_id
  name                        = "${var.output_storage.bucket_name}-${random_id.storage-out[count.index].hex}"
  location                    = var.output_storage.bucket_location
  force_destroy               = true
  uniform_bucket_level_access = true
  depends_on = [
    google_project_service.storage-out
  ]
}

resource "google_storage_bucket_iam_member" "storage-out" {
  count  = var.output_storage.enabled ? 1 : 0
  bucket = google_storage_bucket.storage-out[count.index].name
  role   = "roles/storage.objectCreator"
  member = "serviceAccount:${google_service_account.sa.email}"
}
