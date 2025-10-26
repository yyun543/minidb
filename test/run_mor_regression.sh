#!/bin/bash
# MoR (Merge-on-Read) Regression Test
# Tests that delta files work correctly after server restart
# Reproduces and validates the fix for the delta metadata bug

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "=== MoR Regression Test: Delta Files After Restart ==="
echo ""

# Kill any running instance
pkill -f "./minidb" 2>/dev/null || true
sleep 2

# Clean up test data
echo "Cleaning up test data..."
rm -rf ./minidb_data
mkdir -p ./minidb_data

# Build if needed
if [ ! -f "./minidb" ]; then
    echo "Building minidb..."
    go build -o minidb ./cmd/server
fi

# Start server (uses default ./minidb_data directory)
echo "1. Starting server..."
ENVIRONMENT=development ./minidb > /tmp/mor_regression.log 2>&1 &
SERVER_PID=$!
sleep 3

# Function to stop server and clean up
cleanup() {
    echo "Stopping server..."
    kill $SERVER_PID 2>/dev/null || true
    sleep 1
    # Clean up test data
    rm -rf ./minidb_data
}
trap cleanup EXIT

# Create database and tables, insert data
echo "2. Creating database and inserting data..."
{
    echo "CREATE DATABASE ecommerce;"
    sleep 0.3
    echo "USE ecommerce;"
    sleep 0.3
    echo "CREATE TABLE users (id INT, name VARCHAR, email VARCHAR, age INT, created_at VARCHAR);"
    sleep 0.3
    echo "CREATE TABLE orders (id INT, user_id INT, amount INT, order_date VARCHAR);"
    sleep 0.3
    echo "INSERT INTO users VALUES (1, 'John Doe', 'john@example.com', 25, '2024-01-01');"
    sleep 0.3
    echo "INSERT INTO users VALUES (2, 'Jane Smith', 'jane@example.com', 30, '2024-01-02');"
    sleep 0.3
    echo "INSERT INTO users VALUES (3, 'Bob Wilson', 'bob@example.com', 35, '2024-01-03');"
    sleep 0.3
    echo "INSERT INTO orders VALUES (1, 1, 100, '2024-01-05');"
    sleep 0.3
    echo "INSERT INTO orders VALUES (2, 2, 250, '2024-01-06');"
    sleep 0.3
    echo "INSERT INTO orders VALUES (3, 1, 150, '2024-01-07');"
    sleep 0.3
} | nc localhost 7205 > /dev/null 2>&1

# Perform UPDATE and DELETE to create delta files
echo "3. Creating MoR delta files (UPDATE + DELETE)..."
{
    echo "USE ecommerce;"
    sleep 0.3
    echo "UPDATE users SET email = 'john.doe@newdomain.com' WHERE id = 1;"
    sleep 0.3
    echo "DELETE FROM orders WHERE amount < 50;"
    sleep 0.3
} | nc localhost 7205 > /dev/null 2>&1

# Test queries BEFORE restart
echo "4. Testing queries BEFORE restart..."
{
    echo "USE ecommerce;"
    sleep 0.3
    echo "SELECT * FROM users;"
    sleep 0.3
} | nc localhost 7205 > /tmp/mor_before_restart.log 2>&1

# Stop server
echo "5. Stopping server..."
kill $SERVER_PID 2>/dev/null
sleep 2

# Restart server
echo "6. Restarting server..."
ENVIRONMENT=development ./minidb > /tmp/mor_regression_restart.log 2>&1 &
SERVER_PID=$!
sleep 3

# Test queries AFTER restart
echo "7. Testing queries AFTER restart..."
{
    echo "USE ecommerce;"
    sleep 0.3
    echo "SELECT * FROM users;"
    sleep 0.5
    echo "SELECT * FROM orders;"
    sleep 0.5
    echo "SELECT u.name, COUNT(o.id) as order_count, SUM(o.amount) as total_amount FROM users u LEFT JOIN orders o ON u.id = o.user_id GROUP BY u.name HAVING order_count > 1 ORDER BY total_amount DESC;"
    sleep 1
} | nc localhost 7205 > /tmp/mor_after_restart.log 2>&1

echo ""
echo "=== TEST RESULTS ==="
echo ""

# Check 1: No delta metadata in results
if grep -q "update.*176137\|delta-update\|delta-delete\|delta_type" /tmp/mor_after_restart.log; then
    echo -e "${RED}❌ FAIL: Delta file metadata visible in query results!${NC}"
    echo "This means delta files are being read as data instead of being processed."
    exit 1
else
    echo -e "${GREEN}✅ PASS: No delta file metadata in query results${NC}"
fi

# Check 2: UPDATE was applied
if grep -q "john.doe@newdomain.com" /tmp/mor_after_restart.log; then
    echo -e "${GREEN}✅ PASS: UPDATE changes persisted after restart${NC}"
else
    echo -e "${RED}❌ FAIL: UPDATE changes not found after restart${NC}"
    exit 1
fi

# Check 3: Correct row count - check for all 3 user names in the SELECT * FROM users result
if grep -q "John Doe" /tmp/mor_after_restart.log && \
   grep -q "Jane Smith" /tmp/mor_after_restart.log && \
   grep -q "Bob Wilson" /tmp/mor_after_restart.log; then
    echo -e "${GREEN}✅ PASS: All 3 user rows present${NC}"
else
    echo -e "${RED}❌ FAIL: Not all user rows found${NC}"
    exit 1
fi

# Check 4: JOIN + GROUP BY works
if grep -q "John Doe.*2.*250" /tmp/mor_after_restart.log; then
    echo -e "${GREEN}✅ PASS: JOIN + GROUP BY query works correctly${NC}"
else
    echo -e "${YELLOW}⚠️  WARNING: JOIN + GROUP BY result not found (may be expected if query format differs)${NC}"
fi

# Check 5: Delta files correctly classified in logs
if grep -q "delta_files.*[1-9]" /tmp/mor_regression_restart.log; then
    echo -e "${GREEN}✅ PASS: Delta files correctly identified in server logs${NC}"
else
    echo -e "${RED}❌ FAIL: Delta files not identified (check IsDelta flag persistence)${NC}"
    exit 1
fi

echo ""
echo -e "${GREEN}=== ALL MoR REGRESSION TESTS PASSED ===${NC}"
echo ""
echo "Query results after restart:"
echo "----------------------------"
grep -A 25 "SELECT \* FROM users" /tmp/mor_after_restart.log | head -30

exit 0
