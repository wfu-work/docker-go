APP_NAME ?= nav-docker
SERVER_BIN ?= $(APP_NAME)-server
AGENT_BIN ?= $(APP_NAME)-agent
VERSION ?= m6-dev
GO ?= go
GOCACHE ?= /private/tmp/nav-docker-gocache
BIN_DIR ?= bin
RELEASE_DIR ?= $(BIN_DIR)/release
SERVER_URL ?= http://127.0.0.1:8888/api
AGENT_GUID ?=
GOFLAGS ?= -trimpath
SERVER_LDFLAGS ?= -s -w
AGENT_LDFLAGS ?= -s -w -X main.agentVersion=$(VERSION)

.PHONY: help build server agent release test tidy clean run-server run-agent linux-amd64 linux-arm64 darwin-amd64 darwin-arm64

help:
	@echo "NAV Docker build targets:"
	@echo "  make build              Build server and agent into $(BIN_DIR)/"
	@echo "  make server             Build server only"
	@echo "  make agent              Build agent only"
	@echo "  make release            Build linux/darwin amd64/arm64 release binaries"
	@echo "  make test               Run go tests"
	@echo "  make tidy               Run go mod tidy"
	@echo "  make clean              Remove build outputs"
	@echo "  make run-server         Run server locally"
	@echo "  make run-agent TOKEN=x  Run agent locally"
	@echo ""
	@echo "Variables:"
	@echo "  VERSION=v0.1.0          Inject agent version"
	@echo "  BIN_DIR=bin             Output directory"
	@echo "  SERVER_URL=...          Agent server URL for run-agent"

build: server agent

server:
	@mkdir -p $(BIN_DIR)
	GOCACHE=$(GOCACHE) $(GO) build $(GOFLAGS) -ldflags "$(SERVER_LDFLAGS)" -o $(BIN_DIR)/$(SERVER_BIN) .

agent:
	@mkdir -p $(BIN_DIR)
	GOCACHE=$(GOCACHE) $(GO) build $(GOFLAGS) -ldflags "$(AGENT_LDFLAGS)" -o $(BIN_DIR)/$(AGENT_BIN) ./cmd/nav-docker-agent

release: linux-amd64 linux-arm64 darwin-amd64 darwin-arm64

linux-amd64:
	@mkdir -p $(RELEASE_DIR)
	GOCACHE=$(GOCACHE) GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GO) build $(GOFLAGS) -ldflags "$(SERVER_LDFLAGS)" -o $(RELEASE_DIR)/$(SERVER_BIN)_linux_amd64 .
	GOCACHE=$(GOCACHE) GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GO) build $(GOFLAGS) -ldflags "$(AGENT_LDFLAGS)" -o $(RELEASE_DIR)/$(AGENT_BIN)_linux_amd64 ./cmd/nav-docker-agent

linux-arm64:
	@mkdir -p $(RELEASE_DIR)
	GOCACHE=$(GOCACHE) GOOS=linux GOARCH=arm64 CGO_ENABLED=0 $(GO) build $(GOFLAGS) -ldflags "$(SERVER_LDFLAGS)" -o $(RELEASE_DIR)/$(SERVER_BIN)_linux_arm64 .
	GOCACHE=$(GOCACHE) GOOS=linux GOARCH=arm64 CGO_ENABLED=0 $(GO) build $(GOFLAGS) -ldflags "$(AGENT_LDFLAGS)" -o $(RELEASE_DIR)/$(AGENT_BIN)_linux_arm64 ./cmd/nav-docker-agent

darwin-amd64:
	@mkdir -p $(RELEASE_DIR)
	GOCACHE=$(GOCACHE) GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 $(GO) build $(GOFLAGS) -ldflags "$(SERVER_LDFLAGS)" -o $(RELEASE_DIR)/$(SERVER_BIN)_darwin_amd64 .
	GOCACHE=$(GOCACHE) GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 $(GO) build $(GOFLAGS) -ldflags "$(AGENT_LDFLAGS)" -o $(RELEASE_DIR)/$(AGENT_BIN)_darwin_amd64 ./cmd/nav-docker-agent

darwin-arm64:
	@mkdir -p $(RELEASE_DIR)
	GOCACHE=$(GOCACHE) GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 $(GO) build $(GOFLAGS) -ldflags "$(SERVER_LDFLAGS)" -o $(RELEASE_DIR)/$(SERVER_BIN)_darwin_arm64 .
	GOCACHE=$(GOCACHE) GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 $(GO) build $(GOFLAGS) -ldflags "$(AGENT_LDFLAGS)" -o $(RELEASE_DIR)/$(AGENT_BIN)_darwin_arm64 ./cmd/nav-docker-agent

test:
	GOCACHE=$(GOCACHE) $(GO) test ./...

tidy:
	GOCACHE=$(GOCACHE) $(GO) mod tidy

clean:
	@rm -rf $(BIN_DIR)

run-server:
	GOCACHE=$(GOCACHE) $(GO) run .

run-agent:
	@if [ -z "$(TOKEN)" ]; then echo "missing TOKEN, usage: make run-agent TOKEN=nav_agent_xxx"; exit 1; fi
	GOCACHE=$(GOCACHE) $(GO) run ./cmd/nav-docker-agent -server "$(SERVER_URL)" -agent-guid "$(AGENT_GUID)" -token "$(TOKEN)"
