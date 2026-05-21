.PHONY: run build test tidy docker-build docker-up docker-down lint fmt miniapp-install miniapp-dev miniapp-build

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

miniapp-install:
	cd web/miniapp && npm install

miniapp-dev:
	cd web/miniapp && npm run dev

miniapp-build:
	cd web/miniapp && npm run build

fmt:
	go fmt ./...

lint:
	golangci-lint run ./...
