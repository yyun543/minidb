#!/bin/bash

# WHERE子句单元测试
# 测试各种WHERE条件的正确性

source "$(dirname "${BASH_SOURCE[0]}")/../../utils/test_runner.sh"

# 设置测试数据
setup_where_test_data() {
    echo "Setting up WHERE clause test data..."
    
    assert_query_succeeds "CREATE DATABASE where_test;" "Should create test database"
    assert_query_succeeds "USE where_test;" "Should switch to database"
    assert_query_succeeds "CREATE TABLE users (id INT, name VARCHAR, age INT, active VARCHAR);" "Should create users table"
    
    # 插入测试数据
    assert_query_succeeds "INSERT INTO users VALUES (1, 'Alice', 25, 'yes');" "Insert Alice"
    assert_query_succeeds "INSERT INTO users VALUES (2, 'Bob', 30, 'no');" "Insert Bob"  
    assert_query_succeeds "INSERT INTO users VALUES (3, 'Charlie', 35, 'yes');" "Insert Charlie"
    assert_query_succeeds "INSERT INTO users VALUES (4, 'Diana', 28, 'yes');" "Insert Diana"
    assert_query_succeeds "INSERT INTO users VALUES (5, 'Eve', 22, 'no');" "Insert Eve"
}

# 测试基础比较操作符
test_basic_comparison_operators() {
    echo "Testing basic comparison operators..."
    
    # 等于
    assert_query_row_count "SELECT * FROM users WHERE age = 25;" 1 "Should find 1 user with age 25"
    assert_query_contains_value "SELECT name FROM users WHERE age = 25;" "Alice" "Should find Alice with age 25"
    
    # 不等于  
    assert_query_row_count "SELECT * FROM users WHERE age != 25;" 4 "Should find 4 users with age != 25"
    
    # 大于
    assert_query_row_count "SELECT * FROM users WHERE age > 25;" 3 "Should find 3 users with age > 25"
    
    # 小于
    assert_query_row_count "SELECT * FROM users WHERE age < 30;" 3 "Should find 3 users with age < 30"
    
    # 大于等于
    assert_query_row_count "SELECT * FROM users WHERE age >= 30;" 2 "Should find 2 users with age >= 30"
    
    # 小于等于
    assert_query_row_count "SELECT * FROM users WHERE age <= 28;" 3 "Should find 3 users with age <= 28"
}

# 测试字符串比较
test_string_comparison() {
    echo "Testing string comparison..."
    
    # 字符串相等
    assert_query_row_count "SELECT * FROM users WHERE name = 'Alice';" 1 "Should find Alice by name"
    assert_query_row_count "SELECT * FROM users WHERE active = 'yes';" 3 "Should find 3 active users"
    
    # 字符串不等
    assert_query_row_count "SELECT * FROM users WHERE active != 'yes';" 2 "Should find 2 inactive users"
}

# 测试AND逻辑操作符
test_and_operator() {
    echo "Testing AND operator..."
    
    # 简单AND条件
    assert_query_row_count "SELECT * FROM users WHERE age > 25 AND active = 'yes';" 2 "Should find 2 users: age>25 AND active"
    assert_query_contains_value "SELECT name FROM users WHERE age > 25 AND active = 'yes';" "Charlie" "Should find Charlie"
    assert_query_contains_value "SELECT name FROM users WHERE age > 25 AND active = 'yes';" "Diana" "Should find Diana"
    
    # 多个AND条件
    assert_query_row_count "SELECT * FROM users WHERE age > 20 AND age < 30 AND active = 'yes';" 2 "Should find 2 users with complex AND"
    
    # AND条件无结果
    assert_query_row_count "SELECT * FROM users WHERE age > 40 AND active = 'yes';" 0 "Should find no users with age>40"
}

# 测试OR逻辑操作符
test_or_operator() {
    echo "Testing OR operator..."
    
    # 简单OR条件
    assert_query_row_count "SELECT * FROM users WHERE age = 25 OR age = 35;" 2 "Should find 2 users: age=25 OR age=35"
    assert_query_contains_value "SELECT name FROM users WHERE age = 25 OR age = 35;" "Alice" "Should find Alice"
    assert_query_contains_value "SELECT name FROM users WHERE age = 25 OR age = 35;" "Charlie" "Should find Charlie"
    
    # OR条件覆盖所有
    assert_query_row_count "SELECT * FROM users WHERE active = 'yes' OR active = 'no';" 5 "Should find all 5 users"
}

# 测试LIKE操作符
test_like_operator() {
    echo "Testing LIKE operator..."
    
    # 前缀匹配
    assert_query_row_count "SELECT * FROM users WHERE name LIKE 'A%';" 1 "Should find 1 user with name starting with A"
    assert_query_contains_value "SELECT name FROM users WHERE name LIKE 'A%';" "Alice" "Should find Alice with A% pattern"
    
    # 多字符前缀匹配
    assert_query_row_count "SELECT * FROM users WHERE name LIKE 'Ch%';" 1 "Should find Charlie with Ch% pattern"
    assert_query_contains_value "SELECT name FROM users WHERE name LIKE 'Ch%';" "Charlie" "Should find Charlie"
}

# 测试IN操作符  
test_in_operator() {
    echo "Testing IN operator..."
    
    # 数字IN操作
    assert_query_row_count "SELECT * FROM users WHERE age IN (25, 30, 35);" 3 "Should find 3 users with ages in (25,30,35)"
    assert_query_contains_value "SELECT name FROM users WHERE age IN (25, 30, 35);" "Alice" "Should find Alice"
    assert_query_contains_value "SELECT name FROM users WHERE age IN (25, 30, 35);" "Bob" "Should find Bob" 
    assert_query_contains_value "SELECT name FROM users WHERE age IN (25, 30, 35);" "Charlie" "Should find Charlie"
    
    # 单值IN操作
    assert_query_row_count "SELECT * FROM users WHERE age IN (25);" 1 "Should find 1 user with age in (25)"
}

# 测试复合WHERE条件
test_compound_where_conditions() {
    echo "Testing compound WHERE conditions..."
    
    # AND + OR组合（需要测试优先级）
    assert_query_row_count "SELECT * FROM users WHERE age > 25 AND (active = 'yes' OR name = 'Bob');" 3 "Complex condition with parentheses"
    
    # 多个字段的复合条件
    assert_query_row_count "SELECT * FROM users WHERE (age = 25 OR age = 30) AND active = 'yes';" 1 "Should find Alice only"
    assert_query_contains_value "SELECT name FROM users WHERE (age = 25 OR age = 30) AND active = 'yes';" "Alice" "Should find Alice"
}

# 测试边界情况
test_edge_cases() {
    echo "Testing edge cases..."
    
    # 空结果集
    assert_query_row_count "SELECT * FROM users WHERE age = 100;" 0 "Should return empty set for age=100"
    
    # 所有记录
    assert_query_row_count "SELECT * FROM users WHERE age > 0;" 5 "Should return all records for age>0"
    
    # 字符串大小写敏感性测试
    assert_query_row_count "SELECT * FROM users WHERE name = 'alice';" 0 "Should be case-sensitive (alice vs Alice)"
    assert_query_row_count "SELECT * FROM users WHERE name = 'Alice';" 1 "Should find Alice with correct case"
}

# 清理测试数据
cleanup_where_test_data() {
    execute_query "DROP DATABASE where_test;" >/dev/null 2>&1
}

# 主测试函数
main() {
    # 启动数据库
    start_database || {
        skip_test_suite "WHERE Clause Tests" "Failed to start database"
        return 1
    }
    
    # 运行所有测试
    setup_where_test_data
    test_basic_comparison_operators
    test_string_comparison
    test_and_operator
    test_or_operator
    test_like_operator
    test_in_operator
    test_compound_where_conditions
    test_edge_cases
    
    # 清理
    cleanup_where_test_data
    
    return 0
}

# 如果直接执行此脚本
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi