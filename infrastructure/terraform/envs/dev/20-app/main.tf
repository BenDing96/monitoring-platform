terraform {
  required_version = ">= 1.9"
  required_providers {
    google     = { source = "hashicorp/google";     version = "~> 6.0" }
    kubernetes = { source = "hashicorp/kubernetes"; version = "~> 2.0" }
    helm       = { source = "hashicorp/helm";       version = "~> 2.0" }
  }
  backend "gcs" {
    bucket = ""
    prefix = "monitoring/dev/app"
  }
}

data "terraform_remote_state" "platform" {
  backend = "gcs"
  config = {
    bucket = ""
    prefix = "monitoring/dev/platform"
  }
}

data "google_client_config" "default" {}

data "google_container_cluster" "main" {
  name     = data.terraform_remote_state.platform.outputs.cluster_name
  location = var.region
  project  = var.project_id
}

provider "google" {
  project = var.project_id
  region  = var.region
}

provider "kubernetes" {
  host                   = "https://${data.terraform_remote_state.platform.outputs.cluster_endpoint}"
  token                  = data.google_client_config.default.access_token
  cluster_ca_certificate = base64decode(data.google_container_cluster.main.master_auth[0].cluster_ca_certificate)
}

provider "helm" {
  kubernetes {
    host                   = "https://${data.terraform_remote_state.platform.outputs.cluster_endpoint}"
    token                  = data.google_client_config.default.access_token
    cluster_ca_certificate = base64decode(data.google_container_cluster.main.master_auth[0].cluster_ca_certificate)
  }
}

variable "project_id"       { type = string }
variable "region"           { type = string; default = "us-central1" }
variable "collector_tag"    { type = string; default = "dev" }
variable "ingestor_tag"     { type = string; default = "dev" }
variable "api_tag"          { type = string; default = "dev" }
variable "worker_tag"       { type = string; default = "dev" }
variable "console_tag"      { type = string; default = "dev" }

locals {
  registry = data.terraform_remote_state.platform.outputs.images_repository
  db_host  = data.terraform_remote_state.platform.outputs.db_host
}

resource "kubernetes_namespace" "monitoring" {
  metadata { name = "monitoring" }
}

resource "helm_release" "collector" {
  name       = "collector"
  repository = "oci://${data.terraform_remote_state.platform.outputs.charts_repository}"
  chart      = "collector"
  namespace  = kubernetes_namespace.monitoring.metadata[0].name

  set { name = "image.repository"; value = "${local.registry}/collector" }
  set { name = "image.tag";        value = var.collector_tag }
  set { name = "replicaCount";     value = "2" }

  set_sensitive {
    name  = "env[0].value"
    value = "dev"
  }
}

resource "helm_release" "ingestor" {
  name       = "ingestor"
  repository = "oci://${data.terraform_remote_state.platform.outputs.charts_repository}"
  chart      = "ingestor"
  namespace  = kubernetes_namespace.monitoring.metadata[0].name

  set { name = "image.repository";      value = "${local.registry}/ingestor" }
  set { name = "image.tag";             value = var.ingestor_tag }
  set { name = "clickhouse.addr";       value = "clickhouse.monitoring.svc:9000" }
}

resource "helm_release" "api" {
  name       = "api"
  repository = "oci://${data.terraform_remote_state.platform.outputs.charts_repository}"
  chart      = "api"
  namespace  = kubernetes_namespace.monitoring.metadata[0].name

  set { name = "image.repository"; value = "${local.registry}/api" }
  set { name = "image.tag";        value = var.api_tag }
  set { name = "replicaCount";     value = "2" }
}

resource "helm_release" "console" {
  name       = "console"
  repository = "oci://${data.terraform_remote_state.platform.outputs.charts_repository}"
  chart      = "console"
  namespace  = kubernetes_namespace.monitoring.metadata[0].name

  set { name = "image.repository"; value = "${local.registry}/console" }
  set { name = "image.tag";        value = var.console_tag }
}
