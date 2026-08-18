SHELL := /bin/sh

GO := go
GOFLAGS := -tags sqlite_fts5
COMPOSE := docker compose
WEB_DIR := web

.PHONY: help setup server server-build admin docker docker-release \
	db-start db-down migrate-up migrate-down migrate-clean \
	compose-logging compose-logging-down dashboard-install \
	dev dev-lite web web-install web-dev web-build clean

help:
	@echo "Usage:"
	@echo "  make setup         migrate + seed (cold start)"
	@echo "  make server        run the Go backend (port 8080)"
	@echo "  make server-build  build the Go backend binary"
	@echo "  make admin         run admin CLI"
	@echo "  make migrate-up      run database migrations"
	@echo "  make migrate-down  roll back all migrations"
	@echo "  make migrate-clean drop and re-create the database, then migrate"
	@echo "  make web           install deps + start Vite dev server (port 5173)"
	@echo "  make web-dev       start Vite dev server only"
	@echo "  make web-install   install npm dependencies"
	@echo "  make web-build     build the frontend for production"
	@echo "  make db-start      start postgres via docker-compose"
	@echo "  make db-down       stop postgres and delete its data volume"
	@echo "  make dev           start full stack (db + logging + server)"
	@echo "  make dev-lite      start db + server (no logging stack)"
	@echo "  make compose-logging start logging stack (OpenObserve + Fluent Bit + postgres-exporter)"
	@echo "  make compose-logging-down stop logging stack (keeps data volumes)"
	@echo "  make dashboard-install provision the OpenObserve dashboard"
	@echo "  make clean         remove build artifacts"

setup: db-start
	$(GO) run $(GOFLAGS) ./cmd/admin/ setup

server:
	$(GO) run $(GOFLAGS) ./cmd/server/

server-build:
	$(GO) build $(GOFLAGS) -o build/server ./cmd/server/

admin:
	$(GO) run ./cmd/admin/

docker:
	docker build -t addressbook:latest .

docker-release:
	docker build -t addressbook:$$(git describe --tags --always) .

db-start:
	$(COMPOSE) up -d --wait postgres

db-down:
	$(COMPOSE) down -v postgres

migrate-up:
	$(GO) run ./cmd/admin/ migrate

migrate-down:
	$(GO) run ./cmd/admin/ migrate:down

migrate-clean:
	$(GO) run ./cmd/admin/ db:clean

compose-logging:
	$(COMPOSE) -f docker-compose.yml -f resources/logging/docker-compose.logging.yml up -d

compose-logging-down:
	$(COMPOSE) -f docker-compose.yml -f resources/logging/docker-compose.logging.yml down

dashboard-install:
	./resources/logging/dashboards/install-dashboard.sh

dev: db-start compose-logging
	$(GO) run $(GOFLAGS) ./cmd/server/

dev-lite: db-start
	$(GO) run $(GOFLAGS) ./cmd/server/

web: web-install
	npm --prefix $(WEB_DIR) run dev

web-install:
	npm --prefix $(WEB_DIR) install

web-dev:
	npm --prefix $(WEB_DIR) run dev

web-build:
	npm --prefix $(WEB_DIR) run build

clean:
	rm -rf build/ web/dist/
