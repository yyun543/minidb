#!/bin/bash

# 断言库 - 基于TDD思想的测试断言
# 提供简单、直观的断言函数

source "$(dirname "${BASH_SOURCE[0]}")/../config/test_config.sh"

# 全局测试统计
ASSERTION_TOTAL=0
ASSERTION_PASSED=0
ASSERTION_FAILED=0

# 记录当前测试上下文
CURRENT_TEST_FILE=""
CURRENT_TEST_NAME=""

# 设置测试上下文
set_test_context() {
    CURRENT_TEST_FILE="$1"
    CURRENT_TEST_NAME="$2"
}

# 输出带颜色的消息
print_colored() {
    local color="$1"
    local message="$2"
    if [[ "$VERBOSE" == "true" ]]; then
        echo -e "${color}${message}${COLOR_RESET}"
    fi
}

# 记录断言结果
record_assertion() {
    local result="$1"
    local message="$2"
    local details="${3:-}"
    
    ((ASSERTION_TOTAL++))
    
    if [[ "$result" == "PASS" ]]; then
        ((ASSERTION_PASSED++))
        print_colored "$COLOR_GREEN" "  ✓ $message"
        [[ "$DEBUG" == "true" ]] && echo "    Details: $details"
    else
        ((ASSERTION_FAILED++))
        print_colored "$COLOR_RED" "  ✗ $message"
        echo "    Expected: $details" >&2
        
        if [[ "$STOP_ON_FAIL" == "true" ]]; then
            echo "Stopping on first failure as requested." >&2
            exit 1
        fi
    fi
}

# 基础断言：相等
assert_equals() {
    local expected="$1"
    local actual="$2"
    local message="${3:-Values should be equal}"
    
    if [[ "$expected" == "$actual" ]]; then
        record_assertion "PASS" "$message" "Expected: '$expected', Actual: '$actual'"
    else
        record_assertion "FAIL" "$message" "Expected: '$expected', but got: '$actual'"
        return 1
    fi
}

# 基础断言：不相等
assert_not_equals() {
    local not_expected="$1"
    local actual="$2"
    local message="${3:-Values should not be equal}"
    
    if [[ "$not_expected" != "$actual" ]]; then
        record_assertion "PASS" "$message" "Not expected: '$not_expected', Actual: '$actual'"
    else
        record_assertion "FAIL" "$message" "Expected values to be different, but both are: '$actual'"
        return 1
    fi
}

# 基础断言：包含
assert_contains() {
    local haystack="$1"
    local needle="$2"
    local message="${3:-Text should contain substring}"
    
    if [[ "$haystack" == *"$needle"* ]]; then
        record_assertion "PASS" "$message" "Found '$needle' in text"
    else
        record_assertion "FAIL" "$message" "Text '$haystack' does not contain '$needle'"
        return 1
    fi
}

# 基础断言：不包含
assert_not_contains() {
    local haystack="$1"
    local needle="$2"
    local message="${3:-Text should not contain substring}"
    
    if [[ "$haystack" != *"$needle"* ]]; then
        record_assertion "PASS" "$message" "Text does not contain '$needle'"
    else
        record_assertion "FAIL" "$message" "Text '$haystack' unexpectedly contains '$needle'"
        return 1
    fi
}

# 基础断言：匹配正则表达式
assert_matches() {
    local text="$1"
    local pattern="$2"
    local message="${3:-Text should match pattern}"
    
    if [[ "$text" =~ $pattern ]]; then
        record_assertion "PASS" "$message" "Text matches pattern '$pattern'"
    else
        record_assertion "FAIL" "$message" "Text '$text' does not match pattern '$pattern'"
        return 1
    fi
}

# 基础断言：真值
assert_true() {
    local condition="$1"
    local message="${2:-Condition should be true}"
    
    if [[ "$condition" == "true" || "$condition" == "0" ]]; then
        record_assertion "PASS" "$message" "Condition is true"
    else
        record_assertion "FAIL" "$message" "Expected true, but got: '$condition'"
        return 1
    fi
}

# 基础断言：假值
assert_false() {
    local condition="$1"
    local message="${2:-Condition should be false}"
    
    if [[ "$condition" == "false" || "$condition" != "0" && "$condition" != "true" ]]; then
        record_assertion "PASS" "$message" "Condition is false"
    else
        record_assertion "FAIL" "$message" "Expected false, but got: '$condition'"
        return 1
    fi
}

# 数据库专用断言：查询结果行数
assert_query_row_count() {
    local query="$1"
    local expected_count="$2"
    local message="${3:-Query should return expected number of rows}"
    
    local result=$(execute_query "$query")
    local actual_count=$(echo "$result" | grep -c "^|" | tail -1)
    # 减去表头行
    actual_count=$((actual_count - 1))
    
    if [[ "$actual_count" -eq "$expected_count" ]]; then
        record_assertion "PASS" "$message" "Expected: $expected_count rows, Actual: $actual_count rows"
    else
        record_assertion "FAIL" "$message" "Expected: $expected_count rows, but got: $actual_count rows"
        [[ "$DEBUG" == "true" ]] && echo "Query result: $result"
        return 1
    fi
}

# 数据库专用断言：查询包含特定值
assert_query_contains_value() {
    local query="$1"
    local expected_value="$2"
    local message="${3:-Query result should contain expected value}"
    
    local result=$(execute_query "$query")
    
    if echo "$result" | grep -q "$expected_value"; then
        record_assertion "PASS" "$message" "Found '$expected_value' in query result"
    else
        record_assertion "FAIL" "$message" "Query result does not contain '$expected_value'"
        [[ "$DEBUG" == "true" ]] && echo "Query result: $result"
        return 1
    fi
}

# 数据库专用断言：查询无错误
assert_query_succeeds() {
    local query="$1"
    local message="${2:-Query should execute successfully}"
    
    local result=$(execute_query "$query" 2>&1)
    
    if echo "$result" | grep -q "Error:"; then
        record_assertion "FAIL" "$message" "Query failed with: $result"
        return 1
    else
        record_assertion "PASS" "$message" "Query executed successfully"
    fi
}

# 数据库专用断言：查询有错误
assert_query_fails() {
    local query="$1"
    local expected_error="${2:-Error}"
    local message="${3:-Query should fail with expected error}"
    
    local result=$(execute_query "$query" 2>&1)
    
    if echo "$result" | grep -q "Error:"; then
        if [[ -z "$expected_error" ]] || echo "$result" | grep -q "$expected_error"; then
            record_assertion "PASS" "$message" "Query failed as expected"
        else
            record_assertion "FAIL" "$message" "Query failed but with wrong error. Got: $result"
            return 1
        fi
    else
        record_assertion "FAIL" "$message" "Query should have failed but succeeded. Result: $result"
        return 1
    fi
}

# 获取断言统计信息
get_assertion_stats() {
    cat << EOF
{
  "total": $ASSERTION_TOTAL,
  "passed": $ASSERTION_PASSED,
  "failed": $ASSERTION_FAILED,
  "success_rate": $(bc -l <<< "scale=2; $ASSERTION_PASSED * 100 / $ASSERTION_TOTAL" 2>/dev/null || echo "0")
}
EOF
}

# 重置断言统计
reset_assertion_stats() {
    ASSERTION_TOTAL=0
    ASSERTION_PASSED=0
    ASSERTION_FAILED=0
}

# 打印断言摘要
print_assertion_summary() {
    echo ""
    echo "=== ASSERTION SUMMARY ==="
    echo "Total Assertions: $ASSERTION_TOTAL"
    echo "Passed: $ASSERTION_PASSED"
    echo "Failed: $ASSERTION_FAILED"
    if [[ $ASSERTION_TOTAL -gt 0 ]]; then
        local success_rate=$(bc -l <<< "scale=1; $ASSERTION_PASSED * 100 / $ASSERTION_TOTAL" 2>/dev/null || echo "0")
        echo "Success Rate: ${success_rate}%"
    fi
}