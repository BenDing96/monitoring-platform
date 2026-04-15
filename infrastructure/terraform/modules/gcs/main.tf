variable "project_id" { type = string }
variable "region"     { type = string }
variable "env"        { type = string }

# Terraform remote state bucket
resource "google_storage_bucket" "tf_state" {
  name          = "${var.project_id}-tf-state-${var.env}"
  project       = var.project_id
  location      = var.region
  force_destroy = false

  versioning { enabled = true }

  uniform_bucket_level_access = true

  lifecycle_rule {
    condition { num_newer_versions = 10 }
    action { type = "Delete" }
  }
}

# Span payload storage (prompts, completions)
resource "google_storage_bucket" "payloads" {
  name          = "${var.project_id}-payloads-${var.env}"
  project       = var.project_id
  location      = var.region
  force_destroy = var.env != "prod"

  uniform_bucket_level_access = true

  lifecycle_rule {
    condition { age = 90 }
    action { type = "SetStorageClass"; storage_class = "NEARLINE" }
  }

  lifecycle_rule {
    condition { age = 365 }
    action { type = "SetStorageClass"; storage_class = "COLDLINE" }
  }
}

output "tf_state_bucket"  { value = google_storage_bucket.tf_state.name }
output "payloads_bucket"  { value = google_storage_bucket.payloads.name }
