variable "project_id"  { type = string }
variable "region"      { type = string }
variable "env"         { type = string }
variable "network_id"  { type = string }
variable "db_tier"     { type = string; default = "db-g1-small" }

resource "google_sql_database_instance" "main" {
  name             = "monitoring-${var.env}"
  project          = var.project_id
  region           = var.region
  database_version = "POSTGRES_16"

  settings {
    tier              = var.db_tier
    availability_type = var.env == "prod" ? "REGIONAL" : "ZONAL"
    disk_autoresize   = true
    disk_size         = 20

    ip_configuration {
      ipv4_enabled    = false
      private_network = var.network_id
    }

    backup_configuration {
      enabled                        = true
      point_in_time_recovery_enabled = true
    }

    insights_config {
      query_insights_enabled = true
    }
  }

  deletion_protection = var.env == "prod"
}

resource "google_sql_database" "monitoring" {
  name     = "monitoring"
  instance = google_sql_database_instance.main.name
  project  = var.project_id
}

resource "google_sql_user" "app" {
  name     = "monitoring"
  instance = google_sql_database_instance.main.name
  project  = var.project_id
  password = random_password.db.result
}

resource "random_password" "db" {
  length  = 32
  special = false
}

# Store the password in Secret Manager
resource "google_secret_manager_secret" "db_password" {
  secret_id = "monitoring-${var.env}-db-password"
  project   = var.project_id
  replication { auto {} }
}

resource "google_secret_manager_secret_version" "db_password" {
  secret      = google_secret_manager_secret.db_password.id
  secret_data = random_password.db.result
}

output "instance_name"      { value = google_sql_database_instance.main.name }
output "private_ip"         { value = google_sql_database_instance.main.private_ip_address }
output "db_name"            { value = google_sql_database.monitoring.name }
output "db_user"            { value = google_sql_user.app.name }
output "db_password_secret" { value = google_secret_manager_secret.db_password.secret_id }
