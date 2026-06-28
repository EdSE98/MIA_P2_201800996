#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

mapfile -d '' GO_FILES < <(
  find . \
    -path './MIA_Junio2026-main' -prune -o \
    -path './web/node_modules' -prune -o \
    -type f -name '*.go' -print0
)

UNFORMATTED="$(gofmt -l "${GO_FILES[@]}")"
if [[ -n "$UNFORMATTED" ]]; then
  echo "Archivos Go sin formato:"
  echo "$UNFORMATTED"
  exit 1
fi

export GOCACHE="${GOCACHE:-/tmp/go-build-cache}"

echo "==> go test ./..."
go test ./...

echo "==> go vet ./..."
go vet ./...

echo "==> npm run build"
(
  cd web
  npm run build
)

echo "Validacion completada."
