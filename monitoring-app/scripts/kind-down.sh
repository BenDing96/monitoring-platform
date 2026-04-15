#!/usr/bin/env bash
set -euo pipefail

CLUSTER="${1:-monitoring-dev}"
REG_NAME="kind-registry"

kind delete cluster --name "${CLUSTER}" || true
docker rm -f "${REG_NAME}" 2>/dev/null || true
