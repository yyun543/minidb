#!/bin/bash

# 数据库辅助函数
# 基于KISS原则的数据库操作封装

source "$(dirname "${BASH_SOURCE[0]}")/../config/test_config.sh"

# 数据库进程管理
DB_PID=""
DB_RUNNING=false

# 启动数据库服务器
start_database() {
    local timeout="${1:-$DB_START_WAIT}"
    
    if [[ "$DB_RUNNING" == "true" ]]; then
        echo "Database is already running (PID: $DB_PID)"
        return 0
    fi
    
    # 确保二进制文件存在
    if [[ ! -f "$DB_BINARY" ]]; then
        echo "Building database binary..."
        cd "$DB_BUILD_DIR" && go build -o minidb ./cmd/server
        if [[ $? -ne 0 ]]; then
            echo "Failed to build database binary" >&2
            return 1
        fi
    fi
    
    # 启动数据库
    echo "Starting database server..."
    "$DB_BINARY" &
    DB_PID=$!
    
    # 等待服务器启动
    sleep "$timeout"
    
    # 检查进程是否还在运行
    if kill -0 "$DB_PID" 2>/dev/null; then
        DB_RUNNING=true
        echo "Database started successfully (PID: $DB_PID)"
        return 0
    else
        echo "Failed to start database server" >&2
        return 1
    fi
}

# 停止数据库服务器
stop_database() {
    if [[ "$DB_RUNNING" == "true" && -n "$DB_PID" ]]; then
        echo "Stopping database server (PID: $DB_PID)..."
        kill "$DB_PID" 2>/dev/null
        wait "$DB_PID" 2>/dev/null
        DB_RUNNING=false
        DB_PID=""
        echo "Database stopped."
    fi
    
    # 清理任何残留的数据库进程
    pkill -f minidb 2>/dev/null || true
}

# 重启数据库服务器
restart_database() {
    stop_database
    sleep 1
    start_database
}

# 检查数据库是否运行
is_database_running() {
    if [[ "$DB_RUNNING" == "true" && -n "$DB_PID" ]] && kill -0 "$DB_PID" 2>/dev/null; then
        return 0
    else
        DB_RUNNING=false
        return 1
    fi
}

# 等待数据库就绪
wait_for_database() {
    local timeout="${1:-$DB_TIMEOUT}"
    local counter=0
    
    while [[ $counter -lt $timeout ]]; do
        if echo "SHOW DATABASES;" | nc -w 1 "$DB_HOST" "$DB_PORT" >/dev/null 2>&1; then
            return 0
        fi
        sleep 1
        ((counter++))
    done
    
    echo "Database did not become ready within ${timeout} seconds" >&2
    return 1
}

# 执行SQL查询
execute_query() {
    local query="$1"
    local timeout="${2:-5}"
    
    if ! is_database_running; then
        echo "Database is not running" >&2
        return 1
    fi
    
    # 使用nc发送查询（macOS兼容性处理）
    local result
    if command -v timeout >/dev/null 2>&1; then
        result=$(timeout "$timeout" bash -c "echo '$query' | nc '$DB_HOST' '$DB_PORT'" 2>&1)
    elif command -v gtimeout >/dev/null 2>&1; then
        result=$(gtimeout "$timeout" bash -c "echo '$query' | nc '$DB_HOST' '$DB_PORT'" 2>&1)
    else
        # 没有timeout命令，直接执行（可能会阻塞）
        result=$(bash -c "echo '$query' | nc '$DB_HOST' '$DB_PORT'" 2>&1)
    fi
    local exit_code=$?
    
    if [[ $exit_code -eq 124 ]]; then
        echo "Query timeout after ${timeout} seconds" >&2
        return 1
    elif [[ $exit_code -ne 0 ]]; then
        echo "Query execution failed: $result" >&2
        return 1
    fi
    
    echo "$result"
}

# 执行SQL文件
execute_sql_file() {
    local sql_file="$1"
    local timeout="${2:-10}"
    
    if [[ ! -f "$sql_file" ]]; then
        echo "SQL file not found: $sql_file" >&2
        return 1
    fi
    
    if ! is_database_running; then
        echo "Database is not running" >&2
        return 1
    fi
    
    local result
    if command -v timeout >/dev/null 2>&1; then
        result=$(timeout "$timeout" bash -c "nc '$DB_HOST' '$DB_PORT' < '$sql_file'" 2>&1)
    elif command -v gtimeout >/dev/null 2>&1; then
        result=$(gtimeout "$timeout" bash -c "nc '$DB_HOST' '$DB_PORT' < '$sql_file'" 2>&1)
    else
        result=$(bash -c "nc '$DB_HOST' '$DB_PORT' < '$sql_file'" 2>&1)
    fi
    local exit_code=$?
    
    if [[ $exit_code -eq 124 ]]; then
        echo "SQL file execution timeout after ${timeout} seconds" >&2
        return 1
    elif [[ $exit_code -ne 0 ]]; then
        echo "SQL file execution failed: $result" >&2
        return 1
    fi
    
    echo "$result"
}

# 设置测试数据库环境
setup_test_database() {
    local setup_type="${1:-basic}"
    
    case "$setup_type" in
        "basic")
            setup_basic_test_data | execute_query_batch
            ;;
        "complex")
            setup_complex_query_data | execute_query_batch
            ;;
        "edge")
            setup_edge_case_data | execute_query_batch
            ;;
        "regression")
            setup_regression_data | execute_query_batch
            ;;
        *)
            echo "Unknown setup type: $setup_type" >&2
            return 1
            ;;
    esac
}

# 批量执行SQL语句
execute_query_batch() {
    local timeout="${1:-30}"
    
    if ! is_database_running; then
        echo "Database is not running" >&2
        return 1
    fi
    
    if command -v timeout >/dev/null 2>&1; then
        timeout "$timeout" nc "$DB_HOST" "$DB_PORT"
    elif command -v gtimeout >/dev/null 2>&1; then
        gtimeout "$timeout" nc "$DB_HOST" "$DB_PORT"
    else
        nc "$DB_HOST" "$DB_PORT"
    fi
}

# 清理测试数据
cleanup_test_data() {
    local databases=("testdb" "complexdb" "edgedb" "regressiondb" "perfdb")
    
    for db in "${databases[@]}"; do
        execute_query "DROP DATABASE IF EXISTS $db;" >/dev/null 2>&1
    done
}

# 获取查询执行时间
measure_query_time() {
    local query="$1"
    local iterations="${2:-1}"
    local total_time=0
    
    for ((i=1; i<=iterations; i++)); do
        local start_time=$(date +%s.%N)
        execute_query "$query" >/dev/null
        local end_time=$(date +%s.%N)
        local duration=$(echo "$end_time - $start_time" | bc -l)
        total_time=$(echo "$total_time + $duration" | bc -l)
    done
    
    local avg_time=$(echo "scale=3; $total_time / $iterations" | bc -l)
    echo "$avg_time"
}

# 获取表行数
get_table_row_count() {
    local database="$1"
    local table="$2"
    
    local result=$(execute_query "USE $database; SELECT COUNT(*) FROM $table;" 2>/dev/null)
    if [[ $? -eq 0 ]]; then
        echo "$result" | grep -oE '[0-9]+' | tail -1
    else
        echo "0"
    fi
}

# 检查表是否存在
table_exists() {
    local database="$1"
    local table="$2"
    
    local result=$(execute_query "USE $database; SELECT * FROM $table LIMIT 1;" 2>&1)
    if echo "$result" | grep -q "Error:"; then
        return 1
    else
        return 0
    fi
}

# 数据库健康检查
database_health_check() {
    local checks_passed=0
    local total_checks=4
    
    echo "Performing database health check..."
    
    # 检查1: 进程是否运行
    if is_database_running; then
        echo "✓ Database process is running"
        ((checks_passed++))
    else
        echo "✗ Database process is not running"
    fi
    
    # 检查2: 端口是否可用
    if nc -z "$DB_HOST" "$DB_PORT" 2>/dev/null; then
        echo "✓ Database port is accessible"
        ((checks_passed++))
    else
        echo "✗ Database port is not accessible"
    fi
    
    # 检查3: 基本查询是否工作
    if execute_query "SHOW DATABASES;" >/dev/null 2>&1; then
        echo "✓ Basic queries work"
        ((checks_passed++))
    else
        echo "✗ Basic queries fail"
    fi
    
    # 检查4: DDL操作是否工作
    if execute_query "CREATE DATABASE health_check; DROP DATABASE health_check;" >/dev/null 2>&1; then
        echo "✓ DDL operations work"
        ((checks_passed++))
    else
        echo "✗ DDL operations fail"
    fi
    
    echo "Health check: $checks_passed/$total_checks checks passed"
    
    if [[ $checks_passed -eq $total_checks ]]; then
        return 0
    else
        return 1
    fi
}

# 注册清理函数
cleanup_database() {
    stop_database
    cleanup_test_data
}

# 在脚本退出时清理
trap cleanup_database EXIT