#!/bin/bash

# verify_manual_sql.sh
# This script runs manual_verification.sql and verifies the output matches expected results

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if server is running
if ! nc -z localhost 7205 > /dev/null 2>&1; then
    echo -e "${RED}✗ Server is not running on localhost:7205${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Server is running${NC}"

# Run SQL and capture output
echo -e "${YELLOW}Running manual_verification.sql...${NC}"
OUTPUT=$(nc localhost 7205 < test/manual_verification.sql 2>&1)

# Test 1: Multi-row INSERT - should return 3 rows
echo -n "Test 1: Multi-row INSERT... "
if echo "$OUTPUT" | grep -q "3 rows in set" | head -1; then
    echo -e "${GREEN}✓ PASSED${NC}"
else
    echo -e "${RED}✗ FAILED - Expected 3 rows${NC}"
    exit 1
fi

# Test 2: Numeric data integrity - should show correct values (100, 200, 300)
echo -n "Test 2: Numeric data integrity... "
if echo "$OUTPUT" | grep -q "| 1.*| 100" && echo "$OUTPUT" | grep -q "| 2.*| 200" && echo "$OUTPUT" | grep -q "| 3.*| 300"; then
    echo -e "${GREEN}✓ PASSED${NC}"
else
    echo -e "${RED}✗ FAILED - Numeric values incorrect${NC}"
    exit 1
fi

# Test 3: DELETE without WHERE - should show "Empty set"
echo -n "Test 3: DELETE without WHERE... "
if echo "$OUTPUT" | grep -c "Empty set" | grep -q "[1-9]"; then
    echo -e "${GREEN}✓ PASSED${NC}"
else
    echo -e "${RED}✗ FAILED - Expected empty set after DELETE${NC}"
    exit 1
fi

# Test 4: DROP TABLE - should complete without error
echo -n "Test 4: DROP TABLE... "
if echo "$OUTPUT" | grep -c "OK" | grep -q "[5-9]\|[1-9][0-9]"; then
    echo -e "${GREEN}✓ PASSED${NC}"
else
    echo -e "${RED}✗ FAILED - DROP TABLE failed${NC}"
    exit 1
fi

# Test 5: Index operations - should show idx_name
echo -n "Test 5: Index operations... "
if echo "$OUTPUT" | grep -q "idx_name"; then
    echo -e "${GREEN}✓ PASSED${NC}"
else
    echo -e "${RED}✗ FAILED - Index not found${NC}"
    exit 1
fi

# Test 6: Complex workflow - multiple operations
echo -n "Test 6: Complex workflow... "
if echo "$OUTPUT" | grep -q "5 rows in set" && echo "$OUTPUT" | grep -q "4 rows in set"; then
    echo -e "${GREEN}✓ PASSED${NC}"
else
    echo -e "${RED}✗ FAILED - Row counts don't match expected${NC}"
    exit 1
fi

echo ""
echo -e "${GREEN}✅ All manual verification tests passed!${NC}"
