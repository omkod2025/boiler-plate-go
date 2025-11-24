#!/bin/bash

# Script to change Go module name across the entire project
# Usage: ./scripts/change-module.sh <new-module-name>

set -e

if [ -z "$1" ]; then
    echo "Error: New module name is required"
    echo "Usage: ./scripts/change-module.sh <new-module-name>"
    echo ""
    echo "Example:"
    echo "  ./scripts/change-module.sh omgon-api-gateway"
    exit 1
fi

NEW_MODULE="$1"

# Get current module name from go.mod
if [ ! -f "go.mod" ]; then
    echo "Error: go.mod not found in current directory"
    exit 1
fi

CURRENT_MODULE=$(grep '^module ' go.mod | awk '{print $2}')

if [ -z "$CURRENT_MODULE" ]; then
    echo "Error: Could not find current module name in go.mod"
    exit 1
fi

if [ "$CURRENT_MODULE" = "$NEW_MODULE" ]; then
    echo "Error: New module name is the same as current module name"
    exit 1
fi

echo "Changing module name:"
echo "  From: $CURRENT_MODULE"
echo "  To:   $NEW_MODULE"
echo ""

# Step 1: Update go.mod
echo "Step 1: Updating go.mod..."
if sed --version >/dev/null 2>&1; then
    # GNU sed (Linux)
    sed -i "s|^module $CURRENT_MODULE|module $NEW_MODULE|" go.mod
else
    # BSD sed (macOS) or Git Bash on Windows
    sed -i.bak "s|^module $CURRENT_MODULE|module $NEW_MODULE|" go.mod
    rm -f go.mod.bak
fi
echo "✓ go.mod updated"
echo ""

# Step 2: Replace all import statements in .go files
echo "Step 2: Replacing import statements in .go files..."
find . -type f -name "*.go" ! -path "./vendor/*" | while read -r file; do
    if sed --version >/dev/null 2>&1; then
        # GNU sed (Linux)
        sed -i "s|\"$CURRENT_MODULE|\"$NEW_MODULE|g" "$file"
    else
        # BSD sed (macOS) or Git Bash on Windows
        sed -i.bak "s|\"$CURRENT_MODULE|\"$NEW_MODULE|g" "$file"
        rm -f "$file.bak"
    fi
done
echo "✓ All import statements updated"
echo ""

# Step 3: Run go mod tidy
echo "Step 3: Running 'go mod tidy'..."
go mod tidy
echo "✓ Dependencies cleaned up"
echo ""

echo "Successfully changed module from '$CURRENT_MODULE' to '$NEW_MODULE'"

