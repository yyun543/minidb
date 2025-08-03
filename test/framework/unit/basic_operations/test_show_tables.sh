#!/bin/bash

# Test SHOW TABLES functionality
echo "=== Testing SHOW TABLES Fix ==="

# Clean up any existing WAL files
rm -f minidb.wal

# Start server in background
./minidb -port 7998 &
SERVER_PID=$!

# Wait for server to start
sleep 2

# Function to execute SQL
exec_sql() {
    echo -e "$1" | nc localhost 7998 | grep -v "Welcome\|Session\|Type\|minidb>" | grep -v "^$" | head -20
}

echo "=== Test 1: Create database and tables ==="
exec_sql "CREATE DATABASE testdb;"
exec_sql "USE testdb;"
exec_sql "CREATE TABLE users (id INT, name VARCHAR, email VARCHAR, age INT, created_at VARCHAR);"
exec_sql "CREATE TABLE orders (id INT, user_id INT, amount INT, order_date VARCHAR);"

echo "=== Test 2: SHOW TABLES (should show users and orders) ==="
exec_sql "SHOW TABLES;"

# Clean up
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null
rm -f minidb.wal

echo "=== Test complete ==="