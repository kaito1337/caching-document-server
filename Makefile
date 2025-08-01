


MIGRATE_TOOL_URL:=https://github.com/golang-migrate/migrate/releases/download/v4.15.1/migrate.${PLATFORM}-${ARCH}.tar.gz
OUT_PATH?=./bin/server

MIGRATE_DB_URL?=postgres://postgres:secret@localhost:5432/docs_db?sslmode=disable&&query

run:
	go run ./cmd/server/main.go
.PHONY: run

build:
	go build -o ${OUT_PATH} ./cmd/server/main.go
.PHONY: build

bin/migrate:
	mkdir -p ./bin
	curl -L ${MIGRATE_TOOL_URL} | tar xvz -C ./bin/migrate
	chmod +x $@

migrate-up:
	migrate -database="${MIGRATE_DB_URL}" -path ./migrations up
.PHONY: migrate-up

migrate-down:
	migrate -database="${MIGRATE_DB_URL}" -path ./migrations down 1
.PHONY: migrate-down

db-up:
	docker-compose -f deployments/docker-compose.yml up -d
.PHONY: db-up

db-down:
	docker-compose -f deployments/docker-compose.yml down
.PHONY: db-down

