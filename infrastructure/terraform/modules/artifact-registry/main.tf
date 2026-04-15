variable "project_id" { type = string }
variable "region"     { type = string }

resource "google_artifact_registry_repository" "images" {
  project       = var.project_id
  location      = var.region
  repository_id = "monitoring"
  format        = "DOCKER"
  description   = "monitoring-platform container images"

  cleanup_policies {
    id     = "keep-tagged"
    action = "KEEP"
    condition {
      tag_state = "TAGGED"
    }
  }

  cleanup_policies {
    id     = "delete-untagged-old"
    action = "DELETE"
    condition {
      tag_state  = "UNTAGGED"
      older_than = "604800s" # 7 days
    }
  }
}

resource "google_artifact_registry_repository" "charts" {
  project       = var.project_id
  location      = var.region
  repository_id = "charts"
  format        = "DOCKER"  # OCI-compatible for Helm charts
  description   = "monitoring-platform Helm charts (OCI)"
}

output "images_repository" {
  value = "${var.region}-docker.pkg.dev/${var.project_id}/monitoring"
}

output "charts_repository" {
  value = "${var.region}-docker.pkg.dev/${var.project_id}/charts"
}
