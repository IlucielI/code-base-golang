#!/usr/bin/env sh
set -eu

docker build \
  -f deployment/Dockerfile.base \
  -t code-base-golang-api-base:latest \
  .
