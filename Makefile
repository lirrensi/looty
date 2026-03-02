.PHONY: build run dev install clean

# Build everything and run
all: build run

# Build frontend + Go binary
build:
	@echo "=== BUILDING ==="
	cd web && npm run build
	go build -ldflags "-X github.com/lirrensi/looty/internal/server.BuildTime=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')" -o looty ./cmd/blip
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