#!/usr/bin/env bash
set -Eeuo pipefail

BUILD_GOOS=$(go env GOOS)
if [ -n "${GOOS+x}" ] && [ -n "$GOOS" ]; then
  BUILD_GOOS=$GOOS
fi

BUILD_GOARCH=$(go env GOARCH)
if [ -n "${GOARCH+x}" ] && [ -n "$GOARCH" ]; then
  BUILD_GOARCH=$GOARCH
fi

SERVICE_NAME="bookie_grpc";

# Get git ref, fallback to "unknown" if not in a git repo (e.g., Docker build)
if git describe --always > /dev/null 2>&1; then
  GIT_REF=$(git describe --always)
  VERSION=commit-$GIT_REF
else
  VERSION=docker-build
fi

for directory in ./src/cmd/* ; do
  component=$(basename $directory)
  out=./build/$component

  echo "building $component"
  echo "  → service: $SERVICE_NAME"
  echo "  → version: $VERSION"
  echo "  → output: $out"

  CGO_ENABLED=0 go build -o $out -v \
    -ldflags "-X main.version=$VERSION -X main.serviceName=$SERVICE_NAME -X main.componentName=$component" \
    $directory
done