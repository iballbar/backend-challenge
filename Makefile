init:
	@echo "🔧 Installing dependency package..."
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@go install github.com/swaggo/swag/cmd/swag@latest
	@go install github.com/air-verse/air@latest

	@go mod tidy
	@go mod download
	@echo "Initial install finished."

run:
	@echo "🚀 Application is running..."
	go run ./cmd/api

air:
	@echo "🚀 Application is running..."
	@air

# run only unit test
test:
	go test -count=1 -v -cover ./...

# run only integration tests (requires Docker).
# Linux/macOS/WSL:  make test-integration
# Windows PowerShell: $env:INT="1"; go test -count=1 -v ./tests/integration/...
test-integration:
	INT=1 go test -count=1 -v -timeout 120s ./tests/integration/...

# run all tests
test-all: test test-integration

generate:
	go generate ./...

fmt:
	go fmt ./...

protoc:
	protoc --proto_path=internal/adapters/grpc/proto --go_out=. --go-grpc_out=. internal/adapters/grpc/proto/*.proto

swag:
	@swag init --parseInternal --parseDependency -g ./cmd/api/main.go
	@swag fmt
	@echo "✅ Swagger docs generated successfully."

# run all containers (API and DB)
compose-up:
	docker compose up -d --build
# stop all containers (API and DB)
compose-down:
	docker compose down

# run only db container
compose-db-up:
	docker compose up -d mongo
# stop only db container
compose-db-down:
	docker compose down mongo