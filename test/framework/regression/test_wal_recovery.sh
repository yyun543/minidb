#!/bin/bash

echo "=== Testing WAL Recovery with Schema ==="

# Clean up
rm -f minidb.wal

echo "=== Phase 1: Create data and stop server ==="
./minidb -port 7995 &
SERVER_PID=$!
sleep 2

{
    echo "CREATE DATABASE ecommerce;"
    sleep 1
    echo "USE ecommerce;"  
    sleep 1
    echo "CREATE TABLE users (id INT, name VARCHAR, email VARCHAR, age INT, created_at VARCHAR);"
    sleep 1
    echo "CREATE TABLE orders (id INT, user_id INT, amount INT, order_date VARCHAR);"
    sleep 1
    echo "INSERT INTO users VALUES (1, 'John Doe', 'john@example.com', 25, '2024-01-01');"
    sleep 1
    echo "INSERT INTO users VALUES (2, 'Jane Smith', 'jane@example.com', 30, '2024-01-02');"
    sleep 1
    echo "INSERT INTO orders VALUES (1, 1, 100, '2024-01-05');"
    sleep 1
    echo "INSERT INTO orders VALUES (2, 2, 250, '2024-01-06');"
    sleep 1
    echo "INSERT INTO orders VALUES (3, 1, 150, '2024-01-07');"
    sleep 1
    echo "SELECT * FROM orders;"
    sleep 1
    echo "quit;"
} | telnet localhost 7995 > /dev/null 2>&1

kill $SERVER_PID
wait $SERVER_PID 2>/dev/null

echo "=== Phase 2: Restart server and test schema recovery ==="
./minidb -port 7995 &
SERVER_PID=$!
sleep 3

{
    echo "USE ecommerce;"
    sleep 1
    echo "SELECT * FROM orders;"
    sleep 1  
    echo "SELECT * FROM orders WHERE amount < 200;"  # This should work after fix
    sleep 1
    echo "SELECT * FROM users WHERE age > 25;"  # This should also work
    sleep 1
    echo "quit;"
} | telnet localhost 7995

kill $SERVER_PID
wait $SERVER_PID 2>/dev/null

rm -f minidb.wal
echo "=== Test Complete ==="