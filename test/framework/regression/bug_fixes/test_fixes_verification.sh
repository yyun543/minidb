#!/bin/bash

# Comprehensive test to verify the fixes work correctly
echo "=== Testing Bug Fixes Verification ==="

# Start server in background
../../../../minidb &
MINIDB_PID=$!
echo "Started minidb with PID: $MINIDB_PID"

sleep 3

echo ""
echo "=== Testing JOIN Projection Fix ==="
echo "Running: SELECT u.name, o.amount, o.order_date FROM users u JOIN orders o ON u.id = o.user_id WHERE u.age > 25;"

join_result=$(cat << 'EOF' | nc localhost 7205
DROP DATABASE IF EXISTS testdb;
CREATE DATABASE testdb;
USE testdb;
CREATE TABLE users (id INT, name VARCHAR, email VARCHAR, age INT, created_at VARCHAR);
CREATE TABLE orders (id INT, user_id INT, amount INT, order_date VARCHAR);
INSERT INTO users VALUES (1, 'John Doe', 'john@example.com', 25, '2024-01-01');
INSERT INTO users VALUES (2, 'Jane Smith', 'jane@example.com', 30, '2024-01-02');
INSERT INTO users VALUES (3, 'Bob Wilson', 'bob@example.com', 35, '2024-01-03');
INSERT INTO orders VALUES (1, 1, 100, '2024-01-05');
INSERT INTO orders VALUES (2, 2, 250, '2024-01-06');
INSERT INTO orders VALUES (3, 3, 150, '2024-01-07');
SELECT u.name, o.amount, o.order_date FROM users u JOIN orders o ON u.id = o.user_id WHERE u.age > 25;
EOF
)

echo "JOIN Query Result:"
echo "$join_result" | tail -10

# Check JOIN projection
if echo "$join_result" | grep -q "u.name.*o.amount.*o.order_date"; then
    echo "✅ JOIN projection headers are correct"
else
    echo "❌ JOIN projection headers are incorrect"
fi

if echo "$join_result" | grep -q "email\|created_at\|user_id"; then
    echo "❌ JOIN projection contains unwanted columns"
else
    echo "✅ JOIN projection contains only projected columns"
fi

echo ""
echo "=== Testing LIKE Header Fix ==="
echo "Running: SELECT * FROM users WHERE name LIKE 'J%';"

like_result=$(cat << 'EOF' | nc localhost 7205
USE testdb;
SELECT * FROM users WHERE name LIKE 'J%';
EOF
)

echo "LIKE Query Result:"
echo "$like_result" | tail -10

# Check LIKE headers
if echo "$like_result" | grep -q "| \* *|"; then
    echo "❌ LIKE query shows '*' as header"
else
    echo "✅ LIKE query shows proper column headers"
fi

if echo "$like_result" | grep -q "id.*name.*email.*age.*created_at"; then
    echo "✅ LIKE query shows all expected headers"
else
    echo "❌ LIKE query is missing expected headers"
fi

echo ""
echo "=== Testing LIKE with Specific Columns ==="
echo "Running: SELECT name, email FROM users WHERE name LIKE 'J%';"

like_specific_result=$(cat << 'EOF' | nc localhost 7205
USE testdb;
SELECT name, email FROM users WHERE name LIKE 'J%';
EOF
)

echo "LIKE Specific Columns Result:"
echo "$like_specific_result" | tail -10

# Check specific columns
if echo "$like_specific_result" | grep -q "name.*email"; then
    echo "✅ LIKE with specific columns shows correct headers"
else
    echo "❌ LIKE with specific columns shows incorrect headers"
fi

if echo "$like_specific_result" | grep -q "id\|age\|created_at"; then
    echo "❌ LIKE with specific columns contains unwanted columns"
else
    echo "✅ LIKE with specific columns contains only projected columns"
fi

echo ""
echo "=== Testing Simple JOIN without WHERE ==="
echo "Running: SELECT u.name, o.amount FROM users u JOIN orders o ON u.id = o.user_id;"

simple_join_result=$(cat << 'EOF' | nc localhost 7205
USE testdb;
SELECT u.name, o.amount FROM users u JOIN orders o ON u.id = o.user_id;
EOF
)

echo "Simple JOIN Result:"
echo "$simple_join_result" | tail -10

# Check simple JOIN
if echo "$simple_join_result" | grep -q "u.name.*o.amount"; then
    echo "✅ Simple JOIN shows correct headers"
else
    echo "❌ Simple JOIN shows incorrect headers"
fi

# Cleanup
echo ""
echo "=== Cleanup ==="
kill $MINIDB_PID 2>/dev/null
wait $MINIDB_PID 2>/dev/null
echo "Cleanup complete."

echo ""
echo "=== Summary ==="
echo "Both bugs should now be fixed:"
echo "1. JOIN queries should only return projected columns"
echo "2. LIKE queries should show proper column headers instead of '*'"