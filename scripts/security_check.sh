#!/bin/bash

set -e

echo "Running security checks with gosec..."

GOSEC_PATH="$HOME/go/bin/gosec"
if [ -f "$GOSEC_PATH" ]; then
  # Run gosec with basic configuration, excluding examples
  "$GOSEC_PATH" -quiet -exclude-dir=examples ./...
  echo "✅ Security check passed"
else
  echo "⚠️ gosec not found at $GOSEC_PATH, skipping security check"
  echo "To install: go install github.com/securego/gosec/v2/cmd/gosec@latest"
  exit 1
fi

echo "All security checks passed! ✅" 