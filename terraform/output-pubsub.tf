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

/*
locals {
  pubsub_apis = try(var.output_pubsub.enabled) ? ["pubsub.googleapis.com"] : []
}

resource "google_project_service" "pubsub-out" {
  for_each           = toset(local.pubsub_apis)
  project            = data.google_project.project.project_id
  service            = each.key
  disable_on_destroy = false
}

resource "google_pubsub_topic" "pubsub-out" {
  count   = try(var.output_pubsub.enabled) ? 1 : 0
  project = data.google_project.project.project_id
  name    = var.output_pubsub.topic
  depends_on = [
    google_project_service.pubsub-out
  ]
}

resource "google_pubsub_topic_iam_member" "pubsub-out" {
  count   = try(var.output_pubsub.enabled) ? 1 : 0
  project = data.google_project.project.project_id
  topic   = google_pubsub_topic.pubsub-out[count.index].name
  role    = "roles/pubsub.publisher"
  member  = "serviceAccount:${google_service_account.sa.email}"
}
*/