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
  type = object({
    organization = optional(string, null)
    projects     = optional(list(string), [])
    folders      = optional(list(string), [])
  })
  description = "Configures cluster discovery mechanism."
  validation {
    condition     = var.discovery.organization != null || length(var.discovery.projects) > 0 || length(var.discovery.folders) > 0
    error_message = "At least one of organization, projects, folders has to be defined for cluster discovery"
  }
}

variable "output_storage" {
  type = object({
    enabled         = bool
    bucket_name     = optional(string)
    bucket_location = optional(string)
  })
  description = "Configuration of Cloud Storage output"
  default = {
    enabled = false
  }
  validation {
    condition     = !var.output_storage.enabled || var.output_storage.bucket_name != null
    error_message = "Key 'bucket_name' has to be defined for Cloud Stroage output."
  }
  validation {
    condition     = !var.output_storage.enabled || var.output_storage.bucket_location != null
    error_message = "Key 'bucket_location' has to be defined for Cloud Stroage output."
  }
}

variable "output_pubsub" {
  type = object({
    enabled = bool
    topic   = optional(string)
  })
  description = "Configuration of Pub/Sub output"
  default = {
    "enabled" = false
  }
  validation {
    condition     = !var.output_pubsub.enabled || var.output_pubsub.topic != null
    error_message = "Key 'topic' has to be defined for Pub/Sub output."
  }
}

/*
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
*/