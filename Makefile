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

## deps: Gera go.sum para todos os serviços (requer Go instalado)
deps:
	@for svc in $(SERVICES); do \
		echo "==> $$svc"; \
		cd services/$$svc && go mod tidy && cd ../..; \
	done
	cd pkg/events && go mod tidy && cd ../..
