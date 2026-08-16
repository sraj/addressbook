.PHONY: help server server-build admin migrate migrate-down migrate-clean docker web web-install web-dev web-build setup clean db-start db-down dev dev-lite compose-logging compose-logging-down

help:
	@echo "Usage:"
	@echo "  make setup         migrate + seed (cold start)"
	@echo "  make server        run the Go backend (port 8080)"
	@echo "  make server-build  build the Go backend binary"
	@echo "  make admin         run admin CLI"
	@echo "  make migrate-up       run database migrations"
	@echo "  make migrate-down  roll back all migrations"
	@echo "  make migrate-clean drop and re-create the database, then migrate"
	@echo "  make web           install deps + start Vite dev server (port 5173)"
	@echo "  make web-dev       start Vite dev server only"
	@echo "  make web-install   install npm dependencies"
	@echo "  make web-build     build the frontend for production"
	@echo "  make db-start     start postgres via docker-compose"
	@echo "  make db-down      stop postgres and delete its data volume"
	@echo "  make dev           start full stack (db + logging + server)"
	@echo "  make dev-lite      start db + server (no logging stack)"
	@echo "  make compose-logging      start logging stack (OpenObserve + Fluent Bit + postgres-exporter)"
	@echo "  make compose-logging-down stop logging stack (keeps data volumes)"
	@echo "  make dashboard-install    provision the OpenObserve dashboard"
	@echo "  make clean         remove build artifacts"

setup:
	go run -tags sqlite_fts5 ./cmd/admin/ setup

server:
	go run -tags sqlite_fts5 ./cmd/server/

server-build:
	go build -tags sqlite_fts5 -o build/server ./cmd/server/

admin:
	go run ./cmd/admin/

docker:
	docker build -t addressbook:latest .

docker-release:
	docker build -t addressbook:$$(git describe --tags --always) .

db-start:
	docker compose up -d postgres

db-down:
	docker compose down -v postgres

migrate-up:
	go run ./cmd/admin/ migrate

migrate-down:
	go run ./cmd/admin/ migrate:down

migrate-clean:
	go run ./cmd/admin/ db:clean

compose-logging:
	docker compose -f docker-compose.yml -f resources/logging/docker-compose.logging.yml up -d

compose-logging-down:
	docker compose -f docker-compose.yml -f resources/logging/docker-compose.logging.yml down

dashboard-install:
	./resources/logging/dashboards/install-dashboard.sh

dev: db-start compose-logging
	go run -tags sqlite_fts5 ./cmd/server/

dev-lite: db-start
	go run -tags sqlite_fts5 ./cmd/server/

web:
	cd web && npm install && npm run dev

web-install:
	cd web && npm install

web-dev:
	cd web && npm run dev

web-build:
	cd web && npm run build

clean:
	rm -rf build/ web/dist/
