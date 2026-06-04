#!/usr/bin/env bash
# pod-kill.sh — mata um pod aleatório de um deployment e mede o tempo de recovery.
# Demonstra que o Kubernetes recria o pod automaticamente (self-healing).
#
# Uso:
#   ./chaos/pod-kill.sh ingestion-service
#   ./chaos/pod-kill.sh device-service
#   ./chaos/pod-kill.sh alert-service
set -euo pipefail

DEPLOYMENT="${1:-ingestion-service}"
NAMESPACE="telemetry"

echo "=== Chaos: Pod Kill ==="
echo "Deployment : $DEPLOYMENT"
echo "Namespace  : $NAMESPACE"
echo ""

# Estado antes
echo "--- Pods antes ---"
kubectl get pods -n "$NAMESPACE" -l "app=$DEPLOYMENT"
echo ""

# Escolhe um pod aleatório do deployment
POD=$(kubectl get pods -n "$NAMESPACE" -l "app=$DEPLOYMENT" \
  --field-selector=status.phase=Running \
  -o jsonpath='{.items[*].metadata.name}' \
  | tr ' ' '\n' | shuf | head -1)

if [ -z "$POD" ]; then
  echo "ERRO: nenhum pod Running encontrado para $DEPLOYMENT"
  exit 1
fi

echo "Matando pod: $POD"
kubectl delete pod "$POD" -n "$NAMESPACE"

START=$(date +%s)
echo ""
echo "Aguardando recovery..."

# Aguarda o deployment voltar ao número desejado de réplicas prontas
kubectl rollout status deployment/"$DEPLOYMENT" -n "$NAMESPACE" --timeout=120s

END=$(date +%s)
ELAPSED=$((END - START))

echo ""
echo "--- Pods depois ---"
kubectl get pods -n "$NAMESPACE" -l "app=$DEPLOYMENT"
echo ""
echo "Recovery time: ${ELAPSED}s"
echo "=== Sistema se recuperou automaticamente ==="
