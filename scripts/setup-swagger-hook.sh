#!/bin/bash
# Script to set up a git pre-commit hook for auto-generating Swagger docs

HOOK_FILE=".git/hooks/pre-commit"

echo "Setting up Swagger documentation pre-commit hook..."

# Create hooks directory if it doesn't exist
mkdir -p .git/hooks

# Create or append to pre-commit hook
cat > "$HOOK_FILE" << 'EOF'
#!/bin/bash
# Auto-generate Swagger documentation before commit

echo "Generating Swagger documentation..."

# Check if swag is installed
if ! command -v swag &> /dev/null; then
    echo "Installing swag..."
    go install github.com/swaggo/swag/cmd/swag@latest
fi

# Generate swagger docs
swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal

# Check if docs were modified
if ! git diff --quiet docs/; then
    echo "Swagger docs updated. Adding to commit..."
    git add docs/
fi

echo "Swagger documentation is up to date!"
EOF

# Make the hook executable
chmod +x "$HOOK_FILE"

echo "✓ Pre-commit hook installed successfully!"
echo "Swagger docs will now be auto-generated before each commit."
echo ""
echo "To disable: rm .git/hooks/pre-commit"
