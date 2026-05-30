db_host := env_var_or_default("DB_HOST", "localhost")
db_port := env_var_or_default("DB_PORT", "5432")
db_user := env_var_or_default("DB_USER", "postgres")
db_pass := env_var_or_default("DB_PASS", "postgres")
db_name := env_var_or_default("DB_NAME", "roly_poly")
db_ssl := env_var_or_default("DB_SSL_MODE", "disable")

dsn := "host={{db_host}} port={{db_port}} user={{db_user}} password={{db_pass}} dbname={{db_name}} sslmode={{db_ssl}}"

default:
    @just --list

build:
    go build -o bin/server ./main.go

run:
    go run ./main.go

test:
    go test -race -count=1 ./...

vet:
    go vet ./...

lint:
    golangci-lint run

migrate-up:
    goose -dir infra/migrations postgres "{{dsn}}" up

migrate-down:
    goose -dir infra/migrations postgres "{{dsn}}" down

migrate-create name:
    goose -dir infra/migrations create {{name}} sql

# docker-compose commands
dc-up:
    docker compose -f infra/docker-compose.yml up -d

dc-down:
    docker compose -f infra/docker-compose.yml down

dc-logs:
    docker compose -f infra/docker-compose.yml logs -f
