#!/bin/bash

# Script to diagnose linting issues

echo "===== SYSTEM INFO ====="
go version
uname -a

echo "===== WORKSPACE DETAILS ====="
echo "PWD: $(pwd)"

echo "===== DIRECTORY STRUCTURE ====="
ls -la

echo "===== GO FILES ====="
find . -name "*.go" | sort
echo "Total Go files: $(find . -name "*.go" | wc -l)"

echo "===== MODULE FILES ====="
find . -name "go.mod" | sort
cat $(find . -name "go.mod" | head -1)

echo "===== PKG DIRECTORY ====="
find . -path "*/pkg/*" | sort | head -10

echo "===== GO LIST OUTPUT ====="
go list -m all || echo "go list failed"
go list ./... || echo "go list ./... failed"

echo "===== RUNNING LINTER ====="
export GO111MODULE=on
# Try with verbose option
golangci-lint run --verbose || echo "golangci-lint failed"

echo "===== BASIC GO CHECKS ====="
# Try go vet
go vet ./... || echo "go vet found issues" 