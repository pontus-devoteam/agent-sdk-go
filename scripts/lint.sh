#!/bin/bash

set -e

echo "Running Go formatting checks..."
# Check if code is properly formatted
GOFMT_FILES=$(gofmt -l .)
if [ -n "$GOFMT_FILES" ]; then
  echo "The following files need to be formatted with gofmt:"
  echo "$GOFMT_FILES"
  echo "Please run: gofmt -w ."
  exit 1
fi
echo "✅ Go formatting check passed"

echo "Running Go imports checks..."
# Check if imports are properly organized
GOIMPORTS_PATH="$HOME/go/bin/goimports"
if [ -f "$GOIMPORTS_PATH" ]; then
  GOIMPORTS_FILES=$("$GOIMPORTS_PATH" -l .)
  if [ -n "$GOIMPORTS_FILES" ]; then
    echo "The following files need to be formatted with goimports:"
    echo "$GOIMPORTS_FILES"
    echo "Please run: $GOIMPORTS_PATH -w ."
    exit 1
  fi
  echo "✅ Go imports check passed"
else
  echo "⚠️ goimports not found at $GOIMPORTS_PATH, skipping imports check"
  echo "To install: go install golang.org/x/tools/cmd/goimports@latest"
fi

echo "Running Go vet..."
# Run go vet to catch common errors
go vet -buildvcs=false ./...
echo "✅ Go vet check passed"

echo "Running Go build..."
# Ensure the code compiles
go build -buildvcs=false ./...
echo "✅ Go build check passed"

echo "All checks passed! ✅" 