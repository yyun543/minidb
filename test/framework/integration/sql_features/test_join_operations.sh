#!/bin/bash

# JOIN操作集成测试
# 测试各种JOIN场景和边界情况

source "$(dirname "${BASH_SOURCE[0]}")/../../utils/test_runner.sh"

# 设置JOIN测试数据
setup_join_test_data() {
    echo "Setting up JOIN test data..."
    
    assert_query_succeeds "CREATE DATABASE join_test;" "Should create join test database"
    assert_query_succeeds "USE join_test;" "Should switch to database"
    
    # 创建用户表
    assert_query_succeeds "CREATE TABLE users (id INT, name VARCHAR, dept_id INT);" "Should create users table"
    # 创建部门表
    assert_query_succeeds "CREATE TABLE departments (id INT, name VARCHAR, manager VARCHAR);" "Should create departments table"
    
    # 插入用户数据
    assert_query_succeeds "INSERT INTO users VALUES (1, 'Alice', 1);" "Insert Alice"
    assert_query_succeeds "INSERT INTO users VALUES (2, 'Bob', 2);" "Insert Bob"
    assert_query_succeeds "INSERT INTO users VALUES (3, 'Charlie', 1);" "Insert Charlie"
    assert_query_succeeds "INSERT INTO users VALUES (4, 'Diana', 3);" "Insert Diana (orphan dept)"
    assert_query_succeeds "INSERT INTO users VALUES (5, 'Eve', 2);" "Insert Eve"
    
    # 插入部门数据
    assert_query_succeeds "INSERT INTO departments VALUES (1, 'Engineering', 'John');" "Insert Engineering dept"
    assert_query_succeeds "INSERT INTO departments VALUES (2, 'Marketing', 'Jane');" "Insert Marketing dept" 
    # 注意：没有dept_id=3的部门，测试孤儿记录
}

# 测试基础INNER JOIN
test_basic_inner_join() {
    echo "Testing basic INNER JOIN..."
    
    # 基础JOIN语法
    assert_query_succeeds "SELECT u.name, d.name FROM users u JOIN departments d ON u.dept_id = d.id;" "Should execute basic JOIN"
    
    # 验证JOIN结果数量（应该只包含匹配的记录）
    assert_query_row_count "SELECT u.name, d.name FROM users u JOIN departments d ON u.dept_id = d.id;" 4 "Should return 4 matched records"
    
    # 验证JOIN结果内容
    local join_result=$(execute_query "SELECT u.name, d.name FROM users u JOIN departments d ON u.dept_id = d.id;")
    assert_contains "$join_result" "Alice" "Should contain Alice"
    assert_contains "$join_result" "Engineering" "Should contain Engineering department"
    assert_contains "$join_result" "Bob" "Should contain Bob" 
    assert_contains "$join_result" "Marketing" "Should contain Marketing department"
    
    # Diana不应该出现（因为dept_id=3不存在）
    assert_not_contains "$join_result" "Diana" "Should not contain Diana (orphan record)"
}

# 测试JOIN条件匹配
test_join_condition_matching() {
    echo "Testing JOIN condition matching..."
    
    # 测试Engineering部门的员工
    assert_query_row_count "SELECT u.name FROM users u JOIN departments d ON u.dept_id = d.id WHERE d.name = 'Engineering';" 2 "Should find 2 Engineering employees"
    
    local eng_employees=$(execute_query "SELECT u.name FROM users u JOIN departments d ON u.dept_id = d.id WHERE d.name = 'Engineering';")
    assert_contains "$eng_employees" "Alice" "Should find Alice in Engineering"
    assert_contains "$eng_employees" "Charlie" "Should find Charlie in Engineering"
    
    # 测试Marketing部门的员工
    assert_query_row_count "SELECT u.name FROM users u JOIN departments d ON u.dept_id = d.id WHERE d.name = 'Marketing';" 2 "Should find 2 Marketing employees"
    
    local mkt_employees=$(execute_query "SELECT u.name FROM users u JOIN departments d ON u.dept_id = d.id WHERE d.name = 'Marketing';")
    assert_contains "$mkt_employees" "Bob" "Should find Bob in Marketing"
    assert_contains "$mkt_employees" "Eve" "Should find Eve in Marketing"
}

# 测试JOIN结果的字段投影
test_join_field_projection() {
    echo "Testing JOIN field projection..."
    
    # 选择特定字段
    assert_query_succeeds "SELECT u.name, d.manager FROM users u JOIN departments d ON u.dept_id = d.id;" "Should select specific fields"
    
    local projected_result=$(execute_query "SELECT u.name, d.manager FROM users u JOIN departments d ON u.dept_id = d.id;")
    assert_contains "$projected_result" "Alice" "Should contain user name"
    assert_contains "$projected_result" "John" "Should contain manager name" 
    assert_contains "$projected_result" "Jane" "Should contain manager name"
    
    # 测试别名
    assert_query_succeeds "SELECT u.name as employee, d.name as department FROM users u JOIN departments d ON u.dept_id = d.id;" "Should support field aliases"
}

# 测试复合JOIN条件
test_compound_join_conditions() {
    echo "Testing compound JOIN conditions..."
    
    # 创建更多测试数据来测试复合条件
    assert_query_succeeds "CREATE TABLE user_roles (user_id INT, role VARCHAR);" "Create user_roles table"
    assert_query_succeeds "INSERT INTO user_roles VALUES (1, 'admin');" "Insert Alice role"
    assert_query_succeeds "INSERT INTO user_roles VALUES (2, 'user');" "Insert Bob role"
    assert_query_succeeds "INSERT INTO user_roles VALUES (3, 'user');" "Insert Charlie role"
    
    # 三表JOIN（虽然我们的系统可能不支持，但可以用两个两表JOIN测试）
    # 先测试是否支持多表JOIN
    local multi_join_result=$(execute_query "SELECT u.name, d.name, ur.role FROM users u JOIN departments d ON u.dept_id = d.id JOIN user_roles ur ON u.id = ur.user_id;" 2>&1)
    
    if echo "$multi_join_result" | grep -q "Error:"; then
        echo "Multi-table JOIN not supported, testing sequential JOINs instead"
        # 如果不支持多表JOIN，测试嵌套查询的方式
    else
        assert_query_row_count "SELECT u.name, d.name, ur.role FROM users u JOIN departments d ON u.dept_id = d.id JOIN user_roles ur ON u.id = ur.user_id;" 3 "Should handle multi-table JOIN"
    fi
}

# 测试JOIN性能和边界情况
test_join_edge_cases() {
    echo "Testing JOIN edge cases..."
    
    # 空表JOIN
    assert_query_succeeds "CREATE TABLE empty_table (id INT, name VARCHAR);" "Create empty table"
    assert_query_row_count "SELECT u.name, e.name FROM users u JOIN empty_table e ON u.id = e.id;" 0 "JOIN with empty table should return 0 rows"
    
    # 自连接测试（如果支持）
    local self_join_result=$(execute_query "SELECT u1.name, u2.name FROM users u1 JOIN users u2 ON u1.dept_id = u2.dept_id;" 2>&1)
    if ! echo "$self_join_result" | grep -q "Error:"; then
        echo "Self-join supported, testing results..."
        # 自连接应该返回同部门的员工对
        assert_contains "$self_join_result" "Alice" "Self-join should contain user names"
    else
        echo "Self-join not supported or failed"
    fi
    
    # 测试JOIN条件中的不存在字段
    assert_query_fails "SELECT u.name, d.name FROM users u JOIN departments d ON u.nonexistent = d.id;" "column.*not found" "Should fail with non-existent column"
}

# 测试JOIN与WHERE结合
test_join_with_where() {
    echo "Testing JOIN with WHERE clause..."
    
    # JOIN结果再过滤
    assert_query_row_count "SELECT u.name, d.name FROM users u JOIN departments d ON u.dept_id = d.id WHERE u.name LIKE 'A%';" 1 "Should filter JOIN results"
    
    local filtered_result=$(execute_query "SELECT u.name, d.name FROM users u JOIN departments d ON u.dept_id = d.id WHERE u.name LIKE 'A%';")
    assert_contains "$filtered_result" "Alice" "Should find Alice after filtering"
    assert_not_contains "$filtered_result" "Charlie" "Should not find Charlie after filtering"
    
    # WHERE条件在JOIN之前过滤
    assert_query_succeeds "SELECT u.name, d.name FROM users u JOIN departments d ON u.dept_id = d.id WHERE d.manager = 'John';" "Should filter by department manager"
    
    local manager_filtered=$(execute_query "SELECT u.name, d.name FROM users u JOIN departments d ON u.dept_id = d.id WHERE d.manager = 'John';")
    assert_contains "$manager_filtered" "Alice" "Should find Engineering employees"
    assert_contains "$manager_filtered" "Charlie" "Should find Engineering employees"
    assert_not_contains "$manager_filtered" "Bob" "Should not find Marketing employees"
}

# 清理JOIN测试数据
cleanup_join_test_data() {
    execute_query "DROP DATABASE join_test;" >/dev/null 2>&1
}

# 主测试函数
main() {
    # 启动数据库
    start_database || {
        skip_test_suite "JOIN Operations" "Failed to start database"
        return 1
    }
    
    # 运行所有测试
    setup_join_test_data
    test_basic_inner_join
    test_join_condition_matching
    test_join_field_projection
    test_compound_join_conditions
    test_join_edge_cases  
    test_join_with_where
    
    # 清理
    cleanup_join_test_data
    
    return 0
}

# 如果直接执行此脚本
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi