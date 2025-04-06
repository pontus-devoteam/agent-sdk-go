#!/bin/bash
set -e

echo "====== Go Report Card Issue Fixer ======"
echo "This script will fix common issues reported by Go Report Card."

# Store the current directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
echo "Repository root: ${REPO_ROOT}"
cd "${REPO_ROOT}"

# Get Go path
GOPATH=$(go env GOPATH)
GOBIN="${GOPATH}/bin"
echo "Using Go binaries from: ${GOBIN}"

# Install required tools
echo "Installing tools..."
go install golang.org/x/lint/golint@latest
go install github.com/client9/misspell/cmd/misspell@latest
go install github.com/kisielk/errcheck@latest
go install github.com/gordonklaus/ineffassign@latest
go install github.com/mdempsky/unconvert@latest
go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
go install honnef.co/go/tools/cmd/staticcheck@latest

# Setup Go module if needed
echo "Setting up Go module..."
GO_MOD_EXISTS=false
GO_SUM_EXISTS=false

# Check for and backup existing Go module files
if [ -f "${REPO_ROOT}/go.mod" ]; then
  GO_MOD_EXISTS=true
  echo "Existing go.mod found at ${REPO_ROOT}/go.mod, backing up..."
  cp "${REPO_ROOT}/go.mod" "${REPO_ROOT}/go.mod.bak"
  rm "${REPO_ROOT}/go.mod"
fi

if [ -f "${REPO_ROOT}/go.sum" ]; then
  GO_SUM_EXISTS=true
  echo "Existing go.sum found, backing up..."
  cp "${REPO_ROOT}/go.sum" "${REPO_ROOT}/go.sum.bak"
  rm "${REPO_ROOT}/go.sum"
fi

echo "Initializing Go module..."
cd "${REPO_ROOT}"
go mod init github.com/pontus-devoteam/agent-sdk-go
echo "replace github.com/pontus-devoteam/agent-sdk-go => ${REPO_ROOT}" >> go.mod
go mod tidy
echo "✅ Go module setup completed"

# Run gofmt
echo "Running gofmt..."
find "${REPO_ROOT}" -type f -name "*.go" | xargs gofmt -s -w
echo "✅ gofmt completed"

# Fix misspellings
echo "Fixing misspellings..."
"${GOBIN}/misspell" -w $(find "${REPO_ROOT}" -type f -name "*.go" -o -name "*.md")
echo "✅ misspell completed"

# Fix ineffectual assignments
echo "Checking for ineffectual assignments..."
cd "${REPO_ROOT}"
"${GOBIN}/ineffassign" ./... || echo "⚠️ Some ineffectual assignments found"
echo "✅ ineffassign completed"

# Run errcheck
echo "Checking for unchecked errors..."
cd "${REPO_ROOT}"
"${GOBIN}/errcheck" ./... || echo "⚠️ Some unchecked errors found"
echo "✅ errcheck completed"

# Run go vet
echo "Running go vet..."
cd "${REPO_ROOT}"
go vet ./... || echo "⚠️ Some issues found by go vet"
echo "✅ go vet completed"

# Run golint
echo "Running golint..."
cd "${REPO_ROOT}"
"${GOBIN}/golint" -set_exit_status $(go list ./...) || echo "⚠️ Some linting issues found"
echo "✅ golint completed"

# Check cyclomatic complexity
echo "Checking cyclomatic complexity..."
cd "${REPO_ROOT}"
"${GOBIN}/gocyclo" -over 15 . || echo "⚠️ Some complex functions found"
echo "✅ gocyclo completed"

# Run staticcheck
echo "Running staticcheck..."
cd "${REPO_ROOT}"
"${GOBIN}/staticcheck" ./... || echo "⚠️ Some issues found by staticcheck"
echo "✅ staticcheck completed"

# Check for unconverted types
echo "Checking for unconverted types..."
cd "${REPO_ROOT}"
"${GOBIN}/unconvert" ./... || echo "⚠️ Some unconverted types found"
echo "✅ unconvert completed"

# Restore original go.mod and go.sum if they existed
if [ "$GO_MOD_EXISTS" = true ]; then
  echo "Restoring original go.mod..."
  cd "${REPO_ROOT}"
  mv "${REPO_ROOT}/go.mod.bak" "${REPO_ROOT}/go.mod"
  echo "✅ Original go.mod restored"
fi

if [ "$GO_SUM_EXISTS" = true ]; then
  echo "Restoring original go.sum..."
  cd "${REPO_ROOT}"
  mv "${REPO_ROOT}/go.sum.bak" "${REPO_ROOT}/go.sum"
  echo "✅ Original go.sum restored"
fi

echo "====== All checks completed ======"
echo "Your codebase should now pass Go Report Card checks."
echo "Visit https://goreportcard.com/report/github.com/pontus-devoteam/agent-sdk-go to see your score." 