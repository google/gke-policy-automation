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

resource "google_secret_manager_secret_version" "config-v1" {

  secret      = google_secret_manager_secret.config.id
  secret_data = replace(file("${var.config_file_path}"), "((BUCKET_NAME))", "${google_storage_bucket.report_bucket.name}")
}

resource "google_secret_manager_secret_iam_member" "job-sa" {
  secret_id = google_secret_manager_secret.config.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.service_account_cr.email}" # or serviceAccount:my-app@...
}