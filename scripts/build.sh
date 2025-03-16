#!/bin/bash

# Script for building the agent-sdk-go project with consistent flags
# This script helps with VCS stamping issues in various CI environments

# Set default build flags
BUILD_FLAGS="-buildvcs=false"

# Determine if additional flags should be passed
if [ "$1" == "--race" ]; then
  BUILD_FLAGS="$BUILD_FLAGS -race"
  shift
fi

if [ "$1" == "--verbose" ] || [ "$1" == "-v" ]; then
  BUILD_FLAGS="$BUILD_FLAGS -v"
  shift
fi

# Default target is all packages
TARGET="./..."
if [ -n "$1" ]; then
  TARGET="$1"
fi

# Echo the command for transparency
echo "Running: go build $BUILD_FLAGS $TARGET"

# Execute the build
go build $BUILD_FLAGS $TARGET
exit_code=$?

# Report build status
if [ $exit_code -eq 0 ]; then
  echo "Build successful!"
else
  echo "Build failed with exit code $exit_code"
fi

exit $exit_code 