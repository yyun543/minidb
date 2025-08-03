#!/bin/bash

echo "=== MiniDB GROUP BY Working Features Demo ==="

rm -f minidb.wal
go build -o minidb ./cmd/server

./minidb -port 8100 &
SERVER_PID=$!
sleep 2

{
    echo "CREATE DATABASE demo;"
    sleep 1
    echo "USE demo;"
    sleep 1
    echo "CREATE TABLE sales (region VARCHAR, amount INT);"
    sleep 1
    echo "INSERT INTO sales VALUES ('North', 100);"
    sleep 1
    echo "INSERT INTO sales VALUES ('South', 150);"
    sleep 1
    echo "INSERT INTO sales VALUES ('North', 200);"
    sleep 1
    echo "INSERT INTO sales VALUES ('East', 300);"
    sleep 1
    echo "INSERT INTO sales VALUES ('North', 50);"
    sleep 1
    echo ""
    echo "-- Test 1: Basic GROUP BY with aliases"
    echo "SELECT region, COUNT(*) AS orders, SUM(amount) AS total FROM sales GROUP BY region;"
    sleep 3
    echo ""
    echo "-- Test 2: AVG function"
    echo "SELECT region, AVG(amount) AS avg_amount FROM sales GROUP BY region;"
    sleep 3
    echo ""
    echo "-- Test 3: HAVING clause"
    echo "SELECT region, COUNT(*) AS cnt FROM sales GROUP BY region HAVING cnt >= 2;"
    sleep 3
    echo ""
    echo "quit;"
} | telnet localhost 8100

kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null

echo ""
echo "=== Demo Complete ==="