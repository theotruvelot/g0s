.PHONY: build-agent run-agent run-agent-bin run-agent-dev build-server run-server run-server-bin run-server-dev build-cli run-cli run-cli-bin run-cli-dev clean help test test-nocache test-coverage

build-agent:
	@mkdir -p bin
	@go build -o bin/agent cmd/agent/main.go
	@echo "Agent built successfully: bin/agent"

build-server:
	@mkdir -p bin
	@go build -o bin/server cmd/server/main.go
	@echo "Server built successfully: bin/server"

build-cli:
	@mkdir -p bin
	@go build -o bin/cli cmd/cli/main.go
	@echo "CLI built successfully: bin/cli"

run-agent:
	@if [ -z "$(TOKEN)" ]; then echo "Error: TOKEN is required. Use: make run-agent SERVER=<url> TOKEN=<token>"; exit 1; fi
	@go run cmd/agent/main.go  --grpc-addr $(GRPC_ADDR) --token $(TOKEN) $(if $(INTERVAL),--interval $(INTERVAL),) $(if $(LOG_FORMAT),--log-format $(LOG_FORMAT),) $(if $(LOG_LEVEL),--log-level $(LOG_LEVEL),) $(if $(HEALTH_INTERVAL),--health-check-interval $(HEALTH_INTERVAL),)

run-agent-bin:
	@if [ -z "$(TOKEN)" ]; then echo "Error: TOKEN is required. Use: make run-agent-bin SERVER=<url> TOKEN=<token>"; exit 1; fi
	@bin/agent --token $(TOKEN) --grpc-addr $(GRPC_ADDR) $(if $(INTERVAL),--interval $(INTERVAL),) $(if $(LOG_FORMAT),--log-format $(LOG_FORMAT),) $(if $(LOG_LEVEL),--log-level $(LOG_LEVEL),) $(if $(HEALTH_INTERVAL),--health-check-interval $(HEALTH_INTERVAL),)

run-agent-dev:
	@if [ -z "$(TOKEN)" ]; then echo "Error: TOKEN is required. Use: make run-agent-dev SERVER=<url> TOKEN=<token>"; exit 1; fi
	@go run cmd/agent/main.go --token $(TOKEN) --grpc-addr $(GRPC_ADDR) --log-format console --log-level debug $(if $(INTERVAL),--interval $(INTERVAL),) $(if $(HEALTH_INTERVAL),--health-check-interval $(HEALTH_INTERVAL),)

run-server:
	@go run cmd/server/main.go $(if $(HTTP_ADDR),--http-addr $(HTTP_ADDR),) $(if $(GRPC_ADDR),--grpc-addr $(GRPC_ADDR),) $(if $(LOG_LEVEL),--log-level $(LOG_LEVEL),) $(if $(LOG_FORMAT),--log-format $(LOG_FORMAT),)

run-server-bin:
	@bin/server $(if $(HTTP_ADDR),--http-addr $(HTTP_ADDR),) $(if $(GRPC_ADDR),--grpc-addr $(GRPC_ADDR),) $(if $(LOG_LEVEL),--log-level $(LOG_LEVEL),) $(if $(LOG_FORMAT),--log-format $(LOG_FORMAT),)

run-server-dev:
	@go run cmd/server/main.go $(if $(HTTP_ADDR),--http-addr $(HTTP_ADDR),) $(if $(GRPC_ADDR),--grpc-addr $(GRPC_ADDR),) --log-level debug --log-format console

run-cli:
	@go run cmd/cli/main.go $(if $(SERVER),--server $(SERVER),) $(if $(TOKEN),--token $(TOKEN),) $(if $(LOG_LEVEL),--log-level $(LOG_LEVEL),) $(if $(LOG_FORMAT),--log-format $(LOG_FORMAT),)

run-cli-bin:
	@bin/cli $(if $(SERVER),--server $(SERVER),) $(if $(TOKEN),--token $(TOKEN),) $(if $(LOG_LEVEL),--log-level $(LOG_LEVEL),) $(if $(LOG_FORMAT),--log-format $(LOG_FORMAT),)

run-cli-dev:
	@go run cmd/cli/main.go $(if $(SERVER),--server $(SERVER),) $(if $(TOKEN),--token $(TOKEN),) --log-level debug --log-format console

test:
	@echo "Running tests..."
	@go test ./...

test-nocache:
	@echo "Running tests without cache..."
	@go clean -testcache
	@go test -count=1 ./...

test-coverage:
	@echo "Running tests with coverage..."
	@go test -cover ./...
	@echo "\nDetailed coverage report:"
	@go test -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out
	@rm coverage.out

test-coverage-nocache:
	@echo "Running tests with coverage without cache..."
	@go clean -testcache
	@go test -count=1 -cover ./...
	@echo "\nDetailed coverage report:"
	@go test -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out
	@rm coverage.out

clean:
	@rm -rf bin
	@echo "Cleaned build artifacts"

help:
	@echo "Makefile for g0s"
	@echo ""
	@echo "Usage:"
	@echo "  make build-agent                Build the agent binary"
	@echo "  make build-server               Build the server binary"
	@echo "  make build-cli                  Build the CLI binary"
	@echo "  make run-agent SERVER=URL TOKEN=API_TOKEN [INTERVAL=10] [LOG_FORMAT=json] [LOG_LEVEL=debug] [HEALTH_INTERVAL=30]    Run the agent"
	@echo "  make run-agent-dev SERVER=URL TOKEN=API_TOKEN [INTERVAL=10] [HEALTH_INTERVAL=30]    Run the agent in dev mode (console logs, debug level)"
	@echo "  make run-server [HTTP_ADDR=:8080] [GRPC_ADDR=:9090] [LOG_LEVEL=info] [LOG_FORMAT=json]    Run the server"
	@echo "  make run-server-dev [HTTP_ADDR=:8080] [GRPC_ADDR=:9090]    Run the server in dev mode (console logs, debug level)"
	@echo "  make run-cli [SERVER=URL] [TOKEN=API_TOKEN] [LOG_LEVEL=info] [LOG_FORMAT=json]    Run the CLI with TUI"
	@echo "  make run-cli-dev [SERVER=URL] [TOKEN=API_TOKEN]    Run the CLI in dev mode (console logs, debug level)"
	@echo "  make test                       Run all tests"
	@echo "  make test-nocache              Run all tests without cache"
	@echo "  make test-coverage             Run tests and show coverage report"
	@echo "  make clean                      Remove build artifacts"
	@echo ""
	@echo "Examples:"
	@echo "  make run-agent SERVER=http://localhost:8080 TOKEN=mytoken"
	@echo "  make run-agent-dev SERVER=http://localhost:8080 TOKEN=mytoken INTERVAL=30"
	@echo "  make run-server-dev HTTP_ADDR=:8081"
	@echo "  make run-cli-dev"
	@echo "  make run-cli SERVER=http://localhost:8080 TOKEN=mytoken"

