.PHONY: run build test tidy docker-build docker-up docker-down lint fmt

run:
	go run ./cmd/teeth-bot

build:
	mkdir -p bin
	go build -o bin/teeth-bot ./cmd/teeth-bot

test:
	go test ./...

tidy:
	go mod tidy

docker-build:
	docker compose build

docker-up:
	docker compose up --build

docker-down:
	docker compose down

fmt:
	go fmt ./...

lint:
	golangci-lint run ./...
