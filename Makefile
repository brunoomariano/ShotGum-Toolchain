BINARY  := stg
MODULE  := github.com/brunoomariano/ShotGum-Toolchain
CMD     := ./cmd/stg

GO      := go
GOFLAGS :=

# Inject version from the nearest git tag, e.g. v0.2.0.
# Falls back to DefaultVersion (internal/version/version.go) when there are no tags yet.
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -s -w -X '$(MODULE)/internal/version.Version=$(VERSION)'

COVERAGE_MIN := 85
COVERAGE_OUT := coverage.out

.PHONY: build run snapshot clean tidy fmt vet lint install uninstall test cover ci help

build: ## Build the binary in the current directory
	$(GO) build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o $(BINARY) $(CMD)

run: build ## Build and run the TUI
	./$(BINARY)

snapshot: ## Build release binaries for all platforms into ./dist
	@mkdir -p dist
	@for platform in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64; do \
		os=$$(echo $$platform | cut -d/ -f1); \
		arch=$$(echo $$platform | cut -d/ -f2); \
		ext=$$([ "$$os" = "windows" ] && echo ".exe" || echo ""); \
		out="dist/$(BINARY)-$$os-$$arch$$ext"; \
		printf "  → $$out\n"; \
		GOOS=$$os GOARCH=$$arch $(GO) build \
			-ldflags="$(LDFLAGS)" \
			-o "$$out" $(CMD); \
	done
	@printf "  ✓ dist/ ready ($(VERSION))\n"

install: build ## Install the binary to ~/.local/bin
	install -Dm755 $(BINARY) ~/.local/bin/$(BINARY)

uninstall: ## Remove the binary from ~/.local/bin
	rm -f ~/.local/bin/$(BINARY)

tidy: ## Tidy go.mod and go.sum
	$(GO) mod tidy

fmt: ## Format all Go source files
	$(GO) fmt ./...

vet: ## Run go vet
	$(GO) vet ./...

lint: fmt vet ## Run fmt and vet

test: ## Run tests with race detector
	$(GO) test -race -count=1 ./...

cover: ## Run tests and generate coverage report
	$(GO) test -race -count=1 -coverprofile=$(COVERAGE_OUT) -covermode=atomic ./...
	$(GO) tool cover -func=$(COVERAGE_OUT) | tail -5

ci: ## Format check, vet, tests and coverage >= COVERAGE_MIN%
	@echo "  → Format check..."
	@test -z "$$(gofmt -l .)" && echo "  ✓ Format OK" || { echo "  ✗ Run 'make fmt' to fix"; exit 1; }
	@echo "  → Vet..."
	@$(GO) vet ./... && echo "  ✓ Vet OK"
	@echo "  → Tests..."
	@$(GO) test -race -count=1 -coverprofile=$(COVERAGE_OUT) -covermode=atomic ./...
	@echo "  → Coverage (min $(COVERAGE_MIN)%)..."
	@$(GO) tool cover -func=$(COVERAGE_OUT) | awk \
	  '/^total:/{gsub(/%/,"",$$3); pct=$$3+0; \
	   if(pct<$(COVERAGE_MIN)){printf "  ✗ %.1f%% < $(COVERAGE_MIN)%%\n",pct; exit 1} \
	   else {printf "  ✓ %.1f%%\n",pct}}'
	@echo "  ✓ CI passed"

clean: ## Remove the local binary and dist/
	rm -f $(BINARY)
	rm -rf dist/

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*##' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*##"}; {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'
