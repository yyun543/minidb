#!/bin/bash

# Test script to verify delta_log system table fix
# This script tests that the column headers and data are now aligned correctly

set -e

echo "=== Testing delta_log System Table Fix ==="
echo ""

# Start the server in the background
echo "Starting MiniDB server..."
cd /Users/10270273/codes/minidb
./minidb &
SERVER_PID=$!
sleep 2

echo "Server started with PID: $SERVER_PID"
echo ""

# Function to cleanup on exit
cleanup() {
    echo ""
    echo "Stopping server..."
    kill $SERVER_PID 2>/dev/null || true
    wait $SERVER_PID 2>/dev/null || true
    echo "Server stopped."
}
trap cleanup EXIT

# Connect to server and run test queries
echo "=== Test 1: Create test database and table ==="
echo "CREATE DATABASE ecommerce;" | nc localhost 7205
echo "USE ecommerce;" | nc localhost 7205
echo "CREATE TABLE users (id INT, name VARCHAR);" | nc localhost 7205
echo "INSERT INTO users (id, name) VALUES (1, 'John');" | nc localhost 7205
echo ""

echo "=== Test 2: Query delta_log table (should show correct column order) ==="
echo "SELECT * FROM sys.delta_log;" | nc localhost 7205
echo ""

echo "=== Test 3: Query delta_log with WHERE clause ==="
echo "SELECT version, operation, table_schema, table_name FROM sys.delta_log WHERE table_schema = 'ecommerce';" | nc localhost 7205
echo ""

echo "=== Test 4: Query sys.columns for ecommerce.users ==="
echo "SELECT column_name, data_type FROM sys.columns WHERE table_schema = 'ecommerce' AND table_name = 'users';" | nc localhost 7205
echo ""

echo "=== All tests completed ==="
