.PHONY: all build build-og build-ogmo test test-v vet lint fmt clean install run

BINDIR    := bin
GO        := go
GOPATH    := $(shell go env GOPATH)

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

# Vet
vet:
	$(GO) vet ./...

# Lint (requires golangci-lint)
lint:
	golangci-lint run

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
	rm -rf $(BINDIR)
