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
  type        = string
  description = "Identifier of a GCP project for GKE Policy Automation resources."
}

variable "region" {
  type        = string
  description = "GCP region for GKE Policy Automation resources."
}

variable "job_name" {
  type        = string
  default     = "gke-policy-automation"
  description = "Name of a Cloud Run Job for GKE Policy Automation container."
}

variable "config_file_path" {
  default     = "config.yaml"
  type        = string
  description = "Path to the YAML file with GKE Policy Automation configuration."
}

variable "cron_interval" {
  default     = "0 1 * * *"
  type        = string
  description = "CRON interval for triggering the GKE Policy Automation job."
}

variable "run_script" {
  default     = false
  type        = bool
  description = "Indicates whether to run script for populating Artifact Registry and Cloud Run Jobs"
}

variable "discovery" {
  type        = map(any)
  description = "Configures cluster discovery mechanism."
  default = {
    "enabled" = true
  }
  validation {
    condition     = can(var.discovery.enabled)
    error_message = "Key 'enabled' has to be defined for cluster discovery."
  }
}

variable "output_storage" {
  type        = map(any)
  description = "Configuration of Clud Storage output"
  default = {
    "enabled" = false
  }
  validation {
    condition     = can(var.output_storage.enabled)
    error_message = "Key 'enabled' has to be defined for Cloud Stroage output."
  }
  validation {
    condition     = !try(var.output_storage.enabled, false) || can(var.output_storage.bucket_name)
    error_message = "Key 'bucket_name' has to be defined for Cloud Stroage output."
  }
  validation {
    condition     = !try(var.output_storage.enabled, false) || can(var.output_storage.bucket_location)
    error_message = "Key 'bucket_location' has to be defined for Cloud Stroage output."
  }
}

variable "output_pubsub" {
  type        = map(any)
  description = "Configuration of Pub/Sub output"
  default = {
    "enabled" = false
  }
  validation {
    condition     = can(var.output_pubsub.enabled)
    error_message = "Key 'enabled' has to be defined for Pub/Sub output."
  }
  validation {
    condition     = !try(var.output_pubsub.enabled, false) || can(var.output_pubsub.topic)
    error_message = "Key 'topic' has to be defined for Pub/Sub output."
  }
}

variable "output_scc" {
  type        = map(any)
  description = "Configuration of Security Command Center output"
  default = {
    "enabled" = false
  }
  validation {
    condition     = can(var.output_scc.enabled)
    error_message = "Key 'enabled' has to be defined for Security Command Center output."
  }
  validation {
    condition     = !try(var.output_scc.enabled, false) || can(var.output_scc.organization)
    error_message = "Key 'organization' has to be defined for Security Command Center output."
  }
}
