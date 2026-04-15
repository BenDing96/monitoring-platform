variable "project_id"      { type = string }
variable "region"          { type = string }
variable "env"             { type = string }
variable "network_id"      { type = string }
variable "subnet_id"       { type = string }
variable "pods_range"      { type = string }
variable "services_range"  { type = string }
variable "min_nodes"       { type = number; default = 1 }
variable "max_nodes"       { type = number; default = 10 }

resource "google_container_cluster" "main" {
  name     = "monitoring-${var.env}"
  project  = var.project_id
  location = var.region

  # Autopilot — no node pool management
  enable_autopilot = true

  network    = var.network_id
  subnetwork = var.subnet_id

  ip_allocation_policy {
    cluster_secondary_range_name  = var.pods_range
    services_secondary_range_name = var.services_range
  }

  private_cluster_config {
    enable_private_nodes    = true
    enable_private_endpoint = false
    master_ipv4_cidr_block  = "172.16.0.0/28"
  }

  workload_identity_config {
    workload_pool = "${var.project_id}.svc.id.goog"
  }

  master_authorized_networks_config {
    cidr_blocks {
      cidr_block   = "0.0.0.0/0"
      display_name = "all"
    }
  }

  release_channel {
    channel = "REGULAR"
  }

  maintenance_policy {
    recurring_window {
      start_time = "2024-01-01T04:00:00Z"
      end_time   = "2024-01-01T08:00:00Z"
      recurrence = "FREQ=WEEKLY;BYDAY=SU"
    }
  }
}

output "cluster_name"     { value = google_container_cluster.main.name }
output "cluster_endpoint" { value = google_container_cluster.main.endpoint }
output "cluster_ca_cert"  { value = google_container_cluster.main.master_auth[0].cluster_ca_certificate; sensitive = true }
