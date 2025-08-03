#!/bin/bash

echo "=== 测试BETWEEN操作符 ==="

rm -f minidb.wal

go build -o minidb ./cmd/server || exit 1

echo "启动服务器..."
./minidb -port 8100 &
SERVER_PID=$!
sleep 2

echo "执行BETWEEN查询测试..."
{
    echo "CREATE DATABASE test;"
    sleep 1
    echo "USE test;"
    sleep 1
    echo "CREATE TABLE users (age INT);"
    sleep 1
    echo "INSERT INTO users VALUES (25);"
    sleep 1
    echo "INSERT INTO users VALUES (30);"
    sleep 1
    echo "INSERT INTO users VALUES (35);"
    sleep 1
    echo "SELECT * FROM users WHERE age BETWEEN 25 AND 35;"
    sleep 3
    echo "quit;"
} | telnet localhost 8100

kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null
echo "=== BETWEEN测试完成 ==="