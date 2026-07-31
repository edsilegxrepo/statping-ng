#!/bin/bash
# Test runner script for statping-ng
# Runs all tests with package-level serialization (-p=1) for reliability

echo "=== Running statping-ng test suite ==="
echo ""

# Run all tests with:
# -p=1: One package at a time (prevents shared state conflicts)
# -count=1: No test caching
# -timeout=600s: 10 minute timeout for full suite
go test ./... -p=1 -count=1 -timeout=600s

exit $?
