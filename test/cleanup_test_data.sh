#!/bin/bash
# Script to clean up all test data before running tests

set -e

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEST_DATA_DIR="${SCRIPT_DIR}/test_data"

echo "=== Cleaning Test Data ==="
echo "Test data directory: ${TEST_DATA_DIR}"

# Remove test_data directory if it exists
if [ -d "${TEST_DATA_DIR}" ]; then
    echo "Removing existing test data..."
    rm -rf "${TEST_DATA_DIR}"
    echo "✓ Test data cleaned"
else
    echo "No test data to clean"
fi

# Create fresh test_data directory
mkdir -p "${TEST_DATA_DIR}"
echo "✓ Fresh test_data directory created"

echo ""
echo "=== Test Data Cleanup Complete ==="
