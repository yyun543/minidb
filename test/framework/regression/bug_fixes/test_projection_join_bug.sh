#!/bin/bash

# Test script for JOIN projection bug fix
# Bug: SELECT u.name, o.amount, o.order_date FROM users u JOIN orders o ON u.id = o.user_id WHERE u.age > 25
# returns all columns instead of just the projected ones

source "$(dirname "$0")/../../utils/test_runner.sh"
source "$(dirname "$0")/../../utils/db_helper.sh"
source "$(dirname "$0")/../../config/test_data.sh"

TEST_NAME="JOIN Projection Bug Fix"
TEST_DB="projection_test_db"

# Test setup
setup_test() {
    log_info "Setting up $TEST_NAME"
    
    # Start server
    start_minidb_server
    
    # Setup test database and tables
    execute_sql "CREATE DATABASE $TEST_DB;"
    execute_sql "USE $TEST_DB;"
    
    # Create users table
    execute_sql "CREATE TABLE users (id INT, name VARCHAR, email VARCHAR, age INT, created_at VARCHAR);"
    
    # Create orders table  
    execute_sql "CREATE TABLE orders (id INT, user_id INT, amount INT, order_date VARCHAR);"
    
    # Insert test data
    execute_sql "INSERT INTO users VALUES (1, 'John Doe', 'john@example.com', 25, '2024-01-01');"
    execute_sql "INSERT INTO users VALUES (2, 'Jane Smith', 'jane@example.com', 30, '2024-01-02');"
    execute_sql "INSERT INTO users VALUES (3, 'Bob Wilson', 'bob@example.com', 35, '2024-01-03');"
    
    execute_sql "INSERT INTO orders VALUES (1, 1, 100, '2024-01-05');"
    execute_sql "INSERT INTO orders VALUES (2, 2, 250, '2024-01-06');"
    execute_sql "INSERT INTO orders VALUES (3, 1, 150, '2024-01-07');"
}

# Test JOIN projection - should only return specified columns
test_join_projection() {
    log_info "Testing JOIN projection bug fix"
    
    local query="SELECT u.name, o.amount, o.order_date FROM users u JOIN orders o ON u.id = o.user_id WHERE u.age > 25;"
    local result=$(execute_sql "$query")
    
    log_debug "Query result: $result"
    
    # Expected headers should be: u.name | o.amount | o.order_date
    # NOT all columns from both tables
    
    # Check that result has exactly 3 columns (not all columns from both tables)
    local header_line=$(echo "$result" | head -1)
    local column_count=$(echo "$header_line" | grep -o "|" | wc -l)
    
    # Should have 4 pipe symbols (for 3 columns: | col1 | col2 | col3 |)
    if [ "$column_count" -eq 4 ]; then
        log_success "JOIN projection returns correct number of columns"
    else
        log_error "JOIN projection returns wrong number of columns. Expected 4 pipes, got $column_count"
        return 1
    fi
    
    # Check that headers contain the expected column names
    if echo "$result" | grep -q "u.name" && echo "$result" | grep -q "o.amount" && echo "$result" | grep -q "o.order_date"; then
        log_success "JOIN projection headers are correct"
    else
        log_error "JOIN projection headers are incorrect"
        log_debug "Expected headers: u.name, o.amount, o.order_date"
        log_debug "Actual result: $result"
        return 1
    fi
    
    # Check that we don't have unwanted columns like id, email, etc.
    if echo "$result" | grep -q "email\|created_at\|user_id"; then
        log_error "JOIN projection contains unwanted columns"
        log_debug "Result should not contain: email, created_at, user_id"
        log_debug "Actual result: $result"
        return 1
    fi
    
    return 0
}

# Test simple JOIN without WHERE clause
test_simple_join_projection() {
    log_info "Testing simple JOIN projection"
    
    local query="SELECT u.name, o.amount FROM users u JOIN orders o ON u.id = o.user_id;"
    local result=$(execute_sql "$query")
    
    log_debug "Simple JOIN result: $result"
    
    # Should have exactly 2 columns
    local header_line=$(echo "$result" | head -1)
    local column_count=$(echo "$header_line" | grep -o "|" | wc -l)
    
    if [ "$column_count" -eq 3 ]; then
        log_success "Simple JOIN projection returns correct number of columns"
        return 0
    else
        log_error "Simple JOIN projection returns wrong number of columns. Expected 3 pipes, got $column_count"
        return 1
    fi
}

# Test JOIN with SELECT *
test_join_select_all() {
    log_info "Testing JOIN with SELECT *"
    
    local query="SELECT * FROM users u JOIN orders o ON u.id = o.user_id;"
    local result=$(execute_sql "$query")
    
    log_debug "JOIN SELECT * result: $result"
    
    # SELECT * should return all columns from both tables
    # users: id, name, email, age, created_at (5 columns)
    # orders: id, user_id, amount, order_date (4 columns)
    # Total: 9 columns, so 10 pipes
    local header_line=$(echo "$result" | head -1)
    local column_count=$(echo "$header_line" | grep -o "|" | wc -l)
    
    if [ "$column_count" -eq 10 ]; then
        log_success "JOIN SELECT * returns all columns"
        return 0
    else
        log_error "JOIN SELECT * returns wrong number of columns. Expected 10 pipes, got $column_count"
        return 1
    fi
}

# Test cleanup
cleanup_test() {
    log_info "Cleaning up $TEST_NAME"
    execute_sql "DROP DATABASE $TEST_DB;" 2>/dev/null || true
    stop_minidb_server
}

# Main test execution
main() {
    log_info "Starting $TEST_NAME"
    
    setup_test || { log_error "Setup failed"; exit 1; }
    
    # Run tests
    local failed=0
    
    test_join_projection || failed=$((failed + 1))
    test_simple_join_projection || failed=$((failed + 1))
    test_join_select_all || failed=$((failed + 1))
    
    cleanup_test
    
    if [ $failed -eq 0 ]; then
        log_success "$TEST_NAME: All tests passed"
        exit 0
    else
        log_error "$TEST_NAME: $failed test(s) failed"
        exit 1
    fi
}

# Run if executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi