#!/usr/bin/env bash
set -euo pipefail

CLUSTER="${1:-monitoring-dev}"
REG_NAME="kind-registry"
REG_PORT=5001

if ! docker inspect "${REG_NAME}" >/dev/null 2>&1; then
  docker run -d --restart=always -p "127.0.0.1:${REG_PORT}:5000" --name "${REG_NAME}" registry:2
fi

if ! kind get clusters | grep -qx "${CLUSTER}"; then
  cat <<EOF | kind create cluster --name "${CLUSTER}" --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
containerdConfigPatches:
  - |-
    [plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:${REG_PORT}"]
      endpoint = ["http://${REG_NAME}:5000"]
nodes:
  - role: control-plane
    extraPortMappings:
      - containerPort: 80
        hostPort: 80
        protocol: TCP
      - containerPort: 443
        hostPort: 443
        protocol: TCP
EOF
fi

if ! docker network inspect kind | grep -q "${REG_NAME}"; then
  docker network connect kind "${REG_NAME}" || true
fi

cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: local-registry-hosting
  namespace: kube-public
data:
  localRegistryHosting.v1: |
    host: "localhost:${REG_PORT}"
    help: "https://kind.sigs.k8s.io/docs/user/local-registry/"
EOF

echo "kind cluster '${CLUSTER}' ready; registry on localhost:${REG_PORT}"
