#!/bin/bash

set -e

if [ $# -ne 1 ]; then
    echo "Usage: $0 <version>"
    exit 1
fi

VERSION="$1"

if [[ ! $VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Invalid version format. Use v1.2.3 format."
    exit 1
fi

echo "Checking for uncommitted changes..."
if [[ -n $(git status --porcelain) ]]; then
    echo "Uncommitted changes detected. Please commit or stash them before tagging."
    exit 1
fi

echo "Checking if on main branch..."
BRANCH=$(git rev-parse --abbrev-ref HEAD)
if [[ "$BRANCH" != "main" ]]; then
    echo "You must be on the main branch to tag a release (current: $BRANCH)."
    exit 1
fi

git tag "$VERSION"
git push origin "$VERSION"

echo "Tag $VERSION pushed successfully."
