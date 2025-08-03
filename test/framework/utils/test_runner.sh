#!/bin/bash

# 测试执行器 - 统一的测试执行框架
# 基于TDD思想，提供结构化的测试执行环境

source "$(dirname "${BASH_SOURCE[0]}")/../config/test_config.sh"
source "$(dirname "${BASH_SOURCE[0]}")/../config/test_data.sh"
source "$(dirname "${BASH_SOURCE[0]}")/assertion.sh"
source "$(dirname "${BASH_SOURCE[0]}")/db_helper.sh"

# 测试执行统计
TEST_SUITE_TOTAL=0
TEST_SUITE_PASSED=0
TEST_SUITE_FAILED=0
TEST_SUITE_SKIPPED=0

# 当前运行的测试套件信息
CURRENT_SUITE=""
CURRENT_TEST_START_TIME=""

# 测试结果存储（使用简单数组代替关联数组以兼容老版本bash）
TEST_RESULTS=()
TEST_DURATIONS=()
TEST_DETAILS=()
TEST_SUITE_NAMES=()

# 开始测试套件
begin_test_suite() {
    local suite_name="$1"
    local description="${2:-}"
    
    CURRENT_SUITE="$suite_name"
    CURRENT_TEST_START_TIME=$(date +%s.%N)
    
    echo ""
    print_colored "$COLOR_BLUE" "=== Starting Test Suite: $suite_name ==="
    [[ -n "$description" ]] && echo "Description: $description"
    echo ""
    
    # 重置断言统计
    reset_assertion_stats
}

# 结束测试套件
end_test_suite() {
    local suite_name="$1"
    local end_time=$(date +%s.%N)
    local duration=$(echo "$end_time - $CURRENT_TEST_START_TIME" | bc -l)
    
    # 记录测试结果（使用索引数组）
    local suite_index=${#TEST_SUITE_NAMES[@]}
    TEST_SUITE_NAMES[$suite_index]="$suite_name"
    
    if [[ $ASSERTION_FAILED -eq 0 ]]; then
        TEST_RESULTS[$suite_index]="PASSED"
        ((TEST_SUITE_PASSED++))
        print_colored "$COLOR_GREEN" "✓ Test Suite '$suite_name' PASSED"
    else
        TEST_RESULTS[$suite_index]="FAILED"
        ((TEST_SUITE_FAILED++))
        print_colored "$COLOR_RED" "✗ Test Suite '$suite_name' FAILED"
    fi
    
    TEST_DURATIONS[$suite_index]=$(printf "%.3f" "$duration")
    TEST_DETAILS[$suite_index]="Assertions: $ASSERTION_TOTAL, Passed: $ASSERTION_PASSED, Failed: $ASSERTION_FAILED"
    
    ((TEST_SUITE_TOTAL++))
    
    echo "Duration: ${TEST_DURATIONS[$suite_index]}s"
    print_assertion_summary
    echo ""
}

# 跳过测试套件
skip_test_suite() {
    local suite_name="$1"
    local reason="${2:-No reason provided}"
    
    local suite_index=${#TEST_SUITE_NAMES[@]}
    TEST_SUITE_NAMES[$suite_index]="$suite_name"
    TEST_RESULTS[$suite_index]="SKIPPED"
    TEST_DETAILS[$suite_index]="Skipped: $reason"
    ((TEST_SUITE_SKIPPED++))
    ((TEST_SUITE_TOTAL++))
    
    print_colored "$COLOR_YELLOW" "⚠ Test Suite '$suite_name' SKIPPED: $reason"
}

# 运行单个测试脚本
run_test_script() {
    local test_script="$1"
    local suite_name=$(basename "$test_script" .sh)
    
    if [[ ! -f "$test_script" ]]; then
        skip_test_suite "$suite_name" "Test script not found: $test_script"
        return 1
    fi
    
    if [[ ! -x "$test_script" ]]; then
        chmod +x "$test_script"
    fi
    
    begin_test_suite "$suite_name"
    
    # 设置测试上下文
    set_test_context "$test_script" "$suite_name"
    
    # 执行测试脚本
    local test_output
    local test_exit_code
    
    if [[ "$DEBUG" == "true" ]]; then
        echo "Executing: $test_script"
        test_output=$("$test_script" 2>&1)
        test_exit_code=$?
    else
        test_output=$("$test_script" 2>&1)
        test_exit_code=$?
    fi
    
    # 检查测试执行结果
    if [[ $test_exit_code -eq 0 ]]; then
        [[ "$VERBOSE" == "true" ]] && echo "Test output: $test_output"
    else
        echo "Test script failed with exit code: $test_exit_code" >&2
        [[ "$DEBUG" == "true" ]] && echo "Test output: $test_output" >&2
        # 记录为断言失败
        record_assertion "FAIL" "Test script execution" "Script exited with code $test_exit_code"
    fi
    
    end_test_suite "$suite_name"
    
    return $test_exit_code
}

# 运行目录中的所有测试
run_test_directory() {
    local test_dir="$1"
    local pattern="${2:-test_*.sh}"
    
    if [[ ! -d "$test_dir" ]]; then
        echo "Test directory not found: $test_dir" >&2
        return 1
    fi
    
    local test_files=()
    while IFS= read -r -d '' file; do
        test_files+=("$file")
    done < <(find "$test_dir" -name "$pattern" -type f -print0 | sort -z)
    
    if [[ ${#test_files[@]} -eq 0 ]]; then
        echo "No test files found in $test_dir with pattern $pattern"
        return 1
    fi
    
    echo "Found ${#test_files[@]} test files in $test_dir"
    
    for test_file in "${test_files[@]}"; do
        run_test_script "$test_file"
        
        if [[ "$STOP_ON_FAIL" == "true" ]]; then
            # 检查最后一个测试是否失败
            local last_index=$((${#TEST_RESULTS[@]} - 1))
            if [[ $last_index -ge 0 && "${TEST_RESULTS[$last_index]}" == "FAILED" ]]; then
                echo "Stopping test execution due to failure and --stop-on-fail flag"
                break
            fi
        fi
    done
}

# 运行测试基础设施检查
run_infrastructure_check() {
    echo "Running infrastructure check..."
    
    # 检查必要文件
    local required_files=(
        "$DB_BINARY"
        "$TEST_ROOT_DIR/config/test_config.sh"
        "$TEST_ROOT_DIR/config/test_data.sh"
        "$TEST_ROOT_DIR/utils/assertion.sh"
        "$TEST_ROOT_DIR/utils/db_helper.sh"
    )
    
    for file in "${required_files[@]}"; do
        if [[ ! -f "$file" ]]; then
            echo "✗ Required file missing: $file" >&2
            return 1
        fi
    done
    
    echo "✓ All required files present"
    
    # 检查数据库
    start_database
    if database_health_check; then
        echo "✓ Database infrastructure check passed"
        return 0
    else
        echo "✗ Database infrastructure check failed" >&2
        return 1
    fi
}

# 生成测试报告数据
generate_test_report_data() {
    local report_file="$TEST_REPORTS_DIR/test_results.json"
    
    cat > "$report_file" << EOF
{
  "timestamp": "$(date -Iseconds)",
  "summary": {
    "total_suites": $TEST_SUITE_TOTAL,
    "passed": $TEST_SUITE_PASSED,
    "failed": $TEST_SUITE_FAILED,
    "skipped": $TEST_SUITE_SKIPPED,
    "success_rate": $(bc -l <<< "scale=1; $TEST_SUITE_PASSED * 100 / $TEST_SUITE_TOTAL" 2>/dev/null || echo "0")
  },
  "test_suites": [
EOF
    
    local first=true
    for ((i=0; i<${#TEST_RESULTS[@]}; i++)); do
        if [[ "$first" == "true" ]]; then
            first=false
        else
            echo "," >> "$report_file"
        fi
        
        cat >> "$report_file" << EOF
    {
      "name": "${TEST_SUITE_NAMES[$i]}",
      "status": "${TEST_RESULTS[$i]}",
      "duration": "${TEST_DURATIONS[$i]:-0}",
      "details": "${TEST_DETAILS[$i]:-}"
    }
EOF
    done
    
    cat >> "$report_file" << EOF
  ],
  "assertions": $(get_assertion_stats)
}
EOF
    
    echo "Test report data saved to: $report_file"
}

# 打印最终测试摘要
print_final_summary() {
    echo ""
    print_colored "$COLOR_BLUE" "=== FINAL TEST SUMMARY ==="
    echo "Test Suites Run: $TEST_SUITE_TOTAL"
    print_colored "$COLOR_GREEN" "Passed: $TEST_SUITE_PASSED"
    print_colored "$COLOR_RED" "Failed: $TEST_SUITE_FAILED"
    print_colored "$COLOR_YELLOW" "Skipped: $TEST_SUITE_SKIPPED"
    
    if [[ $TEST_SUITE_TOTAL -gt 0 ]]; then
        local success_rate=$(bc -l <<< "scale=1; $TEST_SUITE_PASSED * 100 / $TEST_SUITE_TOTAL" 2>/dev/null || echo "0")
        echo "Success Rate: ${success_rate}%"
    fi
    
    # 显示失败的测试
    if [[ $TEST_SUITE_FAILED -gt 0 ]]; then
        echo ""
        print_colored "$COLOR_RED" "Failed Test Suites:"
        for ((i=0; i<${#TEST_RESULTS[@]}; i++)); do
            if [[ "${TEST_RESULTS[$i]}" == "FAILED" ]]; then
                echo "  - ${TEST_SUITE_NAMES[$i]}: ${TEST_DETAILS[$i]}"
            fi
        done
    fi
    
    echo ""
    
    # 返回适当的退出代码
    if [[ $TEST_SUITE_FAILED -eq 0 ]]; then
        return 0
    else
        return 1
    fi
}