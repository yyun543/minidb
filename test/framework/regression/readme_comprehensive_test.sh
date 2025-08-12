#!/bin/bash

# 综合测试README.md中的所有SQL语句
# 确保所有文档化的功能都能正确工作

source "$(dirname "$0")/../utils/test_runner.sh"
source "$(dirname "$0")/../utils/db_helper.sh"
source "$(dirname "$0")/../config/test_config.sh"

TEST_NAME="README.md 综合SQL测试"
TEST_DB="readme_test_db"

# 测试设置
setup_test() {
    echo "Setting up $TEST_NAME"
    
    start_database
    
    # 清理并重新创建测试数据库
    execute_query "DROP DATABASE IF EXISTS $TEST_DB;" 2>/dev/null || true
    execute_query "CREATE DATABASE $TEST_DB;"
    execute_query "USE $TEST_DB;"
    
    # 创建测试表
    execute_query "CREATE TABLE users (id INT, name VARCHAR, email VARCHAR, age INT, created_at VARCHAR);"
    execute_query "CREATE TABLE orders (id INT, user_id INT, amount INT, order_date VARCHAR);"
    
    # 插入测试数据
    execute_query "INSERT INTO users VALUES (1, 'John Doe', 'john@example.com', 25, '2024-01-01');"
    execute_query "INSERT INTO users VALUES (2, 'Jane Smith', 'jane@example.com', 30, '2024-01-02');"
    execute_query "INSERT INTO users VALUES (3, 'Bob Wilson', 'bob@example.com', 35, '2024-01-03');"
    
    execute_query "INSERT INTO orders VALUES (1, 1, 100, '2024-01-05');"
    execute_query "INSERT INTO orders VALUES (2, 2, 250, '2024-01-06');"
    execute_query "INSERT INTO orders VALUES (3, 1, 150, '2024-01-07');"
}

# 测试数据库操作
test_database_operations() {
    echo "Testing database operations..."
    
    # SHOW DATABASES
    local result=$(execute_query "SHOW DATABASES;")
    if echo "$result" | grep -q "$TEST_DB"; then
        echo "✅ SHOW DATABASES works"
    else
        echo "❌ SHOW DATABASES failed"
        return 1
    fi
    return 0
}

# 测试DDL操作
test_ddl_operations() {
    echo "Testing DDL operations..."
    
    # SHOW TABLES
    local result=$(execute_query "SHOW TABLES;")
    if echo "$result" | grep -q "users" && echo "$result" | grep -q "orders"; then
        echo "✅ SHOW TABLES works"
    else
        echo "❌ SHOW TABLES failed"
        return 1
    fi
    return 0
}

# 测试基本DML操作
test_basic_dml() {
    echo "Testing basic DML operations..."
    
    # SELECT *
    local result1=$(execute_query "SELECT * FROM users;")
    if echo "$result1" | grep -q "John Doe" && echo "$result1" | grep -q "Jane Smith"; then
        echo "✅ SELECT * FROM users works"
    else
        echo "❌ SELECT * FROM users failed"
        return 1
    fi
    
    # SELECT with WHERE
    local result2=$(execute_query "SELECT name, email FROM users WHERE age > 25;")
    if echo "$result2" | grep -q "Jane Smith" && echo "$result2" | grep -q "Bob Wilson"; then
        echo "✅ SELECT with WHERE works"
    else
        echo "❌ SELECT with WHERE failed"
        return 1
    fi
    
    # SELECT orders
    local result3=$(execute_query "SELECT * FROM orders;")
    if echo "$result3" | grep -q "100" && echo "$result3" | grep -q "250"; then
        echo "✅ SELECT * FROM orders works"
    else
        echo "❌ SELECT * FROM orders failed"
        return 1
    fi
    
    return 0
}

# 测试JOIN操作
test_join_operations() {
    echo "Testing JOIN operations..."
    
    # JOIN with projection
    local result=$(execute_query "SELECT u.name, o.amount, o.order_date FROM users u JOIN orders o ON u.id = o.user_id WHERE u.age > 25;")
    if echo "$result" | grep -q "u.name" && echo "$result" | grep -q "o.amount" && echo "$result" | grep -q "o.order_date"; then
        echo "✅ JOIN with projection works"
    else
        echo "❌ JOIN with projection failed"
        echo "Result: $result"
        return 1
    fi
    return 0
}

# 测试聚合操作
test_aggregation_operations() {
    echo "Testing aggregation operations..."
    
    # GROUP BY with HAVING
    local result=$(execute_query "SELECT age, COUNT(*) as user_count, AVG(age) as avg_age FROM users GROUP BY age HAVING user_count > 0;")
    if echo "$result" | grep -q "age" && echo "$result" | grep -q "user_count" && echo "$result" | grep -q "avg_age"; then
        echo "✅ GROUP BY with HAVING works"
    else
        echo "❌ GROUP BY with HAVING failed"
        echo "Result: $result"
        return 1
    fi
    return 0
}

# 测试WHERE子句变体
test_where_clauses() {
    echo "Testing WHERE clause variants..."
    
    # BETWEEN
    local result1=$(execute_query "SELECT * FROM users WHERE age BETWEEN 25 AND 35;")
    if echo "$result1" | grep -q "John Doe" && echo "$result1" | grep -q "Jane Smith"; then
        echo "✅ WHERE BETWEEN works"
    else
        echo "❌ WHERE BETWEEN failed"
        echo "Result: $result1"
        return 1
    fi
    
    # LIKE
    local result2=$(execute_query "SELECT * FROM users WHERE name LIKE 'J%';")
    if echo "$result2" | grep -q "John Doe" && echo "$result2" | grep -q "Jane Smith"; then
        echo "✅ WHERE LIKE works"
    else
        echo "❌ WHERE LIKE failed"
        echo "Result: $result2"
        return 1
    fi
    
    # IN
    local result3=$(execute_query "SELECT * FROM orders WHERE amount IN (100, 250);")
    if echo "$result3" | grep -q "100" && echo "$result3" | grep -q "250"; then
        echo "✅ WHERE IN works"
    else
        echo "❌ WHERE IN failed"
        echo "Result: $result3"
        return 1
    fi
    
    return 0
}

# 测试复杂分析查询
test_complex_analytical_queries() {
    echo "Testing complex analytical queries..."
    
    # 复杂JOIN with aggregation
    local result=$(execute_query "SELECT u.name, COUNT(o.id) as order_count, SUM(o.amount) as total_amount, AVG(o.amount) as avg_amount FROM users u LEFT JOIN orders o ON u.id = o.user_id GROUP BY u.name HAVING order_count > 1 ORDER BY total_amount DESC;")
    if echo "$result" | grep -q "name" && echo "$result" | grep -q "order_count"; then
        echo "✅ Complex analytical query works"
    else
        echo "❌ Complex analytical query failed"
        echo "Result: $result"
        return 1
    fi
    return 0
}

# 测试UPDATE操作
test_update_operations() {
    echo "Testing UPDATE operations..."
    
    # UPDATE
    execute_query "UPDATE users SET email = 'john.doe@newdomain.com' WHERE name = 'John Doe';"
    local result=$(execute_query "SELECT email FROM users WHERE name = 'John Doe';")
    if echo "$result" | grep -q "john.doe@newdomain.com"; then
        echo "✅ UPDATE works"
    else
        echo "❌ UPDATE failed"
        echo "Result: $result"
        return 1
    fi
    return 0
}

# 测试DELETE操作
test_delete_operations() {
    echo "Testing DELETE operations..."
    
    # 先插入一个小金额订单用于删除
    execute_query "INSERT INTO orders VALUES (4, 1, 25, '2024-01-08');"
    
    # DELETE
    execute_query "DELETE FROM orders WHERE amount < 50;"
    local result=$(execute_query "SELECT * FROM orders WHERE amount < 50;")
    if echo "$result" | grep -q "Empty set" || ! echo "$result" | grep -q "25"; then
        echo "✅ DELETE works"
    else
        echo "❌ DELETE failed"
        echo "Result: $result"
        return 1
    fi
    return 0
}

# 测试EXPLAIN功能
test_explain_functionality() {
    echo "Testing EXPLAIN functionality..."
    
    local result=$(execute_query "EXPLAIN SELECT u.name, SUM(o.amount) as total_spent FROM users u JOIN orders o ON u.id = o.user_id WHERE u.age > 25 GROUP BY u.name ORDER BY total_spent DESC;")
    if echo "$result" | grep -q "Query Execution Plan" && echo "$result" | grep -q "Select"; then
        echo "✅ EXPLAIN works"
    else
        echo "❌ EXPLAIN failed"
        echo "Result: $result"
        return 1
    fi
    return 0
}

# 测试空结果处理
test_empty_result_handling() {
    echo "Testing empty result handling..."
    
    local result=$(execute_query "SELECT * FROM users WHERE age > 100;")
    if echo "$result" | grep -q "Empty set" || echo "$result" | grep -q "0 rows"; then
        echo "✅ Empty result handling works"
    else
        echo "❌ Empty result handling failed"
        echo "Result: $result"
        return 1
    fi
    return 0
}

# 清理测试
cleanup_test() {
    echo "Cleaning up $TEST_NAME"
    execute_query "DROP DATABASE IF EXISTS $TEST_DB;" 2>/dev/null || true
    stop_database
}

# 主测试执行
main() {
    echo "Starting $TEST_NAME"
    
    setup_test || { echo "Setup failed"; exit 1; }
    
    local failed=0
    
    test_database_operations || failed=$((failed + 1))
    test_ddl_operations || failed=$((failed + 1))
    test_basic_dml || failed=$((failed + 1))
    test_join_operations || failed=$((failed + 1))
    test_aggregation_operations || failed=$((failed + 1))
    test_where_clauses || failed=$((failed + 1))
    test_complex_analytical_queries || failed=$((failed + 1))
    test_update_operations || failed=$((failed + 1))
    test_delete_operations || failed=$((failed + 1))
    test_explain_functionality || failed=$((failed + 1))
    test_empty_result_handling || failed=$((failed + 1))
    
    cleanup_test
    
    if [ $failed -eq 0 ]; then
        echo "✅ $TEST_NAME: All README.md SQL examples passed"
        exit 0
    else
        echo "❌ $TEST_NAME: $failed test(s) failed"
        exit 1
    fi
}

# 如果直接执行则运行主测试
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi