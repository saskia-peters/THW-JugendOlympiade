# justfile — THW-JugendOlympiade task runner
# Requires: just (https://just.systems), uv (https://astral.sh/uv), wails, go
#
# Usage:
#   just           → list all available recipes
#   just build     → production build
#   just test      → run all Go tests
#   just docs      → build documentation into docs/
#   just docs-serve → start live-reload docs server

set windows-shell := ["cmd.exe", "/c"]

uv    := "D:/saskia/develop/uv.exe"
wails := "wails"

# ─── Default: list recipes ────────────────────────────────────────────────────
default:
    @just --list

# ─── Go / Wails ───────────────────────────────────────────────────────────────

# Run the application in development mode with hot-reload
dev:
    {{wails}} dev

# Build a production binary (output: build/bin/)
build:
    {{wails}} build

# Build without a console window (Windows only)
build-noconsole:
    {{wails}} build -ldflags "-H windowsgui"

# Run all Go tests with verbose output
test:
    cd test && go test -v ./...

# Run tests with coverage report
test-cover:
    cd test && go test -cover ./...

# Download Go module dependencies
deps:
    go mod download

# ─── Documentation ────────────────────────────────────────────────────────────

# Install/sync Python documentation dependencies
docs-install:
    {{uv}} sync --project assets

# Build documentation from assets/mkdocs/ → docs/
docs: docs-install
    {{uv}} run --project assets mkdocs build --config-file assets/mkdocs.yml --strict

# Start a local live-reload documentation server (http://localhost:8000)
docs-serve: docs-install
    {{uv}} run --project assets mkdocs serve --config-file assets/mkdocs.yml

# Build docs and report any warnings
docs-check: docs-install
    {{uv}} run --project assets mkdocs build --config-file assets/mkdocs.yml --strict --verbose

# ─── Combined ─────────────────────────────────────────────────────────────────

# Run tests and build everything (app + docs)
all: test build docs
    @echo "All done."
