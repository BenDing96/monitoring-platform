variable "project_id"      { type = string }
variable "env"             { type = string }
variable "alert_email"     { type = string; default = "" }

# Enable GCP APIs needed for observability
resource "google_project_service" "monitoring" {
  project            = var.project_id
  service            = "monitoring.googleapis.com"
  disable_on_destroy = false
}

resource "google_project_service" "logging" {
  project            = var.project_id
  service            = "logging.googleapis.com"
  disable_on_destroy = false
}

# Notification channel for alerts (skip if no email set)
resource "google_monitoring_notification_channel" "email" {
  count        = var.alert_email != "" ? 1 : 0
  project      = var.project_id
  display_name = "monitoring-platform alerts (${var.env})"
  type         = "email"
  labels       = { email_address = var.alert_email }
}

# Alert: collector error rate > 1%
resource "google_monitoring_alert_policy" "collector_errors" {
  project      = var.project_id
  display_name = "[${var.env}] Collector high error rate"
  combiner     = "OR"

  conditions {
    display_name = "HTTP 5xx rate"
    condition_threshold {
      filter          = "resource.type=\"k8s_container\" AND resource.labels.namespace_name=\"monitoring\" AND resource.labels.container_name=\"collector\""
      duration        = "60s"
      comparison      = "COMPARISON_GT"
      threshold_value = 0.01
      aggregations {
        alignment_period   = "60s"
        per_series_aligner = "ALIGN_RATE"
      }
    }
  }

  notification_channels = var.alert_email != "" ? [google_monitoring_notification_channel.email[0].id] : []
  severity              = var.env == "prod" ? "CRITICAL" : "WARNING"
}
