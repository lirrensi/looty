.PHONY: run build dev install clean nuke web

# Default target - NUKE EVERYTHING, rebuild fresh, AND run
all: nuke web go run

# Build everything and run
build: nuke web go

# Just rebuild web and go (no nuke)
fresh: web go

# NUCLEAR OPTION - delete ALL build artifacts
nuke:
	@echo "=== NUKING ALL BUILD ARTIFACTS ==="
	rm -f looty.exe
	rm -f looty.html
	rm -rf web/dist
	rm -rf web/.vite
	rm -rf cmd/blip/.vite
	@echo "=== CLEAN COMPLETE ==="

# Build frontend HTML (Vite -> cmd/blip/)
web:
	@echo "=== BUILDING WEB FRONTEND ==="
	cd web && npm run build
	@echo "=== WEB BUILD COMPLETE ==="

# Build Go binary (embeds fresh index.html)
go:
	@echo "=== BUILDING GO BINARY ==="
	go build -o looty.exe ./cmd/blip
	@echo "=== GO BUILD COMPLETE ==="

# Run the server
run:
	@echo "=== STARTING SERVER ==="
	.\looty.exe

# Run web dev server
dev:
	cd web && npm run dev

# Install web dependencies
install:
	cd web && npm install

# Full setup: install deps and build
setup: install build

# Run both backend and frontend (requires two terminals)
start: run
	@echo "Server running at http://localhost:41111"