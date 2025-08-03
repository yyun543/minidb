#!/bin/bash

# 测试报告生成器
# 遵循KISS原则，生成清晰直观的测试报告

source "$(dirname "${BASH_SOURCE[0]}")/../config/test_config.sh"

# 生成HTML报告
generate_html_report() {
    local json_file="$TEST_REPORTS_DIR/test_results.json"
    local html_file="$TEST_REPORTS_DIR/test_report.html"
    
    if [[ ! -f "$json_file" ]]; then
        echo "No test results found. Run tests first." >&2
        return 1
    fi
    
    # 读取JSON数据（简化处理）
    local timestamp=$(grep '"timestamp"' "$json_file" | cut -d'"' -f4)
    local total_suites=$(grep '"total_suites"' "$json_file" | cut -d':' -f2 | tr -d ' ,')
    local passed=$(grep '"passed"' "$json_file" | grep -v success_rate | cut -d':' -f2 | tr -d ' ,')
    local failed=$(grep '"failed"' "$json_file" | cut -d':' -f2 | tr -d ' ,')
    local skipped=$(grep '"skipped"' "$json_file" | cut -d':' -f2 | tr -d ' ,')
    local success_rate=$(grep '"success_rate"' "$json_file" | cut -d':' -f2 | tr -d ' ')
    
    cat > "$html_file" << EOF
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>MiniDB Test Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .header { text-align: center; border-bottom: 2px solid #333; padding-bottom: 20px; margin-bottom: 30px; }
        .summary { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 20px; margin-bottom: 30px; }
        .metric { background: #f8f9fa; padding: 15px; border-radius: 6px; text-align: center; border-left: 4px solid #007bff; }
        .metric.passed { border-left-color: #28a745; }
        .metric.failed { border-left-color: #dc3545; }
        .metric.skipped { border-left-color: #ffc107; }
        .metric h3 { margin: 0 0 10px 0; color: #333; }
        .metric .value { font-size: 2em; font-weight: bold; }
        .test-suites { margin-top: 30px; }
        .test-suite { margin: 10px 0; padding: 15px; border-radius: 6px; border-left: 4px solid #ccc; }
        .test-suite.passed { background: #d4edda; border-left-color: #28a745; }
        .test-suite.failed { background: #f8d7da; border-left-color: #dc3545; }
        .test-suite.skipped { background: #fff3cd; border-left-color: #ffc107; }
        .test-suite h4 { margin: 0 0 10px 0; }
        .test-details { color: #666; font-size: 0.9em; }
        .progress-bar { width: 100%; height: 30px; background: #e9ecef; border-radius: 15px; overflow: hidden; margin: 20px 0; }
        .progress-fill { height: 100%; background: linear-gradient(90deg, #28a745 0%, #28a745 $success_rate%, #dc3545 $success_rate%, #dc3545 100%); }
        .timestamp { text-align: center; color: #666; margin-top: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>MiniDB Test Report</h1>
            <p>Comprehensive testing results for the MiniDB database system</p>
        </div>
        
        <div class="summary">
            <div class="metric">
                <h3>Total Suites</h3>
                <div class="value">$total_suites</div>
            </div>
            <div class="metric passed">
                <h3>Passed</h3>
                <div class="value">$passed</div>
            </div>
            <div class="metric failed">
                <h3>Failed</h3>
                <div class="value">$failed</div>
            </div>
            <div class="metric skipped">
                <h3>Skipped</h3>
                <div class="value">$skipped</div>
            </div>
        </div>
        
        <div class="progress-bar">
            <div class="progress-fill"></div>
        </div>
        <p style="text-align: center;"><strong>Success Rate: $success_rate%</strong></p>
        
        <div class="test-suites">
            <h2>Test Suite Details</h2>
EOF
    
    # 解析测试套件信息（简化处理）
    local in_suites=false
    while IFS= read -r line; do
        if [[ "$line" == *'"test_suites"'* ]]; then
            in_suites=true
            continue
        fi
        
        if [[ "$in_suites" == "true" && "$line" == *'"name"'* ]]; then
            local name=$(echo "$line" | cut -d'"' -f4)
            local status=""
            local duration=""
            local details=""
            
            # 读取接下来的几行获取详细信息
            read -r status_line
            read -r duration_line
            read -r details_line
            
            status=$(echo "$status_line" | cut -d'"' -f4)
            duration=$(echo "$duration_line" | cut -d'"' -f4)
            details=$(echo "$details_line" | cut -d'"' -f4)
            
            local status_class=$(echo "$status" | tr '[:upper:]' '[:lower:]')
            
            cat >> "$html_file" << EOF
            <div class="test-suite $status_class">
                <h4>$name</h4>
                <div class="test-details">
                    <strong>Status:</strong> $status<br>
                    <strong>Duration:</strong> ${duration}s<br>
                    <strong>Details:</strong> $details
                </div>
            </div>
EOF
        fi
    done < "$json_file"
    
    cat >> "$html_file" << EOF
        </div>
        
        <div class="timestamp">
            <p>Report generated on: $timestamp</p>
        </div>
    </div>
</body>
</html>
EOF
    
    echo "HTML report generated: $html_file"
}

# 生成纯文本报告
generate_text_report() {
    local json_file="$TEST_REPORTS_DIR/test_results.json"
    local text_file="$TEST_REPORTS_DIR/test_report.txt"
    
    if [[ ! -f "$json_file" ]]; then
        echo "No test results found. Run tests first." >&2
        return 1
    fi
    
    cat > "$text_file" << 'EOF'
================================================================================
                             MINIDB TEST REPORT
================================================================================

EOF
    
    # 从JSON文件提取基本信息
    local timestamp=$(grep '"timestamp"' "$json_file" | cut -d'"' -f4)
    local total_suites=$(grep '"total_suites"' "$json_file" | cut -d':' -f2 | tr -d ' ,')
    local passed=$(grep '"passed"' "$json_file" | grep -v success_rate | cut -d':' -f2 | tr -d ' ,')
    local failed=$(grep '"failed"' "$json_file" | cut -d':' -f2 | tr -d ' ,')
    local skipped=$(grep '"skipped"' "$json_file" | cut -d':' -f2 | tr -d ' ,')
    local success_rate=$(grep '"success_rate"' "$json_file" | cut -d':' -f2 | tr -d ' ')
    
    cat >> "$text_file" << EOF
SUMMARY:
--------
Generated: $timestamp
Total Test Suites: $total_suites
Passed: $passed
Failed: $failed
Skipped: $skipped
Success Rate: $success_rate%

EOF
    
    # 如果有失败的测试，列出详细信息
    if [[ "$failed" != "0" ]]; then
        cat >> "$text_file" << EOF
FAILED TESTS:
-------------
EOF
        
        # 简化的JSON解析来获取失败的测试
        grep -A 3 -B 1 '"status": "FAILED"' "$json_file" | while read -r line; do
            if [[ "$line" == *'"name"'* ]]; then
                local name=$(echo "$line" | cut -d'"' -f4)
                echo "- $name" >> "$text_file"
            fi
        done
        
        echo "" >> "$text_file"
    fi
    
    cat >> "$text_file" << EOF
DETAILED RESULTS:
----------------
EOF
    
    # 详细的测试结果（简化处理）
    local in_suites=false
    while IFS= read -r line; do
        if [[ "$line" == *'"name"'* ]]; then
            local name=$(echo "$line" | cut -d'"' -f4)
            read -r status_line
            read -r duration_line
            read -r details_line
            
            local status=$(echo "$status_line" | cut -d'"' -f4)
            local duration=$(echo "$duration_line" | cut -d'"' -f4)
            local details=$(echo "$details_line" | cut -d'"' -f4)
            
            printf "%-30s %-10s %8ss  %s\n" "$name" "$status" "$duration" "$details" >> "$text_file"
        fi
    done < <(grep -A 3 '"name"' "$json_file")
    
    cat >> "$text_file" << EOF

================================================================================
EOF
    
    echo "Text report generated: $text_file"
}

# 生成JUnit XML格式报告（用于CI集成）
generate_junit_report() {
    local json_file="$TEST_REPORTS_DIR/test_results.json"
    local junit_file="$TEST_REPORTS_DIR/junit_results.xml"
    
    if [[ ! -f "$json_file" ]]; then
        echo "No test results found. Run tests first." >&2
        return 1
    fi
    
    local timestamp=$(grep '"timestamp"' "$json_file" | cut -d'"' -f4)
    local total_suites=$(grep '"total_suites"' "$json_file" | cut -d':' -f2 | tr -d ' ,')
    local failed=$(grep '"failed"' "$json_file" | grep -v success_rate | cut -d':' -f2 | tr -d ' ,')
    
    cat > "$junit_file" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<testsuite name="MiniDB Test Suite" 
           tests="$total_suites" 
           failures="$failed" 
           errors="0" 
           time="0" 
           timestamp="$timestamp">
EOF
    
    # 添加测试用例
    while IFS= read -r line; do
        if [[ "$line" == *'"name"'* ]]; then
            local name=$(echo "$line" | cut -d'"' -f4)
            read -r status_line
            read -r duration_line
            read -r details_line
            
            local status=$(echo "$status_line" | cut -d'"' -f4)
            local duration=$(echo "$duration_line" | cut -d'"' -f4)
            local details=$(echo "$details_line" | cut -d'"' -f4)
            
            echo "  <testcase classname=\"MiniDB\" name=\"$name\" time=\"$duration\">" >> "$junit_file"
            
            if [[ "$status" == "FAILED" ]]; then
                echo "    <failure message=\"Test failed\">$details</failure>" >> "$junit_file"
            elif [[ "$status" == "SKIPPED" ]]; then
                echo "    <skipped message=\"Test skipped\">$details</skipped>" >> "$junit_file"
            fi
            
            echo "  </testcase>" >> "$junit_file"
        fi
    done < <(grep -A 3 '"name"' "$json_file")
    
    echo "</testsuite>" >> "$junit_file"
    
    echo "JUnit XML report generated: $junit_file"
}

# 生成所有格式的报告
generate_all_reports() {
    echo "Generating test reports..."
    
    generate_html_report
    generate_text_report
    generate_junit_report
    
    echo ""
    echo "All reports generated in: $TEST_REPORTS_DIR"
    echo "- HTML Report: $TEST_REPORTS_DIR/test_report.html"
    echo "- Text Report: $TEST_REPORTS_DIR/test_report.txt"
    echo "- JUnit XML: $TEST_REPORTS_DIR/junit_results.xml"
    echo "- JSON Data: $TEST_REPORTS_DIR/test_results.json"
}