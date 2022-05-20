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

variable "project_id" {
  type = string
  description = "GCP project ID of project to deploy the GKE policy review tool into"
}

variable "region" {
  type        = string
  description = "GCP region in which to deploy the resources"
}

variable "job_region" {
  default = "europe-west9"
  type        = string
  description = "GCP region in which to deploy the Cloud Run Job"
}

variable "job_name" {
  type        = string
  description = "Name for the Cloud Run Job"
}

variable "config_file_path" {
  default = "config.yaml"
  type = string
  description = "Path to the file containing the YAML configuration"
}