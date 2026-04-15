variable "project_id"          { type = string }
variable "env"                 { type = string }
variable "payloads_bucket"     { type = string }
variable "images_repository"   { type = string }

# Service accounts for each workload — principle of least privilege
locals {
  services = ["collector", "ingestor", "api", "worker"]
}

resource "google_service_account" "workload" {
  for_each     = toset(local.services)
  project      = var.project_id
  account_id   = "monitoring-${each.key}-${var.env}"
  display_name = "monitoring-platform ${each.key} (${var.env})"
}

# Workload Identity bindings — pods assume the GSA without key files
resource "google_service_account_iam_member" "workload_identity" {
  for_each           = toset(local.services)
  service_account_id = google_service_account.workload[each.key].name
  role               = "roles/iam.workloadIdentityUser"
  member             = "serviceAccount:${var.project_id}.svc.id.goog[monitoring/${each.key}]"
}

# ingestor + collector write payloads to GCS
resource "google_storage_bucket_iam_member" "payload_writer" {
  for_each = toset(["collector", "ingestor"])
  bucket   = var.payloads_bucket
  role     = "roles/storage.objectCreator"
  member   = "serviceAccount:${google_service_account.workload[each.key].email}"
}

# api reads payloads from GCS
resource "google_storage_bucket_iam_member" "payload_reader" {
  bucket = var.payloads_bucket
  role   = "roles/storage.objectViewer"
  member = "serviceAccount:${google_service_account.workload["api"].email}"
}

# All workloads can read Secret Manager secrets
resource "google_project_iam_member" "secret_accessor" {
  for_each = toset(local.services)
  project  = var.project_id
  role     = "roles/secretmanager.secretAccessor"
  member   = "serviceAccount:${google_service_account.workload[each.key].email}"
}

# CI service account — pushes images, bumps chart versions
resource "google_service_account" "ci" {
  project      = var.project_id
  account_id   = "monitoring-ci"
  display_name = "monitoring-platform CI/CD"
}

resource "google_artifact_registry_repository_iam_member" "ci_image_push" {
  project    = var.project_id
  location   = split("-docker.pkg.dev/", var.images_repository)[0]  # region
  repository = "monitoring"
  role       = "roles/artifactregistry.writer"
  member     = "serviceAccount:${google_service_account.ci.email}"
}

output "service_accounts" {
  value = { for k, v in google_service_account.workload : k => v.email }
}
output "ci_service_account" { value = google_service_account.ci.email }
