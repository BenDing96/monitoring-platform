terraform {
  required_version = ">= 1.9"
  required_providers {
    google = { source = "hashicorp/google"; version = "~> 6.0" }
  }
  backend "gcs" {
    bucket = "" # set via -backend-config or TF_VAR
    prefix = "monitoring/dev/network"
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

variable "project_id" { type = string }
variable "region"     { type = string; default = "us-central1" }

module "network" {
  source     = "../../../modules/network"
  project_id = var.project_id
  region     = var.region
  env        = "dev"
}

output "network_id"     { value = module.network.network_id }
output "subnet_id"      { value = module.network.subnet_id }
output "pods_range"     { value = module.network.pods_range }
output "services_range" { value = module.network.services_range }
