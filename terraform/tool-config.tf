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

resource "google_secret_manager_secret" "config" {

  project = var.project_id

  secret_id = "gke-policy-review-config"

  replication {
    automatic = true
  }

  depends_on = [
    google_project_service.project
  ]
}

data "template_file" "config-template" {
  template = file("${var.config_file_path}")
  vars = {
    PROJECT_ID  = var.project_id
    BUCKET_NAME = google_storage_bucket.report_bucket.name
  }
}

resource "google_secret_manager_secret_version" "config" {
  secret      = google_secret_manager_secret.config.id
  secret_data = data.template_file.config-template.rendered
  depends_on  = [data.template_file.config-template]
}

resource "google_secret_manager_secret_iam_member" "job-sa" {
  secret_id = google_secret_manager_secret.config.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.service_account_cr.email}" # or serviceAccount:my-app@...
}
