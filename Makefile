.PHONY: run build dev install clean

# Default target - build AND run!
all: build run

# Build AND run in one command
run: build
	./blip.exe

# Build BOTH frontend HTML AND Go binary
build: build-web build-go

# Build frontend HTML (Vite -> embed/)
build-web:
	cd web && pnpm run build

# Build Go binary
build-go:
	go build -o blip.exe ./cmd/blip

# Run web dev server
dev:
	cd web && pnpm dev

# Install web dependencies
install:
	cd web && pnpm install

# Clean build artifacts
clean:
	rm -f blip.exe

# Full setup: install deps and build
setup: install build

# Run both backend and frontend (requires two terminals)
start: run
	@echo "Server running at http://localhost:41111"