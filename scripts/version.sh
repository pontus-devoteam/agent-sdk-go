#!/bin/bash

# This script helps create a new version tag based on semantic versioning.
# Usage: ./version.sh bump

set -e

# Function to display usage
usage() {
    echo "Usage: $0 bump"
    echo "  bump    Bump the version and create a new tag"
    exit 1
}

# Check if command is provided
if [ "$#" -ne 1 ] || [ "$1" != "bump" ]; then
    usage
fi

# Get the latest tag
LATEST_TAG=$(git describe --tags $(git rev-list --tags --max-count=1 2>/dev/null) 2>/dev/null || echo "v0.0.0")
echo "Current version: $LATEST_TAG"

# Extract version components
VERSION=${LATEST_TAG#v}
MAJOR=$(echo $VERSION | cut -d. -f1)
MINOR=$(echo $VERSION | cut -d. -f2)
PATCH=$(echo $VERSION | cut -d. -f3)

# Ask which version component to bump
echo "Which version component would you like to bump?"
echo "1) Major (current: $MAJOR)"
echo "2) Minor (current: $MINOR)"
echo "3) Patch (current: $PATCH)"
read -p "Enter your choice (1-3): " choice

case "$choice" in
    1)
        MAJOR=$((MAJOR + 1))
        MINOR=0
        PATCH=0
        ;;
    2)
        MINOR=$((MINOR + 1))
        PATCH=0
        ;;
    3)
        PATCH=$((PATCH + 1))
        ;;
    *)
        echo "Invalid choice. Please select 1, 2, or 3."
        exit 1
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