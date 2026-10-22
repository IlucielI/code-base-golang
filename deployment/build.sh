#!/usr/bin/env sh
set -eu

BASE_IMAGE="code-base-golang-api-base:latest"
RUNTIME_IMAGE="code-base-golang-api-runtime:latest"
REBUILD_BASE="${REBUILD_BASE:-false}"
REBUILD_RUNTIME="${REBUILD_RUNTIME:-false}"

# Check if base image exists or if forced rebuild is requested
if [ "$REBUILD_BASE" = "true" ] || ! docker image inspect "$BASE_IMAGE" >/dev/null 2>&1; then
  echo "==> Base image '$BASE_IMAGE' not found (or REBUILD_BASE=true). Building base image..."
  ./deployment/build-base.sh
else
  echo "==> Base image '$BASE_IMAGE' already exists. Skipping base build."
fi

# Check if runtime base image exists or if forced rebuild is requested
if [ "$REBUILD_RUNTIME" = "true" ] || ! docker image inspect "$RUNTIME_IMAGE" >/dev/null 2>&1; then
  echo "==> Runtime base image '$RUNTIME_IMAGE' not found (or REBUILD_RUNTIME=true). Building runtime base image..."
  ./deployment/build-runtime.sh
else
  echo "==> Runtime base image '$RUNTIME_IMAGE' already exists. Skipping runtime base build."
fi

echo "==> Building application image..."
./deployment/build-api.sh
