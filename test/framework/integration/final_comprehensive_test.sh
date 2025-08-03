#\!/bin/bash

echo "=== Final Comprehensive GROUP BY Test ==="

rm -f minidb.wal
go build -o minidb ./cmd/server

./minidb -port 8100 &
SERVER_PID=$\!
sleep 2

{
    echo "CREATE DATABASE test;"
    sleep 1
    echo "USE test;"
    sleep 1
    echo "CREATE TABLE users (id INT, name VARCHAR);"
    sleep 1
    echo "CREATE TABLE orders (id INT, user_id INT, amount INT);"
    sleep 1
    echo "INSERT INTO users VALUES (1, 'John Doe');"
    sleep 1
    echo "INSERT INTO users VALUES (2, 'Jane Smith');"
    sleep 1
    echo "INSERT INTO orders VALUES (1, 1, 100);"
    sleep 1
    echo "INSERT INTO orders VALUES (2, 1, 150);"
    sleep 1
    echo "INSERT INTO orders VALUES (3, 2, 75);"
    sleep 1
    echo ""
    echo "-- Complex GROUP BY with JOIN, HAVING, ORDER BY"
    echo "SELECT u.name, COUNT(o.id) as order_count, SUM(o.amount) as total_amount, AVG(o.amount) as avg_amount FROM users u LEFT JOIN orders o ON u.id = o.user_id GROUP BY u.name HAVING order_count > 1 ORDER BY total_amount DESC;"
    sleep 5
    echo ""
    echo "quit;"
} | telnet localhost 8100

kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null
echo "=== Test Complete ==="
EOF < /dev/null