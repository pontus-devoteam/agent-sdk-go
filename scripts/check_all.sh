#!/bin/bash

set -e

echo "Running all checks for agent-sdk-go..."

# Check Go version
echo "=== Checking Go version ==="
./scripts/check_go_version.sh

# Run lint checks
echo "=== Running lint checks ==="
./scripts/lint.sh

# Run security checks
echo "=== Running security checks ==="
./scripts/security_check.sh

# Run tests
echo "=== Running tests ==="
cd test && make ci-test
cd ..

echo "All checks passed! ✅"
echo "The codebase is in good shape and ready for review/merge." 