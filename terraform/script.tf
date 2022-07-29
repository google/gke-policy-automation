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

resource "local_file" "envs" {
  filename = "cloudrun-config-env.sh"
  content  = <<EOT
export GKE_PA_REGION=${var.region}
export GKE_PA_PROJECT_ID=${data.google_project.project.project_id}
export GKE_PA_JOB_NAME=${var.job_name}
export GKE_PA_SA_EMAIL=${google_service_account.sa.email}
export GKE_PA_SECRET_NAME=${google_secret_manager_secret.config.secret_id}
  EOT
}

resource "null_resource" "script" {
  count = var.run_script ? 1 : 0
  provisioner "local-exec" {
    command     = "./cloudrun-config.sh"
    interpreter = ["bash"]
    environment = {
      REGION      = var.region
      PROJECT_ID  = data.google_project.project.project_id
      JOB_NAME    = var.job_name
      SA_EMAIL    = google_service_account.sa.email
      SECRET_NAME = google_secret_manager_secret.config.secret_id
    }
  }
  depends_on = [
    data.google_project.project,
    google_service_account.sa,
    google_artifact_registry_repository.mirror,
    google_secret_manager_secret.config
  ]
}
