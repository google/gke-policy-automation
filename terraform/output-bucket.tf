resource "random_id" "random_id" {
  byte_length = 8
}

resource "google_storage_bucket" "report_bucket" {
  name                        = "gke-policy-review-reports-${random_id.random_id.hex}"
  location                    = "EU"
  project                     = var.project_id
  force_destroy               = true
  uniform_bucket_level_access = true
}