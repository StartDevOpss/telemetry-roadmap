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

## k8s-ns: Define telemetry como namespace padrão (requer kubens)
k8s-ns:
	kubens telemetry

## k8s-status: Mostra todos os pods do namespace telemetry
k8s-status:
	kubectl get pods -n telemetry -o wide

## k9s: Abre o k9s já no namespace telemetry
k9s:
	k9s -n telemetry

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

## ── IaC + GitOps (Fase 5) ────────────────────────────────────────────────────

## tf-init: Inicializa o Terraform (baixa o provider kind)
tf-init:
	cd infra/terraform && terraform init

## tf-plan: Planeja a criação do cluster kind
tf-plan:
	cd infra/terraform && terraform plan

## tf-apply: Cria o cluster kind via Terraform
tf-apply:
	cd infra/terraform && terraform apply -auto-approve

## tf-destroy: Destrói o cluster kind via Terraform
tf-destroy:
	cd infra/terraform && terraform destroy -auto-approve

## helm-lint: Valida o chart sem instalar
helm-lint:
	helm lint infra/helm/telemetry-platform

## helm-template: Renderiza os templates localmente (dry-run)
helm-template:
	helm template telemetry-platform infra/helm/telemetry-platform --namespace telemetry

## helm-install: Instala/atualiza o chart no cluster
helm-install:
	helm upgrade --install telemetry-platform infra/helm/telemetry-platform \
	  --namespace telemetry --create-namespace

## helm-uninstall: Remove o release do cluster
helm-uninstall:
	helm uninstall telemetry-platform --namespace telemetry

## argocd-bootstrap: Instala o ArgoCD e cria a Application de GitOps
argocd-bootstrap:
	kubectl create namespace argocd --dry-run=client -o yaml | kubectl apply -f -
	kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/v2.10.0/manifests/install.yaml
	kubectl wait --for=condition=available deployment -l app.kubernetes.io/name=argocd-server \
	  -n argocd --timeout=180s
	kubectl apply -f infra/argocd/apps/telemetry-platform.yaml
	@echo ""
	@echo "ArgoCD instalado. Acesse a UI:"
	@echo "  make argocd-ui        (em outro terminal)"
	@echo "  Usuário: admin"
	@echo "  Senha:   make argocd-password"

## argocd-password: Mostra a senha inicial do ArgoCD
argocd-password:
	@kubectl -n argocd get secret argocd-initial-admin-secret \
	  -o jsonpath="{.data.password}" | base64 -d && echo

## argocd-ui: Port-forward da UI do ArgoCD para https://localhost:8080
argocd-ui:
	kubectl port-forward svc/argocd-server -n argocd 8080:443

## ── Utilitários ──────────────────────────────────────────────────────────────

## deps: Gera go.sum para todos os serviços (requer Go instalado)
deps:
	@for svc in $(SERVICES); do \
		echo "==> $$svc"; \
		cd services/$$svc && go mod tidy && cd ../..; \
	done
	cd pkg/events && go mod tidy && cd ../..
