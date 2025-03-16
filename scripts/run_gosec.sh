#!/bin/bash

# Script for running gosec with consistent flags in CI environments
# This ensures proper handling of VCS stamping issues

# Set default environment variables
export GOFLAGS="-buildvcs=false"

# Try to find the gosec binary
GOSEC=$(which gosec)
if [ -z "$GOSEC" ]; then
  # Try common installation paths
  if [ -f "$HOME/go/bin/gosec" ]; then
    GOSEC="$HOME/go/bin/gosec"
  elif [ -f "/usr/local/bin/gosec" ]; then
    GOSEC="/usr/local/bin/gosec"
  else
    echo "Error: gosec not found in PATH"
    exit 1
  fi
fi

# Parse additional arguments
ARGS=""
for arg in "$@"; do
  ARGS="$ARGS $arg"
done

# If no args provided, default to scanning the entire project
if [ -z "$ARGS" ]; then
  ARGS="./..."
fi

# First attempt - with exclude options
echo "Running: $GOSEC -exclude-dir=examples -exclude=G104 $ARGS"
$GOSEC -exclude-dir=examples -exclude=G104 $ARGS
exit_code=$?

# If first attempt fails, try with environment variable only
if [ $exit_code -ne 0 ]; then
  echo "Retrying security scan with environment variable approach..."
  GOFLAGS="-buildvcs=false" $GOSEC -exclude-dir=examples $ARGS
  exit_code=$?
fi

# If second attempt fails, try with build tags
if [ $exit_code -ne 0 ]; then
  echo "Retrying security scan with tags approach..."
  $GOSEC -tags buildmode=exe -exclude-dir=examples $ARGS
  exit_code=$?
fi

# Final fallback - basic run
if [ $exit_code -ne 0 ]; then
  echo "Retrying security scan with minimal options..."
  $GOSEC $ARGS
  exit_code=$?
fi

# Report scan status
if [ $exit_code -eq 0 ]; then
  echo "Security scan completed successfully!"
else
  echo "Security scan failed with exit code $exit_code"
fi

exit $exit_code 