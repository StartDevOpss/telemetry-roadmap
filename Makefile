SERVICES := ingestion-service device-service alert-service dashboard-service

.PHONY: up down build logs telemetry health clean ps

## up: Sobe o stack completo (build + start)
up:
	docker compose up -d --build

## down: Para todos os containers
down:
	docker compose down

## build: Reconstrói todas as imagens sem cache
build:
	docker compose build --no-cache

## logs: Tail dos logs de todos os serviços
logs:
	docker compose logs -f ingestion-service device-service alert-service dashboard-service

## logs-infra: Tail dos logs da infraestrutura
logs-infra:
	docker compose logs -f redpanda postgres redis

## ps: Status dos containers
ps:
	docker compose ps

## telemetry: Envia telemetria de teste (bateria baixa + temperatura alta + excesso de velocidade)
telemetry:
	@echo ">>> Enviando telemetria com violações de regra..."
	curl -s -X POST http://localhost:8081/telemetry \
	  -H "Content-Type: application/json" \
	  -d '{"device_id":"device-001","payload":{"lat":-15.62,"lon":-47.66,"battery":0.12,"temperature_c":43.5,"speed_kmh":95.0}}' \
	  && echo " [202 Accepted]" || echo " [ERRO]"

## telemetry-normal: Envia telemetria sem violações
telemetry-normal:
	@echo ">>> Enviando telemetria normal..."
	curl -s -X POST http://localhost:8081/telemetry \
	  -H "Content-Type: application/json" \
	  -d '{"device_id":"device-002","payload":{"lat":-23.55,"lon":-46.63,"battery":0.85,"temperature_c":28.0,"speed_kmh":45.0}}' \
	  && echo " [202 Accepted]" || echo " [ERRO]"

## health: Verifica saúde do ingestion-service
health:
	curl -s -o /dev/null -w "%{http_code}" http://localhost:8081/health

## clean: Remove containers, volumes e imagens locais
clean:
	docker compose down -v --rmi local

## ── Kubernetes (Fase 2) ──────────────────────────────────────────────────────

## k8s-setup: Cria cluster kind, build das imagens e aplica todos os manifests
k8s-setup:
	bash infra/k8s/setup-cluster.sh

## k8s-down: Remove o cluster kind
k8s-down:
	kind delete cluster --name telemetry-platform

## k8s-status: Mostra todos os pods do namespace telemetry
k8s-status:
	kubectl get pods -n telemetry -o wide

## k8s-logs: Tail dos logs do dashboard no k8s
k8s-logs:
	kubectl logs -n telemetry -l app=dashboard-service -f --max-log-requests=5

## k8s-telemetry: Envia telemetria de teste para o cluster k8s (porta 30081)
k8s-telemetry:
	curl -s -X POST http://localhost:30081/telemetry \
	  -H "Content-Type: application/json" \
	  -d '{"device_id":"device-k8s","payload":{"lat":-15.62,"lon":-47.66,"battery":0.12,"temperature_c":43.5,"speed_kmh":95.0}}'

## k8s-reload: Reconstrói imagens, recarrega no kind e faz rollout restart
k8s-reload:
	docker build -f services/ingestion-service/Dockerfile -t telemetry/ingestion-service:latest .
	docker build -f services/device-service/Dockerfile    -t telemetry/device-service:latest    .
	docker build -f services/alert-service/Dockerfile     -t telemetry/alert-service:latest     .
	docker build -f services/dashboard-service/Dockerfile -t telemetry/dashboard-service:latest .
	kind load docker-image telemetry/ingestion-service:latest --name telemetry-platform
	kind load docker-image telemetry/device-service:latest    --name telemetry-platform
	kind load docker-image telemetry/alert-service:latest     --name telemetry-platform
	kind load docker-image telemetry/dashboard-service:latest --name telemetry-platform
	kubectl rollout restart deployment -n telemetry

## ── Utilitários ──────────────────────────────────────────────────────────────

## deps: Gera go.sum para todos os serviços (requer Go instalado)
deps:
	@for svc in $(SERVICES); do \
		echo "==> $$svc"; \
		cd services/$$svc && go mod tidy && cd ../..; \
	done
	cd pkg/events && go mod tidy && cd ../..
