.PHONY: build run clean help

build-agent:
	@mkdir -p bin
	@go build -o bin/agent cmd/agent/main.go
	@echo "Agent built successfully: bin/agent"

run-agent:
	@go run cmd/agent/main.go --server $(SERVER) --token $(TOKEN) $(if $(INTERVAL),--interval $(INTERVAL),) $(if $(LOG_FORMAT),--log-format $(LOG_FORMAT),) $(if $(LOG_LEVEL),--log-level $(LOG_LEVEL),) $(if $(HEALTH_INTERVAL),--health-check-interval $(HEALTH_INTERVAL),)

run-agent-bin:
	@bin/agent --server $(SERVER) --token $(TOKEN) $(if $(INTERVAL),--interval $(INTERVAL),) $(if $(LOG_FORMAT),--log-format $(LOG_FORMAT),) $(if $(LOG_LEVEL),--log-level $(LOG_LEVEL),) $(if $(HEALTH_INTERVAL),--health-check-interval $(HEALTH_INTERVAL),)

run-agent-dev:
	@go run cmd/agent/main.go --server $(SERVER) --token $(TOKEN) --log-format console --log-level debug $(if $(INTERVAL),--interval $(INTERVAL),) $(if $(HEALTH_INTERVAL),--health-check-interval $(HEALTH_INTERVAL),)

clean:
	@rm -rf bin
	@echo "Cleaned build artifacts"

help:
	@echo "Makefile for g0s"
	@echo ""
	@echo "Usage:"
	@echo "  make build-agent                Build the agent binary"
	@echo "  make run-agent SERVER=URL TOKEN=API_TOKEN [INTERVAL=10] [LOG_FORMAT=json] [LOG_LEVEL=debug] [HEALTH_INTERVAL=30]    Run the agent"
	@echo "  make run-agent-dev SERVER=URL TOKEN=API_TOKEN [INTERVAL=10] [HEALTH_INTERVAL=30]    Run the agent in dev mode (console logs, debug level)"
	@echo "  make clean                      Remove build artifacts"

