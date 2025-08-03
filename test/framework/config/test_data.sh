#!/bin/bash

# 测试数据定义
# 遵循DRY原则，统一管理测试数据

# 基础测试数据库
setup_basic_test_data() {
    cat << 'EOF'
CREATE DATABASE testdb;
USE testdb;
CREATE TABLE users (id INT, name VARCHAR, age INT);
CREATE TABLE departments (id INT, name VARCHAR);
INSERT INTO users VALUES (1, 'Alice', 25);
INSERT INTO users VALUES (2, 'Bob', 30);
INSERT INTO users VALUES (3, 'Charlie', 35);
INSERT INTO users VALUES (4, 'Diana', 28);
INSERT INTO departments VALUES (1, 'Engineering');
INSERT INTO departments VALUES (2, 'Marketing');
EOF
}

# 复杂查询测试数据
setup_complex_query_data() {
    cat << 'EOF'
CREATE DATABASE complexdb;
USE complexdb;
CREATE TABLE employees (id INT, name VARCHAR, age INT, dept_id INT, salary INT);
CREATE TABLE departments (id INT, name VARCHAR, budget INT);
INSERT INTO employees VALUES (1, 'Alice', 30, 1, 75000);
INSERT INTO employees VALUES (2, 'Bob', 25, 2, 65000);
INSERT INTO employees VALUES (3, 'Charlie', 35, 1, 85000);
INSERT INTO employees VALUES (4, 'Diana', 28, 2, 70000);
INSERT INTO employees VALUES (5, 'Eve', 32, 1, 80000);
INSERT INTO departments VALUES (1, 'Engineering', 500000);
INSERT INTO departments VALUES (2, 'Marketing', 300000);
EOF
}

# 边界测试数据
setup_edge_case_data() {
    cat << 'EOF'
CREATE DATABASE edgedb;
USE edgedb;
CREATE TABLE empty_table (id INT, name VARCHAR);
CREATE TABLE single_row (id INT, value VARCHAR);
CREATE TABLE special_chars (id INT, data VARCHAR);
INSERT INTO single_row VALUES (1, 'only_row');
INSERT INTO special_chars VALUES (1, 'normal');
INSERT INTO special_chars VALUES (2, 'with spaces');
INSERT INTO special_chars VALUES (3, 'with''quote');
EOF
}

# 性能测试数据生成
generate_performance_data() {
    local row_count=${1:-1000}
    cat << EOF
CREATE DATABASE perfdb;
USE perfdb;
CREATE TABLE large_table (id INT, name VARCHAR, value INT);
EOF
    
    for ((i=1; i<=row_count; i++)); do
        echo "INSERT INTO large_table VALUES ($i, 'user$i', $((i % 100)));"
    done
}

# 回归测试数据（基于已修复的问题）
setup_regression_data() {
    cat << 'EOF'
CREATE DATABASE regressiondb;
USE regressiondb;
-- 测试JOIN问题修复
CREATE TABLE test_users (id INT, name VARCHAR, dept_id INT);
CREATE TABLE test_depts (id INT, name VARCHAR);
INSERT INTO test_users VALUES (1, 'John', 1);
INSERT INTO test_users VALUES (2, 'Jane', 2);
INSERT INTO test_depts VALUES (1, 'IT');
INSERT INTO test_depts VALUES (2, 'HR');

-- 测试GROUP BY问题修复
CREATE TABLE sales (id INT, department VARCHAR, amount INT);
INSERT INTO sales VALUES (1, 'IT', 1000);
INSERT INTO sales VALUES (2, 'HR', 800);
INSERT INTO sales VALUES (3, 'IT', 1200);

-- 测试ORDER BY问题修复
CREATE TABLE students (id INT, name VARCHAR, grade INT);
INSERT INTO students VALUES (3, 'Charlie', 85);
INSERT INTO students VALUES (1, 'Alice', 92);
INSERT INTO students VALUES (2, 'Bob', 78);
EOF
}