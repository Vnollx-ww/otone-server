.PHONY: run build test vet fmt migrate-up migrate-down compose-up compose-down tidy

run:
	go run ./cmd/api

build:
	go build -o bin/api ./cmd/api

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -w .

migrate-up:
	go run ./cmd/migrate up

migrate-down:
	go run ./cmd/migrate down

compose-up:
	docker compose -f deployments/docker-compose.yml up -d

compose-down:
	docker compose -f deployments/docker-compose.yml down

docker-build:
	docker build -f deployments/Dockerfile -t otome-server:dev .

tidy:
	go mod tidy
