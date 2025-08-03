#!/bin/bash

# CRUD操作单元测试
# 测试基础的增删改查功能

source "$(dirname "${BASH_SOURCE[0]}")/../../utils/test_runner.sh"

# 测试CREATE操作
test_create_operations() {
    echo "Testing CREATE operations..."
    
    # 测试创建数据库
    assert_query_succeeds "CREATE DATABASE test_crud;" "Should create database"
    
    # 测试使用数据库
    assert_query_succeeds "USE test_crud;" "Should switch to database"
    
    # 测试创建表
    assert_query_succeeds "CREATE TABLE test_table (id INT, name VARCHAR, age INT);" "Should create table"
    
    # 测试重复创建数据库（应该失败）
    assert_query_fails "CREATE DATABASE test_crud;" "duplicate" "Should fail on duplicate database"
}

# 测试INSERT操作
test_insert_operations() {
    echo "Testing INSERT operations..."
    
    # 基础插入测试
    assert_query_succeeds "INSERT INTO test_table VALUES (1, 'Alice', 25);" "Should insert single row"
    
    # 批量插入测试
    assert_query_succeeds "INSERT INTO test_table VALUES (2, 'Bob', 30);" "Should insert second row"
    assert_query_succeeds "INSERT INTO test_table VALUES (3, 'Charlie', 35);" "Should insert third row"
    
    # 验证插入结果
    assert_query_row_count "SELECT * FROM test_table;" 3 "Should have 3 rows after inserts"
}

# 测试SELECT操作
test_select_operations() {
    echo "Testing SELECT operations..."
    
    # 基础查询
    assert_query_succeeds "SELECT * FROM test_table;" "Should select all rows"
    
    # 带条件查询
    assert_query_succeeds "SELECT * FROM test_table WHERE age > 25;" "Should select with WHERE condition"
    assert_query_row_count "SELECT * FROM test_table WHERE age > 25;" 2 "Should return 2 rows with age > 25"
    
    # 特定列查询
    assert_query_succeeds "SELECT name, age FROM test_table;" "Should select specific columns"
    
    # 查询特定值
    assert_query_contains_value "SELECT name FROM test_table WHERE id = 1;" "Alice" "Should find Alice with id=1"
}

# 测试UPDATE操作
test_update_operations() {
    echo "Testing UPDATE operations..."
    
    # 基础更新
    assert_query_succeeds "UPDATE test_table SET age = 26 WHERE id = 1;" "Should update Alice's age"
    
    # 验证更新结果
    assert_query_contains_value "SELECT age FROM test_table WHERE id = 1;" "26" "Alice's age should be 26"
    
    # 多字段更新
    assert_query_succeeds "UPDATE test_table SET name = 'Alice Smith', age = 27 WHERE id = 1;" "Should update multiple fields"
}

# 测试DELETE操作
test_delete_operations() {
    echo "Testing DELETE operations..."
    
    # 删除特定行
    assert_query_succeeds "DELETE FROM test_table WHERE id = 3;" "Should delete Charlie"
    
    # 验证删除结果
    assert_query_row_count "SELECT * FROM test_table;" 2 "Should have 2 rows after delete"
    
    # 验证特定记录被删除
    assert_query_row_count "SELECT * FROM test_table WHERE name = 'Charlie';" 0 "Charlie should be deleted"
}

# 主测试函数
main() {
    # 启动数据库
    start_database || {
        skip_test_suite "CRUD Operations" "Failed to start database"
        return 1
    }
    
    # 运行所有测试
    test_create_operations
    test_insert_operations  
    test_select_operations
    test_update_operations
    test_delete_operations
    
    # 清理
    execute_query "DROP DATABASE test_crud;" >/dev/null 2>&1
    
    return 0
}

# 如果直接执行此脚本
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi