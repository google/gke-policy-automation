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

output "sa_email" {
  value       = google_service_account.sa.email
  description = "GKE Policy Automation service account's email address."
}

output "repository_id" {
  value       = google_artifact_registry_repository.mirror.id
  description = "Identifier of a GKE Policy Automation repository."
}

output "config_secret_id" {
  value       = google_secret_manager_secret.config.secret_id
  description = "Identifier of a GKE Policy Automation configuration secret."
}

output "env_variables_file" {
  value       = local_file.envs.filename
  description = "File with environmental variables for Artifact Registry and Cloud Run configuration."
}
