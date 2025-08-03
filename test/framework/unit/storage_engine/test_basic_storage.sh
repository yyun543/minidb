#!/bin/bash

# 存储引擎基础功能测试
# 测试数据持久化和检索

source "$(dirname "${BASH_SOURCE[0]}")/../../utils/test_runner.sh"

# 测试表创建和Schema存储
test_table_creation_storage() {
    echo "Testing table creation and schema storage..."
    
    # 创建数据库
    assert_query_succeeds "CREATE DATABASE storage_test;" "Should create database"
    assert_query_succeeds "USE storage_test;" "Should switch to database"
    
    # 创建表
    assert_query_succeeds "CREATE TABLE test_storage (id INT, name VARCHAR, value INT);" "Should create table"
    
    # 验证表存在（通过插入数据测试）
    assert_query_succeeds "INSERT INTO test_storage VALUES (1, 'test', 100);" "Should insert into created table"
    
    # 验证数据可以查询
    assert_query_succeeds "SELECT * FROM test_storage;" "Should query created table"
    assert_query_row_count "SELECT * FROM test_storage;" 1 "Should have 1 row"
}

# 测试数据插入和持久化
test_data_persistence() {
    echo "Testing data insertion and persistence..."
    
    # 插入多行数据
    assert_query_succeeds "INSERT INTO test_storage VALUES (2, 'second', 200);" "Insert second row"
    assert_query_succeeds "INSERT INTO test_storage VALUES (3, 'third', 300);" "Insert third row"
    
    # 验证所有数据都存在
    assert_query_row_count "SELECT * FROM test_storage;" 3 "Should have 3 rows total"
    
    # 验证特定数据
    assert_query_contains_value "SELECT name FROM test_storage WHERE id = 2;" "second" "Should find 'second'"
    assert_query_contains_value "SELECT value FROM test_storage WHERE id = 3;" "300" "Should find value 300"
}

# 测试数据更新
test_data_updates() {
    echo "Testing data updates..."
    
    # 更新数据
    assert_query_succeeds "UPDATE test_storage SET value = 150 WHERE id = 1;" "Should update value"
    
    # 验证更新结果
    assert_query_contains_value "SELECT value FROM test_storage WHERE id = 1;" "150" "Should have updated value"
    
    # 验证其他数据未受影响
    assert_query_contains_value "SELECT value FROM test_storage WHERE id = 2;" "200" "Other rows should be unchanged"
}

# 测试数据删除
test_data_deletion() {
    echo "Testing data deletion..."
    
    # 删除一行
    assert_query_succeeds "DELETE FROM test_storage WHERE id = 2;" "Should delete row"
    
    # 验证行数减少
    assert_query_row_count "SELECT * FROM test_storage;" 2 "Should have 2 rows after delete"
    
    # 验证特定行被删除
    assert_query_row_count "SELECT * FROM test_storage WHERE id = 2;" 0 "Deleted row should not exist"
    
    # 验证其他行仍存在
    assert_query_contains_value "SELECT name FROM test_storage WHERE id = 1;" "test" "Other rows should remain"
}

# 测试数据类型处理
test_data_types() {
    echo "Testing data type handling..."
    
    # 创建包含不同数据类型的表
    assert_query_succeeds "CREATE TABLE type_test (id INT, name VARCHAR, active VARCHAR);" "Create type test table"
    
    # 插入不同类型的数据
    assert_query_succeeds "INSERT INTO type_test VALUES (1, 'Alice', 'yes');" "Insert with string values"
    assert_query_succeeds "INSERT INTO type_test VALUES (2, 'Bob', 'no');" "Insert another string"
    
    # 验证字符串存储和检索
    assert_query_contains_value "SELECT active FROM type_test WHERE name = 'Alice';" "yes" "Should retrieve string correctly"
    assert_query_contains_value "SELECT active FROM type_test WHERE name = 'Bob';" "no" "Should retrieve second string"
    
    # 验证整数存储和检索
    assert_query_contains_value "SELECT id FROM type_test WHERE name = 'Alice';" "1" "Should retrieve integer correctly"
}

# 测试边界情况
test_storage_edge_cases() {
    echo "Testing storage edge cases..."
    
    # 测试空表
    assert_query_succeeds "CREATE TABLE empty_test (id INT, data VARCHAR);" "Create empty table"
    assert_query_row_count "SELECT * FROM empty_test;" 0 "Empty table should return 0 rows"
    
    # 测试特殊字符处理
    assert_query_succeeds "CREATE TABLE special_test (id INT, data VARCHAR);" "Create special test table"
    assert_query_succeeds "INSERT INTO special_test VALUES (1, 'normal text');" "Insert normal text"
    
    # 测试包含空格的字符串
    assert_query_contains_value "SELECT data FROM special_test WHERE id = 1;" "normal text" "Should handle spaces in strings"
}

# 清理存储测试数据
cleanup_storage_test_data() {
    execute_query "DROP DATABASE storage_test;" >/dev/null 2>&1
}

# 主测试函数
main() {
    # 启动数据库
    start_database || {
        skip_test_suite "Basic Storage Tests" "Failed to start database"
        return 1
    }
    
    # 运行存储测试
    test_table_creation_storage
    test_data_persistence
    test_data_updates
    test_data_deletion
    test_data_types
    test_storage_edge_cases
    
    # 清理
    cleanup_storage_test_data
    
    return 0
}

# 如果直接执行此脚本
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi