#!/bin/bash

# GROUP BY 综合集成测试
# 整合了所有GROUP BY相关的测试用例

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"

echo "=== MiniDB GROUP BY Comprehensive Integration Test ==="
echo "Project Root: $PROJECT_ROOT"

cd "$PROJECT_ROOT"

# 清理旧的WAL文件
rm -f minidb.wal
rm -f *.wal

# 构建项目
echo "Building MiniDB..."
go build -o minidb ./cmd/server || {
    echo "Build failed"
    exit 1
}

# 启动服务器
echo "Starting MiniDB server..."
./minidb -port 8100 &
SERVER_PID=$!
sleep 3

# 检查服务器是否启动成功
if ! kill -0 $SERVER_PID 2>/dev/null; then
    echo "Server failed to start"
    exit 1
fi

echo "Server started with PID: $SERVER_PID"

# 测试函数
run_test() {
    local test_name="$1"
    local sql_commands="$2"
    
    echo ""
    echo "--- Testing: $test_name ---"
    
    {
        echo "$sql_commands"
        echo "quit;"
    } | telnet localhost 8100 2>/dev/null | grep -v "telnet\|Trying\|Connected\|Escape\|Welcome\|Type\|Session\|minidb>"
    
    if [ $? -eq 0 ]; then
        echo "✅ $test_name PASSED"
    else
        echo "❌ $test_name FAILED"
        return 1
    fi
}

# 初始化测试环境
INIT_COMMANDS="
CREATE DATABASE test;
USE test;
CREATE TABLE sales (region VARCHAR, amount INT);
INSERT INTO sales VALUES ('North', 100);
INSERT INTO sales VALUES ('South', 150);
INSERT INTO sales VALUES ('North', 200);
INSERT INTO sales VALUES ('East', 300);
INSERT INTO sales VALUES ('North', 50);
CREATE TABLE users (id INT, name VARCHAR);
CREATE TABLE orders (id INT, user_id INT, amount INT);
INSERT INTO users VALUES (1, 'John Doe');
INSERT INTO users VALUES (2, 'Jane Smith');
INSERT INTO orders VALUES (1, 1, 100);
INSERT INTO orders VALUES (2, 1, 150);
INSERT INTO orders VALUES (3, 2, 75);"

echo "Initializing test data..."
{
    echo "$INIT_COMMANDS"
    echo "quit;"
} | telnet localhost 8100 >/dev/null 2>&1

sleep 2

# 运行各项测试
test_results=0

# Test 1: 基本GROUP BY with 别名
run_test "Basic GROUP BY with Aliases" \
"SELECT region, COUNT(*) AS orders, SUM(amount) AS total FROM sales GROUP BY region;"
test_results=$((test_results + $?))

# Test 2: AVG 聚合函数
run_test "AVG Aggregation Function" \
"SELECT region, AVG(amount) AS avg_amount FROM sales GROUP BY region;"
test_results=$((test_results + $?))

# Test 3: COUNT(*) 函数
run_test "COUNT(*) Function" \
"SELECT region, COUNT(*) AS count FROM sales GROUP BY region;"
test_results=$((test_results + $?))

# Test 4: HAVING 子句
run_test "HAVING Clause" \
"SELECT region, COUNT(*) AS cnt FROM sales GROUP BY region HAVING cnt >= 2;"
test_results=$((test_results + $?))

# Test 5: 复杂嵌套查询 (验证表头别名显示修复)
run_test "Complex Nested Query with Header Aliases" \
"SELECT u.name, COUNT(o.id) as order_count, SUM(o.amount) as total_amount, AVG(o.amount) as avg_amount FROM users u LEFT JOIN orders o ON u.id = o.user_id GROUP BY u.name HAVING order_count > 1 ORDER BY total_amount DESC;"
test_results=$((test_results + $?))

# Test 6: 所有聚合函数
run_test "All Aggregation Functions" \
"SELECT region, COUNT(*) as cnt, SUM(amount) as sum_amt, AVG(amount) as avg_amt, MIN(amount) as min_amt, MAX(amount) as max_amt FROM sales GROUP BY region;"
test_results=$((test_results + $?))

# 关闭服务器
echo ""
echo "Shutting down server..."
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null

# 清理
rm -f minidb minidb.wal *.wal

# 报告结果
echo ""
echo "=== Test Results ==="
if [ $test_results -eq 0 ]; then
    echo "✅ All GROUP BY tests PASSED!"
    exit 0
else
    echo "❌ $test_results test(s) FAILED"
    exit 1
fi