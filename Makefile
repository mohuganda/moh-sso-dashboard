# -----------------------------
# Project Metadata
# -----------------------------
PROJECT_NAME = moh-sso-dashboard
GO_CMD = go
GO_MAIN = ./cmd/server/main.go
FRONTEND_DIR = ./frontend
BINARY_NAME = sso-server
DOCKER_IMAGE = $(PROJECT_NAME):latest

# -----------------------------
# Go Backend Commands
# -----------------------------
.PHONY: build
build:
	@echo "🚀 Building Go backend..."
	cd backend && $(GO_CMD) build -o ../bin/$(BINARY_NAME) $(GO_MAIN)
	@echo "✅ Build complete: bin/$(BINARY_NAME)"

.PHONY: run
run:
	@echo "🏃 Running Go backend..."
	cd backend && $(GO_CMD) run $(GO_MAIN)

.PHONY: test
test:
	@echo "🧪 Running tests..."
	cd backend && $(GO_CMD) test ./... -v

.PHONY: fmt
fmt:
	@echo "🎨 Formatting code..."
	cd backend && $(GO_CMD) fmt ./...

.PHONY: tidy
tidy:
	@echo "🧹 Tidying dependencies..."
	cd backend && $(GO_CMD) mod tidy

.PHONY: clean
clean:
	@echo "🧽 Cleaning up..."
	rm -rf bin/
	@echo "✅ Clean complete"

# -----------------------------
# React Frontend Commands
# -----------------------------
.PHONY: frontend-install
frontend-install:
	@echo "📦 Installing frontend dependencies..."
	cd $(FRONTEND_DIR) && npm install

.PHONY: frontend-build
frontend-build:
	@echo "🏗️ Building React app..."
	cd $(FRONTEND_DIR) && npm run build

.PHONY: frontend-start
frontend-start:
	@echo "🌐 Starting React development server..."
	cd $(FRONTEND_DIR) && npm start

# -----------------------------
# Docker Commands
# -----------------------------
.PHONY: docker-build
docker-build:
	@echo "🐳 Building Docker image..."
	docker build -t $(DOCKER_IMAGE) .

.PHONY: docker-run
docker-run:
	@echo "🐳 Running Docker container..."
	docker run -p 8080:8080 $(DOCKER_IMAGE)

# -----------------------------
# Docker Compose Commands
# -----------------------------
.PHONY: compose-up
compose-up:
	@echo "📦 Starting services with Docker Compose..."
	docker compose up -d

.PHONY: compose-down
compose-down:
	@echo "🧯 Stopping services..."
	docker compose down

# -----------------------------
# Utility Commands
# -----------------------------
.PHONY: dev
dev:
	@echo "⚡ Starting full-stack dev environment..."
	make -j2 run frontend-start

.PHONY: all
all: tidy build frontend-build
	@echo "✅ All components built successfully!"

# -----------------------------
# Kubernetes Local (Docker Desktop)
# -----------------------------

K8S_NAMESPACE = sso-local
HELM_RELEASE = moh-sso-dashboard
HELM_CHART = ./charts/moh-sso
VALUES_LOCAL = charts/moh-sso/values-local.yaml
VALUES_DEV = charts/moh-sso/values-dev.yaml
VALUES_PROD = charts/moh-sso/values-prod.yaml
HELM_OUT_DIR ?= dist/helm
CHART_VERSION ?=
APP_VERSION ?=

BACKEND_IMAGE = moh-sso-dashboard-backend:local
FRONTEND_IMAGE = moh-sso-dashboard-frontend:local
KIND_CLUSTER ?= desktop

.PHONY: local-namespace
local-namespace:
	@echo "📦 Ensuring namespace exists..."
	kubectl get namespace $(K8S_NAMESPACE) >/dev/null 2>&1 || kubectl create namespace $(K8S_NAMESPACE)

.PHONY: local-build
local-build: local-build-images local-load-images

.PHONY: local-build-backend
local-build-backend: local-build-backend-image local-load-backend-image

.PHONY: local-build-backend-image
local-build-backend-image:
	@echo "🐳 Building backend (Dockerfile) for arm64..."
	docker buildx build \
		--platform linux/arm64 \
		-f backend/Dockerfile \
		-t $(BACKEND_IMAGE) \
		--load \
		./backend

.PHONY: local-build-images
local-build-images: local-build-backend-image
	@echo "🐳 Building frontend (Dockerfile.dev) for arm64..."
	docker buildx build \
		--platform linux/arm64 \
		-f frontend/Dockerfile \
		-t $(FRONTEND_IMAGE) \
		--load \
		./frontend

.PHONY: local-load-images
local-load-images: local-load-backend-image
	@if command -v kind >/dev/null 2>&1 && kind get clusters 2>/dev/null | grep -qx "$(KIND_CLUSTER)"; then \
		echo "📥 Loading frontend image into kind cluster '$(KIND_CLUSTER)'..."; \
		kind load docker-image $(FRONTEND_IMAGE) --name $(KIND_CLUSTER); \
	elif command -v docker >/dev/null 2>&1 && docker ps --format '{{.Names}}' | grep -qx "$(KIND_CLUSTER)-control-plane"; then \
		echo "📥 Loading frontend image into Docker Desktop Kubernetes node '$(KIND_CLUSTER)-control-plane'..."; \
		docker save $(FRONTEND_IMAGE) | docker exec -i $(KIND_CLUSTER)-control-plane ctr -n k8s.io images import -; \
	else \
		echo "ℹ️  Kubernetes node '$(KIND_CLUSTER)-control-plane' not detected. If pods show ErrImageNeverPull, install kind or load images manually:"; \
		echo "   brew install kind"; \
		echo "   kind load docker-image $(BACKEND_IMAGE) --name $(KIND_CLUSTER)"; \
		echo "   kind load docker-image $(FRONTEND_IMAGE) --name $(KIND_CLUSTER)"; \
	fi
	@echo "✅ Local arm64 images loaded into containerd"

.PHONY: local-load-backend-image
local-load-backend-image:
	@if command -v kind >/dev/null 2>&1 && kind get clusters 2>/dev/null | grep -qx "$(KIND_CLUSTER)"; then \
		echo "📥 Loading backend image into kind cluster '$(KIND_CLUSTER)'..."; \
		kind load docker-image $(BACKEND_IMAGE) --name $(KIND_CLUSTER); \
	elif command -v docker >/dev/null 2>&1 && docker ps --format '{{.Names}}' | grep -qx "$(KIND_CLUSTER)-control-plane"; then \
		echo "📥 Loading backend image into Docker Desktop Kubernetes node '$(KIND_CLUSTER)-control-plane'..."; \
		docker save $(BACKEND_IMAGE) | docker exec -i $(KIND_CLUSTER)-control-plane ctr -n k8s.io images import -; \
	else \
		echo "Kubernetes node '$(KIND_CLUSTER)-control-plane' was not detected."; \
		exit 1; \
	fi
	@echo "✅ Local backend image loaded into containerd"


.PHONY: local-up
local-up: local-namespace
	@echo "🚀 Deploying to Kubernetes (local)..."
	helm upgrade --install $(HELM_RELEASE) $(HELM_CHART) \
		-n $(K8S_NAMESPACE) \
		-f $(VALUES_LOCAL)
	@echo "✅ Deployment applied"

.PHONY: local-down
local-down:
	@echo "🧯 Uninstalling local release..."
	helm uninstall $(HELM_RELEASE) -n $(K8S_NAMESPACE) || true

.PHONY: local-reset-data
local-reset-data: local-down
	@echo "⏳ Waiting for local release pods to release persistent volumes..."
	kubectl wait --for=delete pod -l app.kubernetes.io/instance=$(HELM_RELEASE) -n $(K8S_NAMESPACE) --timeout=120s || true
	@echo "⚠️ Deleting persistent volumes owned by release $(HELM_RELEASE)..."
	kubectl delete pvc -l app.kubernetes.io/instance=$(HELM_RELEASE) -n $(K8S_NAMESPACE) --ignore-not-found=true
	kubectl delete pvc $(HELM_RELEASE)-postgres-app-pvc $(HELM_RELEASE)-postgres-keycloak-pvc -n $(K8S_NAMESPACE) --ignore-not-found=true
	@echo "🔎 Remaining PVCs in $(K8S_NAMESPACE):"
	kubectl get pvc -n $(K8S_NAMESPACE) || true
	@echo "✅ Local release data reset. Run 'make local-up' to recreate the stack."

.PHONY: local-db-migration-status
local-db-migration-status:
	@echo "🧾 Current local app DB migration state..."
	kubectl exec -n $(K8S_NAMESPACE) statefulset/$(HELM_RELEASE)-postgres-app -- \
		sh -c 'psql -U "$$POSTGRES_USER" -d "$$POSTGRES_DB" -c "SELECT * FROM schema_migrations;"'

.PHONY: local-db-repair-dirty-10
local-db-repair-dirty-10:
	@state=$$(kubectl exec -n $(K8S_NAMESPACE) statefulset/$(HELM_RELEASE)-postgres-app -- \
		sh -c 'psql -At -U "$$POSTGRES_USER" -d "$$POSTGRES_DB" -c "SELECT version::text || chr(58) || dirty::text FROM schema_migrations;"'); \
	if [ "$$state" != "10:true" ]; then \
		echo "Refusing repair: expected migration state 10:true, found $$state" >&2; \
		exit 1; \
	fi
	@echo "🛑 Scaling backend down before repairing migration state..."
	kubectl scale deployment/$(HELM_RELEASE)-backend -n $(K8S_NAMESPACE) --replicas=0 || true
	kubectl wait --for=delete pod -l app=$(HELM_RELEASE)-backend -n $(K8S_NAMESPACE) --timeout=120s || true
	@echo "🧹 Repairing dirty migration 10 in local app DB..."
	kubectl exec -n $(K8S_NAMESPACE) statefulset/$(HELM_RELEASE)-postgres-app -- \
		sh -c 'psql -v ON_ERROR_STOP=1 -U "$$POSTGRES_USER" -d "$$POSTGRES_DB" \
			-c "DROP TABLE IF EXISTS facilities;" \
			-c "UPDATE schema_migrations SET version = 9, dirty = false WHERE version = 10 AND dirty = true;" \
			-c "SELECT * FROM schema_migrations;"'
	@echo "▶️ Scaling backend back up so migration 10 can rerun cleanly..."
	kubectl scale deployment/$(HELM_RELEASE)-backend -n $(K8S_NAMESPACE) --replicas=1

.PHONY: local-migration-logs
local-migration-logs:
	@MIGRATION_POD=$$(kubectl get pods -n $(K8S_NAMESPACE) -l app=$(HELM_RELEASE)-backend -o jsonpath='{.items[0].metadata.name}'); \
	if [ -z "$$MIGRATION_POD" ]; then \
		echo "No backend pod found in $(K8S_NAMESPACE)."; \
		exit 1; \
	fi; \
	echo "Migration logs from $$MIGRATION_POD:"; \
	kubectl logs -n $(K8S_NAMESPACE) "$$MIGRATION_POD" -c migrate-database

.PHONY: local-migration-status
local-migration-status:
	@kubectl get pods -n $(K8S_NAMESPACE) -l app=$(HELM_RELEASE)-backend
	@kubectl get pods -n $(K8S_NAMESPACE) -l app=$(HELM_RELEASE)-backend -o jsonpath='{range .items[*].status.initContainerStatuses[*]}{.name}{"\t"}{.state}{"\n"}{end}'

.PHONY: local-restart
local-restart:
	@echo "🔄 Restarting backend & frontend..."
	kubectl rollout restart deployment $(HELM_RELEASE)-backend -n $(K8S_NAMESPACE)
	kubectl rollout restart deployment $(HELM_RELEASE)-frontend -n $(K8S_NAMESPACE)

.PHONY: local-logs
local-logs:
	kubectl get pods -n $(K8S_NAMESPACE)

.PHONY: local-backend-logs
local-backend-logs:
	kubectl logs -f deployment/$(HELM_RELEASE)-backend -n $(K8S_NAMESPACE)

.PHONY: local-keycloak-logs
local-keycloak-logs:
	kubectl logs -f statefulset/$(HELM_RELEASE)-keycloak -n $(K8S_NAMESPACE)

.PHONY: local-forward
local-forward:
	@echo "🌐 Port forwarding..."
	@echo "Frontend:  http://localhost:3000"
	@echo "Backend:   http://localhost:9000"
	@echo "Keycloak:  http://localhost:8081"
	kubectl port-forward svc/$(HELM_RELEASE)-frontend 3000:80 -n $(K8S_NAMESPACE) & \
	kubectl port-forward svc/$(HELM_RELEASE)-backend 9000:9000 -n $(K8S_NAMESPACE) & \
	kubectl port-forward svc/$(HELM_RELEASE)-keycloak 8081:8080 -n $(K8S_NAMESPACE)

.PHONY: local
local: local-build local-up
	@echo "🎉 Local stack deployed!"
	@echo "Run 'make local-forward' to access services."

.PHONY: helm-lint
helm-lint:
	@echo "🔎 Linting Helm chart..."
	helm lint $(HELM_CHART)

.PHONY: helm-template
helm-template:
	@echo "🧾 Rendering base Helm values..."
	helm template $(HELM_RELEASE) $(HELM_CHART)

.PHONY: helm-template-local
helm-template-local:
	@echo "🧾 Rendering local Helm values..."
	helm template $(HELM_RELEASE) $(HELM_CHART) -f $(VALUES_LOCAL)

.PHONY: helm-template-dev
helm-template-dev:
	@echo "🧾 Rendering dev Helm values..."
	helm template $(HELM_RELEASE) $(HELM_CHART) -f $(VALUES_DEV)

.PHONY: helm-template-prod
helm-template-prod:
	@test -n "$(BACKEND_TAG)" || (echo "BACKEND_TAG is required for helm-template-prod" && exit 1)
	@test -n "$(FRONTEND_TAG)" || (echo "FRONTEND_TAG is required for helm-template-prod" && exit 1)
	@echo "🧾 Rendering prod Helm values..."
	helm template $(HELM_RELEASE) $(HELM_CHART) -f $(VALUES_PROD) \
		--set backend.image.tag=$(BACKEND_TAG) \
		--set frontend.image.tag=$(FRONTEND_TAG)

.PHONY: helm-check
helm-check: helm-lint helm-template helm-template-local helm-template-dev
	@echo "✅ Helm chart checks passed"

.PHONY: helm-check-prod
helm-check-prod: helm-lint helm-template-prod
	@echo "✅ Helm production chart checks passed"

.PHONY: helm-package
helm-package:
	@test -n "$(BACKEND_TAG)" || (echo "BACKEND_TAG is required for helm-package" && exit 1)
	@test -n "$(FRONTEND_TAG)" || (echo "FRONTEND_TAG is required for helm-package" && exit 1)
	@echo "📦 Packaging Helm chart..."
	BACKEND_TAG="$(BACKEND_TAG)" \
	FRONTEND_TAG="$(FRONTEND_TAG)" \
	CHART_VERSION="$(CHART_VERSION)" \
	APP_VERSION="$(APP_VERSION)" \
	HELM_RELEASE="$(HELM_RELEASE)" \
	HELM_OUT_DIR="$(HELM_OUT_DIR)" \
	CHART_DIR="$(HELM_CHART)" \
	./scripts/package-helm-chart.sh

.PHONY: helm-package-check
helm-package-check: helm-package
	@echo "✅ Helm chart package checks passed"

# -----------------------------
# Kubernetes DEV (Cluster)
# -----------------------------

DEV_NAMESPACE = sso-dev
.PHONY: dev-up
dev-up:
	@echo "🚀 Deploying DEV environment..."
	helm upgrade --install $(HELM_RELEASE) $(HELM_CHART) \
		-n $(DEV_NAMESPACE) \
		-f $(DEV_VALUES)

.PHONY: dev-down
dev-down:
	@echo "🧯 Removing DEV release..."
	helm uninstall $(HELM_RELEASE) -n $(DEV_NAMESPACE) || true

.PHONY: dev-restart
dev-restart:
	@echo "🔄 Restarting backend & frontend..."
	kubectl rollout restart deployment $(HELM_RELEASE)-backend -n $(DEV_NAMESPACE)
	kubectl rollout restart deployment $(HELM_RELEASE)-frontend -n $(DEV_NAMESPACE)

.PHONY: dev-logs
dev-logs:
	kubectl get pods -n $(DEV_NAMESPACE)

.PHONY: dev-logs-watch
dev-logs-watch:
	kubectl get pods -n $(DEV_NAMESPACE) -w

.PHONY: dev-backend-logs
dev-backend-logs:
	kubectl logs -f deployment/$(HELM_RELEASE)-backend -n $(DEV_NAMESPACE)

.PHONY: dev-frontend-logs
dev-frontend-logs:
	kubectl logs -f deployment/$(HELM_RELEASE)-frontend -n $(DEV_NAMESPACE)

# -----------------------------
# Kubernetes PROD
# -----------------------------

PROD_NAMESPACE = sso-prod
.PHONY: prod-up
prod-up:
	@echo "🚀 Deploying PROD environment..."
	@test -n "$(BACKEND_TAG)" || (echo "BACKEND_TAG is required for prod-up" && exit 1)
	@test -n "$(FRONTEND_TAG)" || (echo "FRONTEND_TAG is required for prod-up" && exit 1)
	helm upgrade --install $(HELM_RELEASE) $(HELM_CHART) \
		-n $(PROD_NAMESPACE) \
		-f $(VALUES_PROD) \
		--set backend.image.tag=$(BACKEND_TAG) \
		--set frontend.image.tag=$(FRONTEND_TAG)

.PHONY: prod-down
prod-down:
	@echo "🧯 Removing PROD release..."
	helm uninstall $(HELM_RELEASE) -n $(PROD_NAMESPACE) || true


.PHONY: prod-restart
prod-restart:
	@echo "🔄 Restarting backend & frontend..."
	kubectl rollout restart deployment $(HELM_RELEASE)-backend -n $(PROD_NAMESPACE)
	kubectl rollout restart deployment $(HELM_RELEASE)-frontend -n $(PROD_NAMESPACE)

.PHONY: prod-logs
prod-logs:
	kubectl get pods -n $(PROD_NAMESPACE)

.PHONY: prod-backend-logs
prod-backend-logs:
	kubectl logs -f deployment/$(HELM_RELEASE)-backend -n $(PROD_NAMESPACE)

.PHONY: prod-frontend-logs
prod-frontend-logs:
	kubectl logs -f deployment/$(HELM_RELEASE)-frontend -n $(PROD_NAMESPACE)
