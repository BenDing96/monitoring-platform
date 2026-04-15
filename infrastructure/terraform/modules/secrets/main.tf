variable "project_id" { type = string }
variable "env"        { type = string }

# Placeholder secrets — values populated by other modules or manually.
# Apps read via External Secrets Operator → k8s Secrets.
locals {
  secret_ids = [
    "monitoring-${var.env}-clickhouse-password",
    "monitoring-${var.env}-redis-auth",
  ]
}

resource "google_secret_manager_secret" "secrets" {
  for_each  = toset(local.secret_ids)
  secret_id = each.key
  project   = var.project_id
  replication { auto {} }
}

output "secret_ids" {
  value = { for k, v in google_secret_manager_secret.secrets : k => v.secret_id }
}
