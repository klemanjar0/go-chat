SHELL := /bin/bash

ENV_FILE       ?= .env.local
MIGRATIONS_DIR := migrations
APP_NAME       := chat
BIN            := bin/$(APP_NAME)
IMAGE          := go-chat:local

# ---------- Pretty output ----------

# Disable colors when stdout is not a TTY (CI logs stay clean).
ifneq (,$(findstring xterm,$(TERM)))
  C_RESET := \033[0m
  C_BOLD  := \033[1m
  C_DIM   := \033[2m
  C_RED   := \033[31m
  C_GRN   := \033[32m
  C_YEL   := \033[33m
  C_BLU   := \033[34m
  C_MAG   := \033[35m
  C_CYA   := \033[36m
endif

# log_step "message"
define log_step
	printf "$(C_CYA)▶$(C_RESET)  $(C_BOLD)%s$(C_RESET)\n" "$(1)"
endef
define log_ok
	printf "$(C_GRN)✓$(C_RESET)  %s\n" "$(1)"
endef
define log_warn
	printf "$(C_YEL)⚠$(C_RESET)  %s\n" "$(1)"
endef
define log_info
	printf "$(C_DIM)ℹ$(C_RESET)  $(C_DIM)%s$(C_RESET)\n" "$(1)"
endef

# Load env vars from $(ENV_FILE) (silent if file missing).
define load_env
	set -a; [ -f $(ENV_FILE) ] && source $(ENV_FILE); set +a;
endef

# Load env + warn if file missing.
define load_env_required
	@if [ ! -f $(ENV_FILE) ]; then \
		printf "$(C_RED)✗$(C_RESET)  $(C_BOLD)$(ENV_FILE) not found$(C_RESET) — copy from .env.example or create one.\n"; \
		exit 1; \
	fi
	@$(call log_info,using env file: $(ENV_FILE))
endef

DB_URL = postgres://$$DB_USER:$$DB_PASSWORD@$$DB_HOST:$$DB_PORT/$$DB_NAME?sslmode=$$DB_SSL_MODE

.DEFAULT_GOAL := help

# ---------- Help ----------

.PHONY: help
help: ## Show this help
	@printf "\n$(C_BOLD)go-chat$(C_RESET)  $(C_DIM)— development commands$(C_RESET)\n\n"
	@awk 'BEGIN {FS = ":.*?## "; section=""} \
		/^# ==== / { sub(/^# ==== /, ""); sub(/ ====$$/, ""); section=$$0; printf "\n  $(C_MAG)%s$(C_RESET)\n", section; next } \
		/^[a-zA-Z_-]+:.*?## / { printf "    $(C_CYA)%-22s$(C_RESET) %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@printf "\n  $(C_DIM)tip: most targets accept ENV_FILE=path/to/.env to override defaults$(C_RESET)\n\n"

# ==== App ====

.PHONY: run
run: ## Run the app locally with .env.local loaded
	$(call load_env_required)
	@$(call log_step,starting $(APP_NAME) (go run ./cmd))
	@$(load_env) go run ./cmd

.PHONY: build
build: ## Build the binary into ./bin/chat
	@$(call log_step,building $(BIN))
	@mkdir -p bin
	@go build -trimpath -ldflags="-s -w" -o $(BIN) ./cmd
	@$(call log_ok,built $(BIN) ($(shell du -h $(BIN) 2>/dev/null | cut -f1 || echo "?")))

.PHONY: tidy
tidy: ## go mod tidy
	@$(call log_step,go mod tidy)
	@go mod tidy
	@$(call log_ok,modules tidy)

.PHONY: fmt
fmt: ## go fmt ./...
	@$(call log_step,formatting (go fmt ./...))
	@go fmt ./...
	@$(call log_ok,formatted)

.PHONY: vet
vet: ## go vet ./...
	@$(call log_step,vet (go vet ./...))
	@go vet ./...
	@$(call log_ok,vet clean)

.PHONY: test
test: ## go test ./... -race -count=1
	@$(call log_step,running tests (race, no-cache))
	@go test ./... -race -count=1
	@$(call log_ok,tests passed)

.PHONY: check
check: fmt vet test ## Run fmt + vet + test

# ==== Local infra ====

.PHONY: infra-up
infra-up: ## Start postgres + redis via docker compose
	$(call load_env_required)
	@$(call log_step,starting infra (postgres, redis))
	@$(load_env) docker compose up -d postgres redis
	@$(call log_step,waiting for healthchecks)
	@$(load_env) timeout=60; \
	  while [ $$timeout -gt 0 ]; do \
	    pg=$$(docker inspect --format='{{.State.Health.Status}}' go-chat-postgres 2>/dev/null || echo missing); \
	    rd=$$(docker inspect --format='{{.State.Health.Status}}' go-chat-redis    2>/dev/null || echo missing); \
	    if [ "$$pg" = "healthy" ] && [ "$$rd" = "healthy" ]; then \
	      printf "$(C_GRN)✓$(C_RESET)  postgres=$$pg  redis=$$rd\n"; break; \
	    fi; \
	    printf "$(C_DIM)…$(C_RESET)  postgres=$$pg  redis=$$rd  (waiting)\r"; \
	    sleep 2; timeout=$$((timeout-2)); \
	  done; \
	  if [ $$timeout -le 0 ]; then \
	    printf "$(C_RED)✗$(C_RESET)  infra did not become healthy in 60s — check 'make infra-logs'\n"; exit 1; \
	  fi
	@$(call log_info,postgres → localhost:$$(grep '^DB_PORT' $(ENV_FILE) | cut -d= -f2)  redis → localhost:$$(grep '^REDIS_PORT' $(ENV_FILE) | cut -d= -f2))
	@$(call log_info,next: 'make migrate-up' to apply schema)

.PHONY: infra-down
infra-down: ## Stop infra (keeps volumes)
	@$(call log_step,stopping infra)
	@docker compose down
	@$(call log_ok,infra stopped (volumes preserved))

.PHONY: infra-nuke
infra-nuke: ## Stop infra AND remove volumes (DATA LOSS)
	@$(call log_warn,this will DELETE all postgres + redis data)
	@printf "    type 'nuke' to confirm: "; read ans; [ "$$ans" = "nuke" ] || (printf "$(C_DIM)aborted$(C_RESET)\n"; exit 1)
	@$(call log_step,destroying infra and volumes)
	@docker compose down -v
	@$(call log_ok,infra and volumes destroyed)

.PHONY: infra-logs
infra-logs: ## Tail infra logs
	@$(call log_step,tailing logs (Ctrl-C to exit))
	@docker compose logs -f postgres redis

.PHONY: infra-status
infra-status: ## Show status of infra containers
	@$(call log_step,infra status)
	@docker compose ps

.PHONY: infra-psql
infra-psql: ## Open psql shell against the local DB
	$(call load_env_required)
	@$(call log_step,connecting to $$DB_NAME as $$DB_USER)
	@$(load_env) docker compose exec -it postgres psql -U $$DB_USER -d $$DB_NAME

.PHONY: infra-redis-cli
infra-redis-cli: ## Open redis-cli against the local Redis
	@$(call log_step,connecting to redis)
	@docker compose exec -it redis redis-cli

.PHONY: docker-build
docker-build: ## Build the production app image
	@$(call log_step,building docker image $(IMAGE))
	@docker build -t $(IMAGE) .
	@$(call log_ok,image built: $(IMAGE))

# ==== Migrations (golang-migrate) ====

.PHONY: migrate-up
migrate-up: ## Apply all pending migrations
	$(call load_env_required)
	@$(call log_step,applying pending migrations)
	@$(load_env) migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up
	@$(call log_ok,migrations up-to-date)
	@$(MAKE) -s migrate-version

.PHONY: migrate-down
migrate-down: ## Roll back the most recent migration
	$(call load_env_required)
	@$(call log_warn,rolling back the most recent migration)
	@$(load_env) migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down 1
	@$(call log_ok,rollback complete)
	@$(MAKE) -s migrate-version

.PHONY: migrate-down-all
migrate-down-all: ## Roll back ALL migrations (DATA LOSS)
	$(call load_env_required)
	@$(call log_warn,this will roll back ALL migrations and drop tables)
	@printf "    type 'down-all' to confirm: "; read ans; [ "$$ans" = "down-all" ] || (printf "$(C_DIM)aborted$(C_RESET)\n"; exit 1)
	@$(load_env) migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down -all
	@$(call log_ok,all migrations rolled back)

.PHONY: migrate-force
migrate-force: ## Force version (usage: make migrate-force v=1)
	$(call load_env_required)
	@if [ -z "$(v)" ]; then printf "$(C_RED)✗$(C_RESET)  usage: $(C_BOLD)make migrate-force v=<version>$(C_RESET)\n"; exit 1; fi
	@$(call log_warn,forcing schema version to $(v) (use only to recover from a dirty state))
	@$(load_env) migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" force $(v)
	@$(call log_ok,version forced to $(v))

.PHONY: migrate-version
migrate-version: ## Print current schema version
	$(call load_env_required)
	@printf "$(C_DIM)ℹ$(C_RESET)  current schema version: "
	@$(load_env) migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" version 2>&1 | sed 's/^/  /'

.PHONY: migrate-new
migrate-new: ## Create a new migration pair (usage: make migrate-new name=add_messages)
	@if [ -z "$(name)" ]; then printf "$(C_RED)✗$(C_RESET)  usage: $(C_BOLD)make migrate-new name=<snake_case>$(C_RESET)\n"; exit 1; fi
	@$(call log_step,creating migration '$(name)' in $(MIGRATIONS_DIR)/)
	@migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)
	@$(call log_ok,migration files created — edit them and run 'make migrate-up')

# ==== Codegen ====

.PHONY: sqlc-generate
sqlc-generate: ## Run sqlc to regenerate DB code
	@$(call log_step,regenerating sqlc code (internal/db/sqlcgen))
	@sqlc generate
	@$(call log_ok,sqlc code regenerated)

.PHONY: sqlc-vet
sqlc-vet: ## Lint sqlc queries against the schema
	@$(call log_step,running sqlc vet)
	@sqlc vet
	@$(call log_ok,sqlc queries OK)

# ==== Toolchain ====

.PHONY: tools
tools: ## Install sqlc + golang-migrate locally
	@$(call log_step,installing sqlc)
	@go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	@$(call log_step,installing golang-migrate)
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@$(call log_ok,tools installed in $$(go env GOPATH)/bin — make sure that's on your PATH)

.PHONY: doctor
doctor: ## Check that required tools are installed
	@printf "$(C_BOLD)checking required tools$(C_RESET)\n"
	@for t in go docker sqlc migrate; do \
	  if command -v $$t >/dev/null 2>&1; then \
	    printf "  $(C_GRN)✓$(C_RESET) %-10s %s\n" "$$t" "$$($$t version 2>/dev/null | head -1 || echo installed)"; \
	  else \
	    printf "  $(C_RED)✗$(C_RESET) %-10s $(C_DIM)not found$(C_RESET)\n" "$$t"; \
	  fi; \
	done
	@if [ -f $(ENV_FILE) ]; then \
	  printf "  $(C_GRN)✓$(C_RESET) %-10s %s\n" "$(ENV_FILE)" "present"; \
	else \
	  printf "  $(C_RED)✗$(C_RESET) %-10s $(C_DIM)missing$(C_RESET)\n" "$(ENV_FILE)"; \
	fi

# ==== Composite workflows ====

.PHONY: dev-up
dev-up: infra-up migrate-up sqlc-generate ## Bring up infra, run migrations, regenerate sqlc
	@$(call log_ok,dev environment ready — run 'make run' to start the app)

.PHONY: dev-reset
dev-reset: infra-nuke infra-up migrate-up ## Wipe data, restart infra, re-apply migrations
	@$(call log_ok,dev environment reset)
