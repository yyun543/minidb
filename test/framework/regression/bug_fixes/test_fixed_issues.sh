#!/bin/bash

# 回归测试 - 已修复问题验证
# 确保之前修复的问题不会重现

source "$(dirname "${BASH_SOURCE[0]}")/../../utils/test_runner.sh"

# 设置回归测试数据
setup_regression_test_data() {
    echo "Setting up regression test data..."
    
    assert_query_succeeds "CREATE DATABASE regression_test;" "Should create regression test database"
    assert_query_succeeds "USE regression_test;" "Should switch to database"
}

# 测试JOIN问题修复 - 之前返回Empty set
test_join_empty_set_fix() {
    echo "Testing JOIN empty set issue fix..."
    
    # 创建测试表
    assert_query_succeeds "CREATE TABLE test_users (id INT, name VARCHAR, dept_id INT);" "Create test_users table"
    assert_query_succeeds "CREATE TABLE test_depts (id INT, name VARCHAR);" "Create test_depts table"
    
    # 插入测试数据
    assert_query_succeeds "INSERT INTO test_users VALUES (1, 'John', 1);" "Insert John"
    assert_query_succeeds "INSERT INTO test_users VALUES (2, 'Jane', 2);" "Insert Jane"
    assert_query_succeeds "INSERT INTO test_depts VALUES (1, 'IT');" "Insert IT dept"
    assert_query_succeeds "INSERT INTO test_depts VALUES (2, 'HR');" "Insert HR dept"
    
    # 验证JOIN不再返回Empty set
    assert_query_succeeds "SELECT u.name, d.name FROM test_users u JOIN test_depts d ON u.dept_id = d.id;" "JOIN should work"
    assert_query_row_count "SELECT u.name, d.name FROM test_users u JOIN test_depts d ON u.dept_id = d.id;" 2 "Should return 2 joined rows"
    
    local join_result=$(execute_query "SELECT u.name, d.name FROM test_users u JOIN test_depts d ON u.dept_id = d.id;")
    assert_contains "$join_result" "John" "Should contain John in JOIN result"
    assert_contains "$join_result" "IT" "Should contain IT department in JOIN result"
    assert_contains "$join_result" "Jane" "Should contain Jane in JOIN result"
    assert_contains "$join_result" "HR" "Should contain HR department in JOIN result"
}

# 测试GROUP BY问题修复 - 之前不支持
test_group_by_support_fix() {
    echo "Testing GROUP BY support fix..."
    
    # 创建销售表
    assert_query_succeeds "CREATE TABLE test_sales (id INT, department VARCHAR, amount INT);" "Create test_sales table"
    
    # 插入测试数据
    assert_query_succeeds "INSERT INTO test_sales VALUES (1, 'IT', 1000);" "Insert IT sale"
    assert_query_succeeds "INSERT INTO test_sales VALUES (2, 'HR', 800);" "Insert HR sale"
    assert_query_succeeds "INSERT INTO test_sales VALUES (3, 'IT', 1200);" "Insert another IT sale"
    assert_query_succeeds "INSERT INTO test_sales VALUES (4, 'Finance', 1500);" "Insert Finance sale"
    assert_query_succeeds "INSERT INTO test_sales VALUES (5, 'HR', 900);" "Insert another HR sale"
    
    # 验证GROUP BY现在可以工作
    assert_query_succeeds "SELECT department FROM test_sales GROUP BY department;" "GROUP BY should work"
    assert_query_row_count "SELECT department FROM test_sales GROUP BY department;" 3 "Should return 3 departments"
    
    local group_result=$(execute_query "SELECT department FROM test_sales GROUP BY department;")
    assert_contains "$group_result" "IT" "Should contain IT department"
    assert_contains "$group_result" "HR" "Should contain HR department"  
    assert_contains "$group_result" "Finance" "Should contain Finance department"
    
    # 验证GROUP BY计数功能
    # 基于当前实现，GROUP BY会显示COUNT信息
    # IT部门应该有2条记录，HR部门应该有2条记录，Finance部门应该有1条记录
    if echo "$group_result" | grep -E "IT.*2|2.*IT"; then
        record_assertion "PASS" "IT department count" "IT shows correct count"
    else
        record_assertion "FAIL" "IT department count" "IT should show count of 2"
    fi
}

# 测试ORDER BY问题修复 - 之前不支持
test_order_by_support_fix() {
    echo "Testing ORDER BY support fix..."
    
    # 创建学生表
    assert_query_succeeds "CREATE TABLE test_students (id INT, name VARCHAR, grade INT);" "Create test_students table"
    
    # 插入测试数据（故意不按顺序插入）
    assert_query_succeeds "INSERT INTO test_students VALUES (3, 'Charlie', 85);" "Insert Charlie"
    assert_query_succeeds "INSERT INTO test_students VALUES (1, 'Alice', 92);" "Insert Alice"
    assert_query_succeeds "INSERT INTO test_students VALUES (2, 'Bob', 78);" "Insert Bob"
    assert_query_succeeds "INSERT INTO test_students VALUES (4, 'Diana', 88);" "Insert Diana"
    
    # 验证ORDER BY ASC工作
    assert_query_succeeds "SELECT name, grade FROM test_students ORDER BY grade ASC;" "ORDER BY ASC should work"
    
    local asc_result=$(execute_query "SELECT name, grade FROM test_students ORDER BY grade ASC;")
    # 验证排序顺序：Bob(78), Charlie(85), Diana(88), Alice(92)
    local lines=($(echo "$asc_result" | grep "|" | tail -n +2))
    local first_student=$(echo "${lines[0]}" | cut -d'|' -f2 | tr -d ' ')
    assert_equals "Bob" "$first_student" "First student should be Bob (lowest grade)"
    
    # 验证ORDER BY DESC工作
    assert_query_succeeds "SELECT name, grade FROM test_students ORDER BY grade DESC;" "ORDER BY DESC should work"
    
    local desc_result=$(execute_query "SELECT name, grade FROM test_students ORDER BY grade DESC;")
    local desc_lines=($(echo "$desc_result" | grep "|" | tail -n +2))
    local first_desc_student=$(echo "${desc_lines[0]}" | cut -d'|' -f2 | tr -d ' ')
    assert_equals "Alice" "$first_desc_student" "First student should be Alice (highest grade)"
}

# 测试IN表达式问题修复 - 之前返回nil条件
test_in_expression_fix() {
    echo "Testing IN expression fix..."
    
    # 使用现有的test_students表
    # 验证IN表达式现在可以工作
    assert_query_succeeds "SELECT name FROM test_students WHERE grade IN (85, 92);" "IN expression should work"
    assert_query_row_count "SELECT name FROM test_students WHERE grade IN (85, 92);" 2 "Should return 2 students"
    
    local in_result=$(execute_query "SELECT name FROM test_students WHERE grade IN (85, 92);")
    assert_contains "$in_result" "Alice" "Should find Alice (grade 92)"
    assert_contains "$in_result" "Charlie" "Should find Charlie (grade 85)"
    assert_not_contains "$in_result" "Bob" "Should not find Bob (grade 78)"
    assert_not_contains "$in_result" "Diana" "Should not find Diana (grade 88)"
    
    # 测试单值IN表达式
    assert_query_succeeds "SELECT name FROM test_students WHERE grade IN (78);" "Single value IN should work"
    assert_query_row_count "SELECT name FROM test_students WHERE grade IN (78);" 1 "Should return 1 student"
    
    local single_in_result=$(execute_query "SELECT name FROM test_students WHERE grade IN (78);")
    assert_contains "$single_in_result" "Bob" "Should find Bob with single IN"
}

# 测试LIKE表达式问题修复 - 之前在常规执行器中不工作
test_like_expression_fix() {
    echo "Testing LIKE expression fix..."
    
    # 创建名字测试表
    assert_query_succeeds "CREATE TABLE test_names (id INT, name VARCHAR);" "Create test_names table"
    
    # 插入测试数据
    assert_query_succeeds "INSERT INTO test_names VALUES (1, 'Alice');" "Insert Alice"
    assert_query_succeeds "INSERT INTO test_names VALUES (2, 'Andrew');" "Insert Andrew"
    assert_query_succeeds "INSERT INTO test_names VALUES (3, 'Bob');" "Insert Bob"
    assert_query_succeeds "INSERT INTO test_names VALUES (4, 'Charlie');" "Insert Charlie"
    
    # 验证LIKE表达式现在在常规执行器中工作
    assert_query_succeeds "SELECT name FROM test_names WHERE name LIKE 'A%';" "LIKE pattern should work"
    assert_query_row_count "SELECT name FROM test_names WHERE name LIKE 'A%';" 2 "Should find 2 names starting with A"
    
    local like_result=$(execute_query "SELECT name FROM test_names WHERE name LIKE 'A%';")
    assert_contains "$like_result" "Alice" "Should find Alice"
    assert_contains "$like_result" "Andrew" "Should find Andrew"
    assert_not_contains "$like_result" "Bob" "Should not find Bob"
    assert_not_contains "$like_result" "Charlie" "Should not find Charlie"
    
    # 测试不同的LIKE模式
    assert_query_succeeds "SELECT name FROM test_names WHERE name LIKE '%ie';" "LIKE suffix pattern should work"
    
    local suffix_result=$(execute_query "SELECT name FROM test_names WHERE name LIKE '%ie';")
    assert_contains "$suffix_result" "Charlie" "Should find Charlie (ends with 'ie')"
    assert_not_contains "$suffix_result" "Alice" "Should not find Alice (ends with 'ce')"
}

# 测试复合WHERE条件修复 - 之前AND/OR不工作
test_compound_where_fix() {
    echo "Testing compound WHERE conditions fix..."
    
    # 使用现有的test_students表测试复合条件
    # 测试AND条件
    assert_query_succeeds "SELECT name FROM test_students WHERE grade > 80 AND grade < 90;" "AND condition should work"
    assert_query_row_count "SELECT name FROM test_students WHERE grade > 80 AND grade < 90;" 2 "Should find 2 students"
    
    local and_result=$(execute_query "SELECT name FROM test_students WHERE grade > 80 AND grade < 90;")
    assert_contains "$and_result" "Charlie" "Should find Charlie (grade 85)"
    assert_contains "$and_result" "Diana" "Should find Diana (grade 88)"
    assert_not_contains "$and_result" "Alice" "Should not find Alice (grade 92)"
    assert_not_contains "$and_result" "Bob" "Should not find Bob (grade 78)"
    
    # 测试OR条件
    assert_query_succeeds "SELECT name FROM test_students WHERE grade = 78 OR grade = 92;" "OR condition should work"
    assert_query_row_count "SELECT name FROM test_students WHERE grade = 78 OR grade = 92;" 2 "Should find 2 students"
    
    local or_result=$(execute_query "SELECT name FROM test_students WHERE grade = 78 OR grade = 92;")
    assert_contains "$or_result" "Bob" "Should find Bob (grade 78)"
    assert_contains "$or_result" "Alice" "Should find Alice (grade 92)"
    assert_not_contains "$or_result" "Charlie" "Should not find Charlie (grade 85)"
    assert_not_contains "$or_result" "Diana" "Should not find Diana (grade 88)"
}

# 测试执行器选择机制修复
test_executor_selection_fix() {
    echo "Testing executor selection mechanism fix..."
    
    # 验证复杂查询正确回退到常规执行器
    # JOIN查询应该使用常规执行器
    assert_query_succeeds "SELECT u.name, d.name FROM test_users u JOIN test_depts d ON u.dept_id = d.id;" "JOIN should use regular executor"
    
    # GROUP BY查询应该使用常规执行器
    assert_query_succeeds "SELECT department FROM test_sales GROUP BY department;" "GROUP BY should use regular executor"
    
    # ORDER BY查询应该使用常规执行器
    assert_query_succeeds "SELECT name FROM test_students ORDER BY grade;" "ORDER BY should use regular executor"
    
    # LIKE查询应该使用常规执行器
    assert_query_succeeds "SELECT name FROM test_names WHERE name LIKE 'A%';" "LIKE should use regular executor"
    
    # 简单查询仍可以使用向量化执行器（如果支持）
    assert_query_succeeds "SELECT * FROM test_students;" "Simple queries should still work"
}

# 清理回归测试数据
cleanup_regression_test_data() {
    execute_query "DROP DATABASE regression_test;" >/dev/null 2>&1
}

# 主测试函数
main() {
    # 启动数据库
    start_database || {
        skip_test_suite "Regression Tests" "Failed to start database"
        return 1
    }
    
    # 运行所有回归测试
    setup_regression_test_data
    test_join_empty_set_fix
    test_group_by_support_fix
    test_order_by_support_fix
    test_in_expression_fix
    test_like_expression_fix
    test_compound_where_fix
    test_executor_selection_fix
    
    # 清理
    cleanup_regression_test_data
    
    return 0
}

# 如果直接执行此脚本
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi