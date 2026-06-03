#!/usr/bin/env bash
# setup-cluster.sh — cria o cluster kind, carrega imagens e aplica os manifests
set -euo pipefail

CLUSTER_NAME="telemetry-platform"
K8S_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$K8S_DIR/../.." && pwd)"

echo "=== [1/5] Verificando pré-requisitos ==="
command -v kind    >/dev/null || { echo "ERRO: kind não encontrado."; exit 1; }
command -v kubectl >/dev/null || { echo "ERRO: kubectl não encontrado."; exit 1; }
command -v docker  >/dev/null || { echo "ERRO: docker não encontrado."; exit 1; }

echo "=== [2/5] Criando cluster kind '$CLUSTER_NAME' ==="
if kind get clusters | grep -q "^$CLUSTER_NAME$"; then
  echo "Cluster já existe, pulando criação."
else
  kind create cluster --config "$K8S_DIR/kind-config.yaml"
fi

echo "=== [3/5] Construindo imagens Docker ==="
cd "$ROOT_DIR"
docker build -f services/ingestion-service/Dockerfile -t telemetry/ingestion-service:latest .
docker build -f services/device-service/Dockerfile    -t telemetry/device-service:latest    .
docker build -f services/alert-service/Dockerfile     -t telemetry/alert-service:latest     .
docker build -f services/dashboard-service/Dockerfile -t telemetry/dashboard-service:latest .

echo "=== [4/5] Carregando imagens no kind ==="
kind load docker-image telemetry/ingestion-service:latest --name "$CLUSTER_NAME"
kind load docker-image telemetry/device-service:latest    --name "$CLUSTER_NAME"
kind load docker-image telemetry/alert-service:latest     --name "$CLUSTER_NAME"
kind load docker-image telemetry/dashboard-service:latest --name "$CLUSTER_NAME"

echo "=== [5/5] Aplicando manifests Kubernetes ==="
kubectl apply -f "$K8S_DIR/00-namespace.yaml"
kubectl apply -f "$K8S_DIR/01-configmap.yaml"
kubectl apply -f "$K8S_DIR/02-secrets.yaml"

kubectl apply -f "$K8S_DIR/10-redpanda.yaml"
kubectl apply -f "$K8S_DIR/11-postgres.yaml"
kubectl apply -f "$K8S_DIR/12-redis.yaml"

echo "Aguardando Redpanda ficar pronto (~60s)..."
kubectl wait --for=condition=ready pod -l app=redpanda -n telemetry --timeout=120s

echo "Aguardando PostgreSQL e Redis..."
kubectl wait --for=condition=ready pod -l app=postgres -n telemetry --timeout=60s
kubectl wait --for=condition=ready pod -l app=redis    -n telemetry --timeout=60s

kubectl apply -f "$K8S_DIR/20-ingestion-service.yaml"
kubectl apply -f "$K8S_DIR/21-device-service.yaml"
kubectl apply -f "$K8S_DIR/22-alert-service.yaml"
kubectl apply -f "$K8S_DIR/23-dashboard-service.yaml"

echo ""
echo "=== Cluster pronto! ==="
echo "ingestion-service: http://localhost:30081/health"
echo ""
echo "Para ver os pods:"
echo "  kubectl get pods -n telemetry"
echo ""
echo "Para enviar telemetria de teste:"
echo "  curl -X POST http://localhost:30081/telemetry -H 'Content-Type: application/json' \\"
echo "    -d '{\"device_id\":\"device-001\",\"payload\":{\"lat\":-15.62,\"lon\":-47.66,\"battery\":0.12,\"temperature_c\":43.5,\"speed_kmh\":95.0}}'"
echo ""
echo "Para ver logs do dashboard:"
echo "  kubectl logs -n telemetry -l app=dashboard-service -f"
