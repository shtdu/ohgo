.PHONY: all build build-og build-ogmo test test-v test-pkg coverage vet lint check ci fmt clean install run

BINDIR    := bin
GO        := go
GOPATH    := $(shell go env GOPATH)
GOLANGCI_LINT ?= golangci-lint
LINT_FLAGS    ?=
COVERAGE_FILE ?= coverage.out
COVERAGE_HTML ?= coverage.html

all: build

# Build both binaries
build: build-og build-ogmo

build-og: | $(BINDIR)
	$(GO) build -o $(BINDIR)/og ./cmd/og

build-ogmo: | $(BINDIR)
	$(GO) build -o $(BINDIR)/ogmo ./cmd/ogmo

$(BINDIR):
	mkdir -p $(BINDIR)

# Run all tests
test:
	$(GO) test ./...

# Verbose tests
test-v:
	$(GO) test -v ./...

# Run a single package's tests (usage: make test-pkg PKG=./internal/engine)
test-pkg:
	$(GO) test -v $(PKG)

# Run all tests and generate coverage reports
coverage:
	$(GO) test -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	$(GO) tool cover -func=$(COVERAGE_FILE) | tail -1
	$(GO) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)

# Vet
vet:
	$(GO) vet ./...

# Lint (requires golangci-lint)
lint:
	$(GOLANGCI_LINT) run $(LINT_FLAGS)

# Run all local verification checks
check: build lint test

# CI entrypoint
ci: check

# Format
fmt:
	gofmt -w .

# Install both binaries to GOPATH/bin
install: build
	cp $(BINDIR)/og $(GOPATH)/bin/
	cp $(BINDIR)/ogmo $(GOPATH)/bin/

# Run the main binary
run: build-og
	$(BINDIR)/og

# Clean build artifacts
clean:
	rm -rf $(BINDIR) $(COVERAGE_FILE) $(COVERAGE_HTML)
