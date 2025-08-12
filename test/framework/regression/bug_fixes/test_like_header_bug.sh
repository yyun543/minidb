#!/bin/bash

# Test script for LIKE query header bug fix
# Bug: SELECT * FROM users WHERE name LIKE 'J%' 
# returns "*" as header instead of actual column names

source "$(dirname "$0")/../../utils/test_runner.sh"
source "$(dirname "$0")/../../utils/db_helper.sh"
source "$(dirname "$0")/../../config/test_data.sh"

TEST_NAME="LIKE Query Header Bug Fix"
TEST_DB="like_header_test_db"

# Test setup
setup_test() {
    log_info "Setting up $TEST_NAME"
    
    # Start server
    start_minidb_server
    
    # Setup test database and tables
    execute_sql "CREATE DATABASE $TEST_DB;"
    execute_sql "USE $TEST_DB;"
    
    # Create test table
    execute_sql "CREATE TABLE users (id INT, name VARCHAR, email VARCHAR, age INT, created_at VARCHAR);"
    
    # Insert test data
    execute_sql "INSERT INTO users VALUES (1, 'John Doe', 'john@example.com', 25, '2024-01-01');"
    execute_sql "INSERT INTO users VALUES (2, 'Jane Smith', 'jane@example.com', 30, '2024-01-02');"
    execute_sql "INSERT INTO users VALUES (3, 'Bob Wilson', 'bob@example.com', 35, '2024-01-03');"
}

# Test LIKE query with SELECT * - should show proper column headers
test_like_select_all_headers() {
    log_info "Testing LIKE query with SELECT * headers"
    
    local query="SELECT * FROM users WHERE name LIKE 'J%';"
    local result=$(execute_sql "$query")
    
    log_debug "LIKE SELECT * result: $result"
    
    # Check that result doesn't have "*" as the only header
    if echo "$result" | grep -q "| \* *|"; then
        log_error "LIKE query shows '*' as header instead of column names"
        log_debug "Result: $result"
        return 1
    fi
    
    # Check that result has proper column headers: id, name, email, age, created_at
    local expected_headers=("id" "name" "email" "age" "created_at")
    local header_line=$(echo "$result" | head -1)
    
    for header in "${expected_headers[@]}"; do
        if ! echo "$header_line" | grep -q "$header"; then
            log_error "Missing expected header: $header"
            log_debug "Header line: $header_line"
            return 1
        fi
    done
    
    log_success "LIKE query shows correct column headers"
    return 0
}

# Test LIKE query with specific columns
test_like_specific_columns() {
    log_info "Testing LIKE query with specific columns"
    
    local query="SELECT name, email FROM users WHERE name LIKE 'J%';"
    local result=$(execute_sql "$query")
    
    log_debug "LIKE specific columns result: $result"
    
    # Check that result has exactly the requested columns
    local header_line=$(echo "$result" | head -1)
    
    if echo "$header_line" | grep -q "name" && echo "$header_line" | grep -q "email"; then
        log_success "LIKE query with specific columns shows correct headers"
    else
        log_error "LIKE query with specific columns shows incorrect headers"
        log_debug "Expected: name, email"
        log_debug "Actual header line: $header_line"
        return 1
    fi
    
    # Check that it doesn't show unwanted columns
    if echo "$header_line" | grep -q "id\|age\|created_at"; then
        log_error "LIKE query shows unwanted columns"
        log_debug "Header line: $header_line"
        return 1
    fi
    
    return 0
}

# Test other pattern matching operators for consistency
test_other_pattern_operators() {
    log_info "Testing other pattern operators for header consistency"
    
    # Test different LIKE patterns
    local patterns=("'J%'" "'%Smith'" "'%o%'")
    
    for pattern in "${patterns[@]}"; do
        local query="SELECT * FROM users WHERE name LIKE $pattern;"
        local result=$(execute_sql "$query")
        
        log_debug "Pattern $pattern result: $result"
        
        # Check that headers are not "*"
        if echo "$result" | grep -q "| \* *|"; then
            log_error "Pattern $pattern shows '*' as header"
            return 1
        fi
    done
    
    log_success "All pattern operators show correct headers"
    return 0
}

# Test LIKE with no results
test_like_no_results() {
    log_info "Testing LIKE query with no matching results"
    
    local query="SELECT * FROM users WHERE name LIKE 'Z%';"
    local result=$(execute_sql "$query")
    
    log_debug "No results LIKE query: $result"
    
    # Even with no results, headers should be correct
    if echo "$result" | grep -q "Empty set"; then
        log_success "LIKE query with no results handled correctly"
        return 0
    fi
    
    # If not empty set, check headers are still correct
    if echo "$result" | grep -q "| \* *|"; then
        log_error "LIKE query with no results shows '*' as header"
        return 1
    fi
    
    return 0
}

# Test mixed conditions with LIKE
test_like_with_other_conditions() {
    log_info "Testing LIKE with other WHERE conditions"
    
    local query="SELECT name, age FROM users WHERE name LIKE 'J%' AND age > 25;"
    local result=$(execute_sql "$query")
    
    log_debug "Mixed conditions result: $result"
    
    # Check headers are correct
    local header_line=$(echo "$result" | head -1)
    
    if echo "$header_line" | grep -q "name" && echo "$header_line" | grep -q "age"; then
        log_success "LIKE with mixed conditions shows correct headers"
        return 0
    else
        log_error "LIKE with mixed conditions shows incorrect headers"
        log_debug "Header line: $header_line"
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
    
    test_like_select_all_headers || failed=$((failed + 1))
    test_like_specific_columns || failed=$((failed + 1))
    test_other_pattern_operators || failed=$((failed + 1))
    test_like_no_results || failed=$((failed + 1))
    test_like_with_other_conditions || failed=$((failed + 1))
    
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