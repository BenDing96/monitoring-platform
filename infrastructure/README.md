# infrastructure

Terraform modules and cluster addons for monitoring-platform on GCP.

## Structure

```
terraform/
  modules/             reusable GCP resource modules
    network/           VPC, subnets, Cloud NAT
    gke/               GKE Autopilot cluster
    cloudsql/          Postgres 16 (private IP, backups, Secret Manager password)
    gcs/               GCS buckets (TF state, payloads)
    artifact-registry/ Docker image + OCI Helm chart repos
    iam/               Service accounts, Workload Identity, least-privilege bindings
    secrets/           Secret Manager entries
    observability/     Cloud Monitoring APIs, alert policies
  envs/
    dev/               development environment
    staging/           staging environment (mirrors dev, larger sizing)
    prod/              production environment (REGIONAL HA, deletion protection)
      00-network/      independent state layer: VPC
      10-platform/     independent state layer: GKE, CloudSQL, GCS, IAM
      20-app/          independent state layer: helm_release per service
      30-observability/ independent state layer: alerts, dashboards
cluster-addons/        Helm values + bootstrap script for cluster-wide addons
```

## First-time bootstrap

**Prerequisites:** `gcloud` CLI authenticated, `terraform >= 1.9`, `helm >= 3.14`, `kubectl`.

### 1. Create the Terraform state bucket (one-time, outside Terraform)

```bash
PROJECT=your-gcp-project-id
REGION=us-central1
gsutil mb -p $PROJECT -l $REGION gs://${PROJECT}-tf-state-dev
gsutil versioning set on gs://${PROJECT}-tf-state-dev
```

### 2. Apply infrastructure layers in order

```bash
# Set env vars (or use a .env file)
export TF_VAR_project_id=your-gcp-project-id
export TF_VAR_region=us-central1
STATE_BUCKET=${TF_VAR_project_id}-tf-state-dev

# Network (VPC)
make plan  ENV=dev LAYER=00-network
make apply ENV=dev LAYER=00-network

# Platform (GKE, CloudSQL, GCS, IAM)
make plan  ENV=dev LAYER=10-platform
make apply ENV=dev LAYER=10-platform

# App services (Helm releases)
make plan  ENV=dev LAYER=20-app
make apply ENV=dev LAYER=20-app
```

### 3. Install cluster addons

```bash
# Connect kubectl to the new cluster
gcloud container clusters get-credentials monitoring-dev --region $REGION --project $PROJECT

# Install cert-manager, ingress-nginx, external-secrets, KEDA
./cluster-addons/install.sh dev
```

### 4. Configure GitHub Actions secrets

In the GitHub repo settings, add:

| Secret | Value |
|---|---|
| `GCP_WORKLOAD_IDENTITY_PROVIDER` | Output of: `gcloud iam workload-identity-pools providers describe github --workload-identity-pool=github-pool --location=global --format='value(name)'` |
| `GCP_TF_SERVICE_ACCOUNT` | Terraform SA email from `10-platform` outputs |
| `GCP_CI_SERVICE_ACCOUNT` | CI SA email from `10-platform` outputs |

Add as repository variables (not secrets):

| Variable | Value |
|---|---|
| `GCP_PROJECT` | Your GCP project ID |
| `GCP_REGION` | e.g. `us-central1` |
| `TF_STATE_BUCKET` | e.g. `your-project-tf-state-dev` |

## Day-to-day operations

```bash
# Plan a specific layer
make plan ENV=dev LAYER=20-app

# Apply a specific layer
make apply ENV=dev LAYER=20-app

# Format all Terraform
make fmt

# Validate all modules
make validate
```

## Bumping a service image (normally done by CI)

Edit `terraform/envs/dev/20-app/terraform.tfvars`:

```hcl
api_tag = "abc1234def5678"
```

Then `make apply ENV=dev LAYER=20-app`.

## State layers — why four?

Each layer has its own GCS state file and can be planned/applied independently.
This limits blast radius: a bad plan on `20-app` (Helm releases) cannot accidentally
destroy the GKE cluster (`10-platform`). The dependency order is strict:
`00-network` → `10-platform` → `20-app`. Never apply a higher layer before a lower one
is healthy.
