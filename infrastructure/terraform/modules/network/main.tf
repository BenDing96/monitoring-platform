variable "project_id" { type = string }
variable "region"     { type = string }
variable "env"        { type = string }

# VPC
resource "google_compute_network" "main" {
  name                    = "monitoring-${var.env}"
  project                 = var.project_id
  auto_create_subnetworks = false
}

# Primary subnet — nodes live here
resource "google_compute_subnetwork" "main" {
  name                     = "monitoring-${var.env}-nodes"
  project                  = var.project_id
  region                   = var.region
  network                  = google_compute_network.main.id
  ip_cidr_range            = "10.0.0.0/20"
  private_ip_google_access = true

  secondary_ip_range {
    range_name    = "pods"
    ip_cidr_range = "10.1.0.0/16"
  }
  secondary_ip_range {
    range_name    = "services"
    ip_cidr_range = "10.2.0.0/20"
  }
}

# Cloud NAT — lets private nodes reach internet for image pulls
resource "google_compute_router" "main" {
  name    = "monitoring-${var.env}"
  project = var.project_id
  region  = var.region
  network = google_compute_network.main.id
}

resource "google_compute_router_nat" "main" {
  name                               = "monitoring-${var.env}"
  project                            = var.project_id
  router                             = google_compute_router.main.name
  region                             = var.region
  nat_ip_allocate_option             = "AUTO_ONLY"
  source_subnetwork_ip_ranges_to_nat = "ALL_SUBNETWORKS_ALL_IP_RANGES"
}

output "network_id"    { value = google_compute_network.main.id }
output "subnet_id"     { value = google_compute_subnetwork.main.id }
output "pods_range"    { value = "pods" }
output "services_range" { value = "services" }
