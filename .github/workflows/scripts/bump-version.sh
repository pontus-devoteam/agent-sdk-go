#!/bin/bash

# This script helps create a new version tag based on semantic versioning.
# Usage: ./bump-version.sh [major|minor|patch]

set -e

# Validate input
if [ "$#" -ne 1 ] || ! [[ "$1" =~ ^(major|minor|patch)$ ]]; then
  echo "Usage: $0 [major|minor|patch]"
  exit 1
fi

# Get the latest tag
LATEST_TAG=$(git describe --tags $(git rev-list --tags --max-count=1 2>/dev/null) 2>/dev/null || echo "v0.0.0")
echo "Current version: $LATEST_TAG"

# Extract version components
VERSION=${LATEST_TAG#v}
MAJOR=$(echo $VERSION | cut -d. -f1)
MINOR=$(echo $VERSION | cut -d. -f2)
PATCH=$(echo $VERSION | cut -d. -f3)

# Bump the version according to the specified level
case "$1" in
  major)
    MAJOR=$((MAJOR + 1))
    MINOR=0
    PATCH=0
    ;;
  minor)
    MINOR=$((MINOR + 1))
    PATCH=0
    ;;
  patch)
    PATCH=$((PATCH + 1))
    ;;
esac

NEW_VERSION="v$MAJOR.$MINOR.$PATCH"
echo "New version: $NEW_VERSION"

# Create and push the new tag
read -p "Create and push tag $NEW_VERSION? (y/n) " -n 1 -r
echo    # move to a new line
if [[ $REPLY =~ ^[Yy]$ ]]; then
  git tag -a "$NEW_VERSION" -m "Release $NEW_VERSION"
  git push origin "$NEW_VERSION"
  echo "Tag $NEW_VERSION created and pushed."
else
  echo "Operation cancelled."
fi 