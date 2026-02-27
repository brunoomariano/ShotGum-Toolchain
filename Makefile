BINARY   := stg
MODULE   := github.com/brunoomariano/ShotGum-Toolchain
CMD      := ./cmd/stg

GO       := go
GOFLAGS  :=
SRC      := src
BUILDDIR := build
UNAME_S  := $(shell uname -s)

ifeq ($(UNAME_S),Linux)
	IS_LINUX := true
else
	IS_LINUX := false
endif

# Read version directly from internal/version/version.go (single source of truth).
VERSION := $(shell sed -n 's/^const DefaultVersion = "\(.*\)"/\1/p' $(SRC)/internal/version/version.go | head -n1)
LDFLAGS := -s -w

COVERAGE_MIN := 85
COVERAGE_OUT := coverage.out

.PHONY: build run snapshot clean tidy fmt vet lint install uninstall test cover ci help

build: ## Build the binary into ./build/
	@if [ "$(IS_LINUX)" != "true" ]; then echo "  ✗ Linux only (Debian/Ubuntu)."; exit 1; fi
	@mkdir -p $(CURDIR)/$(BUILDDIR)
	$(GO) -C $(SRC) build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o $(CURDIR)/$(BUILDDIR)/$(BINARY) $(CMD)

run: build ## Build and run the TUI
	./$(BUILDDIR)/$(BINARY)

snapshot: ## Build release binaries for Linux only into ./dist
	@if [ "$(IS_LINUX)" != "true" ]; then echo "  ✗ Linux only (Debian/Ubuntu)."; exit 1; fi
	@mkdir -p $(CURDIR)/dist
	@for platform in linux/amd64 linux/arm64; do \
		os=$$(echo $$platform | cut -d/ -f1); \
		arch=$$(echo $$platform | cut -d/ -f2); \
		out="$(CURDIR)/dist/$(BINARY)-$$os-$$arch"; \
		printf "  → $$out\n"; \
		GOOS=$$os GOARCH=$$arch $(GO) -C $(SRC) build \
			-ldflags="$(LDFLAGS)" \
			-o "$$out" $(CMD); \
	done
	@printf "  ✓ dist/ ready ($(VERSION))\n"

install: build ## Install the binary to ~/.local/bin
	install -Dm755 $(CURDIR)/$(BUILDDIR)/$(BINARY) ~/.local/bin/$(BINARY)

uninstall: ## Remove the binary from ~/.local/bin
	rm -f ~/.local/bin/$(BINARY)

tidy: ## Tidy go.mod and go.sum
	$(GO) -C $(SRC) mod tidy

fmt: ## Format all Go source files
	$(GO) -C $(SRC) fmt ./...

vet: ## Run go vet
	$(GO) -C $(SRC) vet ./...

lint: fmt vet ## Run fmt and vet

test: ## Run tests with race detector
	$(GO) -C $(SRC) test -race -count=1 ./...

cover: ## Run tests and generate coverage report
	$(GO) -C $(SRC) test -race -count=1 -coverprofile=$(CURDIR)/$(COVERAGE_OUT) -covermode=atomic ./...
	$(GO) -C $(SRC) tool cover -func=$(CURDIR)/$(COVERAGE_OUT) | tail -5

ci: ## Format check, vet, tests and coverage >= COVERAGE_MIN%
	@echo "  → Format check..."
	@test -z "$$(gofmt -l ./$(SRC))" && echo "  ✓ Format OK" || { echo "  ✗ Run 'make fmt' to fix"; exit 1; }
	@echo "  → Vet..."
	@$(GO) -C $(SRC) vet ./... && echo "  ✓ Vet OK"
	@echo "  → Tests..."
	@$(GO) -C $(SRC) test -race -count=1 -coverprofile=$(CURDIR)/$(COVERAGE_OUT) -covermode=atomic ./...
	@echo "  → Coverage (min $(COVERAGE_MIN)%)..."
	@$(GO) -C $(SRC) tool cover -func=$(CURDIR)/$(COVERAGE_OUT) | awk \
	  '/^total:/{gsub(/%/,"",$$3); pct=$$3+0; \
	   if(pct<$(COVERAGE_MIN)){printf "  ✗ %.1f%% < $(COVERAGE_MIN)%%\n",pct; exit 1} \
	   else {printf "  ✓ %.1f%%\n",pct}}'
	@echo "  ✓ CI passed"

clean: ## Remove ./build/ and ./dist/
	rm -rf $(CURDIR)/$(BUILDDIR) $(CURDIR)/dist/

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*##' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*##"}; {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'
