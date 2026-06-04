#!/usr/bin/env bash
# scale-zero.sh — escala um deployment para 0 réplicas e restaura.
# Simula uma falha total de um serviço e valida que o sistema volta
# ao estado desejado quando as réplicas são restauradas.
#
# Uso:
#   ./chaos/scale-zero.sh device-service 30
#   (mantém em 0 por 30 segundos antes de restaurar)
set -euo pipefail

DEPLOYMENT="${1:-device-service}"
DOWNTIME="${2:-20}"
NAMESPACE="telemetry"

# Descobre o número atual de réplicas para restaurar depois
ORIGINAL_REPLICAS=$(kubectl get deployment "$DEPLOYMENT" -n "$NAMESPACE" \
  -o jsonpath='{.spec.replicas}')

echo "=== Chaos: Scale to Zero ==="
echo "Deployment       : $DEPLOYMENT"
echo "Réplicas originais: $ORIGINAL_REPLICAS"
echo "Downtime          : ${DOWNTIME}s"
echo ""

echo "--- Pods antes ---"
kubectl get pods -n "$NAMESPACE" -l "app=$DEPLOYMENT"
echo ""

echo "Escalando $DEPLOYMENT para 0..."
kubectl scale deployment "$DEPLOYMENT" -n "$NAMESPACE" --replicas=0
kubectl wait --for=delete pod -l "app=$DEPLOYMENT" -n "$NAMESPACE" --timeout=60s 2>/dev/null || true

echo "Serviço indisponível por ${DOWNTIME}s — envie telemetria agora para ver o comportamento."
sleep "$DOWNTIME"

echo ""
echo "Restaurando $DEPLOYMENT para $ORIGINAL_REPLICAS réplica(s)..."
kubectl scale deployment "$DEPLOYMENT" -n "$NAMESPACE" --replicas="$ORIGINAL_REPLICAS"

START=$(date +%s)
kubectl rollout status deployment/"$DEPLOYMENT" -n "$NAMESPACE" --timeout=120s
END=$(date +%s)

echo ""
echo "--- Pods depois ---"
kubectl get pods -n "$NAMESPACE" -l "app=$DEPLOYMENT"
echo ""
echo "Recovery time: $((END - START))s"
echo "=== Serviço restaurado ==="
