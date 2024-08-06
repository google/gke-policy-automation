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

resource "google_cloud_scheduler_job" "job" {
  name             = "gke-policy-automation"
  schedule         = var.cron_interval
  description      = "Job triggering GKE Policy Automation"
  time_zone        = "Europe/London"
  attempt_deadline = "320s"
  project          = data.google_project.project.project_id
  region           = var.region

  retry_config {
    retry_count = 1
  }

  http_target {
    http_method = "POST"
    uri         = "https://${var.region}-run.googleapis.com/apis/run.googleapis.com/v1/namespaces/${data.google_project.project.project_id}/jobs/${google_cloud_run_v2_job.gke_policy_automation.name}:run"
    oauth_token {
      service_account_email = google_service_account.sa.email
    }
  }

  depends_on = [
    google_project_service.project
  ]
}
