terraform {
  required_version = ">= 1.9"
  required_providers {
    google = { source = "hashicorp/google"; version = "~> 6.0" }
    random = { source = "hashicorp/random"; version = "~> 3.0" }
  }
  backend "gcs" {
    bucket = ""
    prefix = "monitoring/dev/platform"
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

variable "project_id"  { type = string }
variable "region"      { type = string; default = "us-central1" }

# Read network outputs from remote state
data "terraform_remote_state" "network" {
  backend = "gcs"
  config = {
    bucket = "" # same bucket as backend above
    prefix = "monitoring/dev/network"
  }
}

module "gke" {
  source         = "../../../modules/gke"
  project_id     = var.project_id
  region         = var.region
  env            = "dev"
  network_id     = data.terraform_remote_state.network.outputs.network_id
  subnet_id      = data.terraform_remote_state.network.outputs.subnet_id
  pods_range     = data.terraform_remote_state.network.outputs.pods_range
  services_range = data.terraform_remote_state.network.outputs.services_range
}

module "cloudsql" {
  source      = "../../../modules/cloudsql"
  project_id  = var.project_id
  region      = var.region
  env         = "dev"
  network_id  = data.terraform_remote_state.network.outputs.network_id
  db_tier     = "db-g1-small"
}

module "gcs" {
  source     = "../../../modules/gcs"
  project_id = var.project_id
  region     = var.region
  env        = "dev"
}

module "artifact_registry" {
  source     = "../../../modules/artifact-registry"
  project_id = var.project_id
  region     = var.region
}

module "secrets" {
  source     = "../../../modules/secrets"
  project_id = var.project_id
  env        = "dev"
}

module "iam" {
  source              = "../../../modules/iam"
  project_id          = var.project_id
  env                 = "dev"
  payloads_bucket     = module.gcs.payloads_bucket
  images_repository   = module.artifact_registry.images_repository
}

output "cluster_name"        { value = module.gke.cluster_name }
output "cluster_endpoint"    { value = module.gke.cluster_endpoint }
output "db_host"             { value = module.cloudsql.private_ip }
output "db_name"             { value = module.cloudsql.db_name }
output "db_user"             { value = module.cloudsql.db_user }
output "payloads_bucket"     { value = module.gcs.payloads_bucket }
output "images_repository"   { value = module.artifact_registry.images_repository }
output "charts_repository"   { value = module.artifact_registry.charts_repository }
output "service_accounts"    { value = module.iam.service_accounts }
output "ci_service_account"  { value = module.iam.ci_service_account }
