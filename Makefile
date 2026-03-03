.PHONY: build run dev install clean release

# Version (defaults to VERSION file, can be overridden: make build VERSION=1.0.0)
VERSION ?= $(shell cat VERSION 2>/dev/null || echo "dev")
DATE := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

# Build everything and run
all: build run

# Build frontend + Go binary
build:
	@echo "=== BUILDING v$(VERSION) ==="
	cd web && pnpm build
	go build -ldflags "-X github.com/lirrensi/looty/internal/server.Version=$(VERSION) -X github.com/lirrensi/looty/internal/server.BuildTime=$(DATE)" -o looty ./cmd/blip
	@echo "=== DONE ==="

# Run the server
run:
	@echo "=== STARTING ==="
	./looty

# Run web dev server
dev:
	cd web && pnpm dev

# Install web dependencies
install:
	cd web && pnpm install

# Clean build artifacts
clean:
	rm -f looty looty.html
	rm -rf web/dist

# Read version from VERSION file
VERSION_FILE := $(shell cat VERSION 2>/dev/null || echo "0.0.0")

# Create a new release (bump version in VERSION file first!)
release:
	@echo "=== CREATING RELEASE v$(VERSION) ==="
	@echo "Make sure you've updated VERSION file to the new version!"
	@read -p "Ready to create git tag v$(VERSION) and push? [y/N] " confirm; \
	if [ "$$confirm" = "y" ]; then \
		git add VERSION; \
		git commit -m "Release v$(VERSION)"; \
		git tag -a "v$(VERSION)" -m "Release v$(VERSION)"; \
		git push && git push --tags; \
		echo "=== PUSHED! CI will create the release ==="; \
	else \
		echo "Cancelled."; \
	fi
