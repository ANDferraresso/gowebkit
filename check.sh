#!/usr/bin/env bash
set -e

echo "🔍 Code formatting (go fmt)..."
go fmt ./...

echo "🔍 Static analysis (go vet)..."
go vet ./...

echo "🔍 Compiling (go build)..."
go build ./...

echo "🔍  Running test (go test)..."
go test ./...

if command -v golangci-lint >/dev/null 2>&1; then
    echo "🔍 Advanced lint analysis (golangci-lint)..."
    golangci-lint run ./...
else
    echo "⚠️ golangci-lint is not installed. You can install it with:"
    echo "   brew install golangci-lint"
fi

echo "✅ Checks completed successfully!"