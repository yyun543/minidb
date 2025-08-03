#!/bin/bash

# MiniDB测试框架全局配置
# 基于KISS原则的简单配置管理

# 服务器配置
export DB_HOST="localhost"
export DB_PORT="7205"
export DB_TIMEOUT=30
export DB_START_WAIT=2

# 测试配置
export TEST_ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
export TEST_REPORTS_DIR="${TEST_ROOT_DIR}/reports"
export TEST_DATA_DIR="${TEST_ROOT_DIR}/data"
export TEST_TEMP_DIR="/tmp/minidb_test_$$"

# 数据库二进制文件路径
export DB_BINARY="${TEST_ROOT_DIR}/../../minidb"
export DB_BUILD_DIR="${TEST_ROOT_DIR}/../../"

# 测试结果状态
export TEST_PASSED=0
export TEST_FAILED=1
export TEST_SKIPPED=2

# 颜色配置（用于输出）
export COLOR_RED='\033[0;31m'
export COLOR_GREEN='\033[0;32m'
export COLOR_YELLOW='\033[1;33m'
export COLOR_BLUE='\033[0;34m'
export COLOR_RESET='\033[0m'

# 测试执行选项
export VERBOSE=${VERBOSE:-false}
export DEBUG=${DEBUG:-false}
export STOP_ON_FAIL=${STOP_ON_FAIL:-false}

# 创建必要的目录
mkdir -p "${TEST_REPORTS_DIR}"
mkdir -p "${TEST_DATA_DIR}"
mkdir -p "${TEST_TEMP_DIR}"

# 清理函数
cleanup_test_env() {
    if [[ -d "${TEST_TEMP_DIR}" ]]; then
        rm -rf "${TEST_TEMP_DIR}"
    fi
    
    # 终止可能残留的数据库进程
    pkill -f minidb 2>/dev/null || true
    wait 2>/dev/null || true
}

# 注册退出时清理
trap cleanup_test_env EXIT