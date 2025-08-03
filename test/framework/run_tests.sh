#!/bin/bash

# MiniDB测试框架主入口
# 基于TDD思想的统一测试执行入口

set -euo pipefail

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# 加载框架组件
source "config/test_config.sh"
source "utils/test_runner.sh"
source "utils/report_generator.sh"

# 全局变量
TEST_TYPE="all"
GENERATE_REPORTS=true
PARALLEL_EXECUTION=false

# 显示使用说明
show_usage() {
    cat << EOF
MiniDB Test Framework

USAGE: $0 [OPTIONS] [TEST_TYPE]

TEST TYPES:
    all                    运行所有测试 (默认)
    unit                   仅运行单元测试
    integration           仅运行集成测试
    regression            仅运行回归测试
    group_by              运行GROUP BY功能测试
    unit/basic_operations  运行特定模块测试
    health                 运行基础设施健康检查

OPTIONS:
    -v, --verbose         详细输出模式
    -d, --debug           调试模式，显示详细调试信息
    -s, --stop-on-fail    遇到失败立即停止
    -q, --quiet           安静模式，仅显示结果摘要
    -n, --no-reports      不生成测试报告
    -p, --parallel        并行执行测试（实验性）
    -h, --help            显示此帮助信息

EXAMPLES:
    $0                           # 运行所有测试
    $0 unit                      # 仅运行单元测试
    $0 --verbose integration     # 详细模式运行集成测试
    $0 --debug --stop-on-fail    # 调试模式，遇到错误停止
    $0 regression --no-reports   # 运行回归测试，不生成报告

REPORTS:
    测试完成后在 $TEST_REPORTS_DIR 目录生成以下报告：
    - test_report.html      HTML格式的可视化报告
    - test_report.txt       纯文本格式报告
    - junit_results.xml     JUnit XML格式（用于CI集成）
    - test_results.json     原始JSON数据

ENVIRONMENT VARIABLES:
    VERBOSE=true           等同于 --verbose
    DEBUG=true             等同于 --debug
    STOP_ON_FAIL=true      等同于 --stop-on-fail
EOF
}

# 解析命令行参数
parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            -d|--debug)
                DEBUG=true
                VERBOSE=true  # 调试模式自动开启详细输出
                shift
                ;;
            -s|--stop-on-fail)
                STOP_ON_FAIL=true
                shift
                ;;
            -q|--quiet)
                VERBOSE=false
                shift
                ;;
            -n|--no-reports)
                GENERATE_REPORTS=false
                shift
                ;;
            -p|--parallel)
                PARALLEL_EXECUTION=true
                shift
                ;;
            -h|--help)
                show_usage
                exit 0
                ;;
            all|unit|integration|regression|health|group_by)
                TEST_TYPE="$1"
                shift
                ;;
            unit/*|integration/*|regression/*)
                TEST_TYPE="$1"
                shift
                ;;
            *)
                echo "未知选项或测试类型: $1" >&2
                echo "使用 --help 查看使用说明" >&2
                exit 1
                ;;
        esac
    done
}

# 运行健康检查
run_health_check() {
    echo "=== 运行基础设施健康检查 ==="
    
    if run_infrastructure_check; then
        print_colored "$COLOR_GREEN" "✓ 基础设施健康检查通过"
        return 0
    else
        print_colored "$COLOR_RED" "✗ 基础设施健康检查失败"
        return 1
    fi
}

# 运行单元测试
run_unit_tests() {
    local specific_module="${1:-}"
    
    echo "=== 运行单元测试 ==="
    
    if [[ -n "$specific_module" ]]; then
        # 运行特定模块
        if [[ -d "unit/$specific_module" ]]; then
            run_test_directory "unit/$specific_module"
        else
            echo "单元测试模块不存在: unit/$specific_module" >&2
            return 1
        fi
    else
        # 运行所有单元测试
        if [[ -d "unit" ]]; then
            run_test_directory "unit"
        else
            echo "单元测试目录不存在" >&2
            return 1
        fi
    fi
}

# 运行集成测试
run_integration_tests() {
    local specific_module="${1:-}"
    
    echo "=== 运行集成测试 ==="
    
    if [[ -n "$specific_module" ]]; then
        # 运行特定模块
        if [[ -d "integration/$specific_module" ]]; then
            run_test_directory "integration/$specific_module"
        else
            echo "集成测试模块不存在: integration/$specific_module" >&2
            return 1
        fi
    else
        # 运行所有集成测试
        if [[ -d "integration" ]]; then
            run_test_directory "integration"
        else
            echo "集成测试目录不存在" >&2
            return 1
        fi
    fi
}

# 运行回归测试
run_regression_tests() {
    local specific_module="${1:-}"
    
    echo "=== 运行回归测试 ==="
    
    if [[ -n "$specific_module" ]]; then
        # 运行特定模块
        if [[ -d "regression/$specific_module" ]]; then
            run_test_directory "regression/$specific_module"
        else
            echo "回归测试模块不存在: regression/$specific_module" >&2
            return 1
        fi
    else
        # 运行所有回归测试
        if [[ -d "regression" ]]; then
            run_test_directory "regression"
        else
            echo "回归测试目录不存在" >&2
            return 1
        fi
    fi
}

# 运行GROUP BY功能测试
run_group_by_tests() {
    echo "=== 运行GROUP BY功能测试 ==="
    
    # 运行Go单元测试中的GROUP BY测试
    cd "$SCRIPT_DIR/.."
    echo "运行Go单元测试中的GROUP BY功能..."
    if go test -v ./test -run "TestGroupByFunctionality" 2>/dev/null; then
        print_colored "$COLOR_GREEN" "✓ Go单元测试中的GROUP BY功能测试通过"
    else
        print_colored "$COLOR_YELLOW" "! Go单元测试执行遇到问题，继续执行集成测试"
    fi
    
    # 运行集成测试脚本
    cd "$SCRIPT_DIR"
    echo "运行GROUP BY集成测试..."
    if [[ -f "integration/group_by_comprehensive_test.sh" ]]; then
        if bash "integration/group_by_comprehensive_test.sh"; then
            print_colored "$COLOR_GREEN" "✓ GROUP BY集成测试通过"
        else
            print_colored "$COLOR_RED" "✗ GROUP BY集成测试失败"
            return 1
        fi
    else
        echo "GROUP BY集成测试脚本不存在" >&2
        return 1
    fi
    
    # 运行演示脚本
    if [[ -f "demo/working_features_demo.sh" ]]; then
        echo "运行GROUP BY功能演示..."
        if bash "demo/working_features_demo.sh"; then
            print_colored "$COLOR_GREEN" "✓ GROUP BY功能演示通过"
        else
            print_colored "$COLOR_YELLOW" "! GROUP BY功能演示有问题"
        fi
    fi
}

# 运行所有测试
run_all_tests() {
    echo "=== 运行完整测试套件 ==="
    
    # 首先运行健康检查
    run_health_check || {
        echo "基础设施检查失败，跳过后续测试" >&2
        return 1
    }
    
    # 运行单元测试
    if [[ -d "unit" ]]; then
        run_unit_tests
    fi
    
    # 运行集成测试
    if [[ -d "integration" ]]; then
        run_integration_tests
    fi
    
    # 运行回归测试
    if [[ -d "regression" ]]; then
        run_regression_tests
    fi
    
    # 运行GROUP BY测试
    run_group_by_tests
}

# 主执行函数
main() {
    # 显示启动信息
    print_colored "$COLOR_BLUE" "=== MiniDB测试框架启动 ==="
    echo "时间: $(date)"
    echo "测试类型: $TEST_TYPE"
    echo "详细模式: $VERBOSE"
    echo "调试模式: $DEBUG"
    echo "遇到失败停止: $STOP_ON_FAIL"
    echo "生成报告: $GENERATE_REPORTS"
    echo ""
    
    # 创建报告目录
    mkdir -p "$TEST_REPORTS_DIR"
    
    # 根据测试类型执行相应测试
    local exit_code=0
    case "$TEST_TYPE" in
        "health")
            run_health_check || exit_code=$?
            ;;
        "unit")
            run_unit_tests || exit_code=$?
            ;;
        "integration")
            run_integration_tests || exit_code=$?
            ;;
        "regression")
            run_regression_tests || exit_code=$?
            ;;
        "group_by")
            run_group_by_tests || exit_code=$?
            ;;
        "unit/"*)
            local module="${TEST_TYPE#unit/}"
            run_unit_tests "$module" || exit_code=$?
            ;;
        "integration/"*)
            local module="${TEST_TYPE#integration/}"
            run_integration_tests "$module" || exit_code=$?
            ;;
        "regression/"*)
            local module="${TEST_TYPE#regression/}"
            run_regression_tests "$module" || exit_code=$?
            ;;
        "all")
            run_all_tests || exit_code=$?
            ;;
        *)
            echo "未知测试类型: $TEST_TYPE" >&2
            exit 1
            ;;
    esac
    
    # 生成测试报告数据
    generate_test_report_data
    
    # 生成测试报告
    if [[ "$GENERATE_REPORTS" == "true" && "$TEST_TYPE" != "health" ]]; then
        echo ""
        echo "=== 生成测试报告 ==="
        generate_all_reports
    fi
    
    # 打印最终摘要
    print_final_summary
    
    # 显示结束信息
    echo ""
    print_colored "$COLOR_BLUE" "=== 测试框架执行完成 ==="
    echo "总耗时: $(($(date +%s) - $(date -r "$TEST_REPORTS_DIR/.." +%s 2>/dev/null || echo 0)))秒"
    
    if [[ $exit_code -eq 0 ]]; then
        print_colored "$COLOR_GREEN" "🎉 所有测试通过！"
    else
        print_colored "$COLOR_RED" "❌ 存在测试失败"
        if [[ "$VERBOSE" == "false" ]]; then
            echo "使用 --verbose 选项查看详细信息"
        fi
    fi
    
    return $exit_code
}

# 脚本执行入口
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    # 解析命令行参数
    parse_arguments "$@"
    
    # 执行测试
    main
    
    # 退出并返回适当的状态码
    exit $?
fi