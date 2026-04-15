#!/usr/bin/env bash
# Bootstrap cluster-wide addons after GKE cluster is ready.
# Run once per cluster: ./install.sh <env>
set -euo pipefail

ENV="${1:-dev}"

echo "==> cert-manager"
helm upgrade --install cert-manager cert-manager \
  --repo https://charts.jetstack.io \
  --namespace cert-manager --create-namespace \
  --set crds.enabled=true \
  --version "v1.17.x" --wait

echo "==> ingress-nginx"
helm upgrade --install ingress-nginx ingress-nginx \
  --repo https://kubernetes.github.io/ingress-nginx \
  --namespace ingress-nginx --create-namespace \
  -f "$(dirname "$0")/values-ingress-nginx.yaml" \
  --version "4.12.x" --wait

echo "==> external-secrets"
helm upgrade --install external-secrets external-secrets \
  --repo https://charts.external-secrets.io \
  --namespace external-secrets --create-namespace \
  -f "$(dirname "$0")/values-external-secrets.yaml" \
  --version "0.14.x" --wait

echo "==> keda"
helm upgrade --install keda keda \
  --repo https://kedacore.github.io/charts \
  --namespace keda --create-namespace \
  --version "2.17.x" --wait

echo "==> cluster addons installed for env=${ENV}"
