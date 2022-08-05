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
  scc_apis = try(var.output_scc.enabled) ? ["securitycenter.googleapis.com"] : []
}

resource "google_project_service" "scc-out" {
  for_each           = toset(local.scc_apis)
  project            = data.google_project.project.project_id
  service            = each.key
  disable_on_destroy = false
}

resource "google_organization_iam_member" "scc-out-findings" {
  count  = try(var.output_scc.enabled) ? 1 : 0
  org_id = var.output_scc.organization
  role   = "roles/securitycenter.findingsEditor"
  member = "serviceAccount:${google_service_account.sa.email}"
}

resource "google_organization_iam_member" "scc-out-sources" {
  count  = try(var.output_scc.enabled, false) && try(var.output_scc.provision_source, true) ? 1 : 0
  org_id = var.output_scc.organization
  role   = "roles/securitycenter.sourcesAdmin"
  member = "serviceAccount:${google_service_account.sa.email}"
}
