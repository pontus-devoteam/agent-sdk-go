#!/bin/bash

set -e

# CI Setup Script to ensure consistent Go version and toolchain settings
# This helps with VCS stamping issues and version mismatches in CI environments

echo "Setting up CI environment..."

# Check Go version
echo "Checking Go version..."
./scripts/check_go_version.sh || {
  echo "Go version check failed. Please install Go 1.23 or later."
  exit 1
}

# Check Go version
GO_VERSION=$(go version | awk '{print $3}')
echo "Detected Go version: $GO_VERSION"

echo "Updating go.mod file to ensure compatibility..."
# Remove toolchain directive
sed -i '' '/^toolchain/d' go.mod || true
# Change Go version from 1.24.0 to 1.23.0
sed -i '' 's/go 1.24.0/go 1.23.0/g' go.mod || true
# Change Go version from 1.23.4 to 1.23.0
sed -i '' 's/go 1.23.4/go 1.23.0/g' go.mod || true

echo "go.mod after update:"
cat go.mod

echo "Installing required tools..."

# Define paths
GOBIN=$(go env GOPATH)/bin
export PATH=$GOBIN:$PATH

echo "Installing golangci-lint v1.54.2..."
# Install specific version of golangci-lint that works with Go 1.23
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.54.2

echo "Installing gosec latest..."
# Install gosec
go install github.com/securego/gosec/v2/cmd/gosec@latest

echo "Installing goimports..."
# Install goimports
go install golang.org/x/tools/cmd/goimports@latest

echo "Verifying tool installation..."
echo "golangci-lint version:"
which golangci-lint && golangci-lint --version || echo "golangci-lint not found in PATH"

echo "gosec version:"
which gosec && gosec --version || echo "gosec not found in PATH"

echo "goimports version:"
which goimports && echo "goimports installed" || echo "goimports not found in PATH"

# Set environment variables to avoid VCS stamping issues
export GOFLAGS="-buildvcs=false"

echo "Go environment:"
echo "$GO_VERSION"
echo "-buildvcs=false"

echo "CI setup complete!" 