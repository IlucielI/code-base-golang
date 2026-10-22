#!/usr/bin/env sh
set -eu

docker build \
  -f deployment/Dockerfile.runtime \
  -t code-base-golang-api-runtime:latest \
  .
