.PHONY: build run dev install clean

# Version (can be overridden: make build VERSION=1.0.0)
VERSION ?= dev
DATE := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

# Build everything and run
all: build run

# Build frontend + Go binary
build:
	@echo "=== BUILDING v$(VERSION) ==="
	cd web && npm run build
	go build -ldflags "-X github.com/lirrensi/looty/internal/server.Version=$(VERSION) -X github.com/lirrensi/looty/internal/server.BuildTime=$(DATE)" -o looty ./cmd/blip
	@echo "=== DONE ==="

# Run the server
run:
	@echo "=== STARTING ==="
	./looty

# Run web dev server
dev:
	cd web && npm run dev

# Install web dependencies
install:
	cd web && npm install

# Clean build artifacts
clean:
	rm -f looty looty.html
	rm -rf web/dist
