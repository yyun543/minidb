#!/bin/bash

# Lakehouse Features Test Runner
# This script runs all tests for Lakehouse v2.0 features

set -e

echo "====================================="
echo "MiniDB Lakehouse v2.0 Test Suite"
echo "====================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counters
PASSED=0
FAILED=0
TOTAL=0

# Function to run a test
run_test() {
    local test_name=$1
    local test_file=$2
    local test_func=$3

    echo -e "${YELLOW}Running: $test_name${NC}"
    TOTAL=$((TOTAL + 1))

    if go test -v "$test_file" -run "$test_func" -timeout 3m 2>&1 | tee /tmp/minidb_test_output.log | grep -q "PASS"; then
        echo -e "${GREEN}✓ PASSED: $test_name${NC}"
        PASSED=$((PASSED + 1))
    else
        echo -e "${RED}✗ FAILED: $test_name${NC}"
        FAILED=$((FAILED + 1))
        # Show last 20 lines of output on failure
        echo "Last 20 lines of output:"
        tail -20 /tmp/minidb_test_output.log
    fi
    echo ""
}

# 1. Delta Lake ACID Transaction Tests
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "1. Delta Lake ACID Properties"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
run_test "ACID Atomicity" "./delta_acid_test.go" "TestDeltaLakeACID/Atomicity"
run_test "ACID Consistency" "./delta_acid_test.go" "TestDeltaLakeACID/Consistency"
run_test "ACID Isolation" "./delta_acid_test.go" "TestDeltaLakeACID/Isolation"
run_test "ACID Durability" "./delta_acid_test.go" "TestDeltaLakeACID/Durability"
run_test "Version Control" "./delta_acid_test.go" "TestDeltaLakeACID/VersionControl"
run_test "Snapshot Isolation" "./delta_acid_test.go" "TestDeltaLogSnapshot"

# 2. Time Travel Tests
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "2. Time Travel Queries"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
run_test "Version-Based Time Travel" "./time_travel_test.go" "TestTimeTravelQueries/VersionBasedTimeTravel"
run_test "Snapshot Isolation" "./time_travel_test.go" "TestTimeTravelQueries/SnapshotIsolation"
run_test "Delta Log Version Tracking" "./time_travel_test.go" "TestTimeTravelQueries/DeltaLogVersionTracking"
run_test "Snapshot Retrieval" "./time_travel_test.go" "TestDeltaLogSnapshotRetrieval"
run_test "File Tracking (ADD/REMOVE)" "./time_travel_test.go" "TestDeltaLogFileTracking"

# 3. Predicate Pushdown Tests
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "3. Predicate Pushdown Optimization"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
run_test "Integer Predicate Pushdown" "./predicate_pushdown_test.go" "TestPredicatePushdown/IntegerPredicatePushdown"
run_test "String Predicate Pushdown" "./predicate_pushdown_test.go" "TestPredicatePushdown/StringPredicatePushdown"
run_test "Float Predicate Pushdown" "./predicate_pushdown_test.go" "TestPredicatePushdown/FloatPredicatePushdown"
run_test "Complex Predicates Pushdown" "./predicate_pushdown_test.go" "TestPredicatePushdown/ComplexPredicatesPushdown"
run_test "Data Skipping with Statistics" "./predicate_pushdown_test.go" "TestDataSkippingWithStatistics"

# 4. Parquet Statistics Tests
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "4. Parquet Statistics Collection"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
run_test "Int64 Statistics" "./parquet_statistics_test.go" "TestParquetStatisticsCollection/Int64Statistics"
run_test "Float64 Statistics" "./parquet_statistics_test.go" "TestParquetStatisticsCollection/Float64Statistics"
run_test "String Statistics" "./parquet_statistics_test.go" "TestParquetStatisticsCollection/StringStatistics"
run_test "Boolean Statistics" "./parquet_statistics_test.go" "TestParquetStatisticsCollection/BooleanStatistics"
run_test "Mixed Type Statistics" "./parquet_statistics_test.go" "TestParquetStatisticsCollection/MixedTypeStatistics"
run_test "Integer Types Statistics" "./parquet_statistics_test.go" "TestParquetStatisticsCollection/Int32Int16Int8Statistics"
run_test "Float32 Statistics" "./parquet_statistics_test.go" "TestParquetStatisticsCollection/Float32Statistics"
run_test "Statistics Roundtrip" "./parquet_statistics_test.go" "TestStatisticsRoundtrip"

# 5. Arrow IPC Serialization Tests
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "5. Arrow IPC Serialization"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
run_test "Basic Schema Roundtrip" "./arrow_ipc_test.go" "TestArrowIPCSerialization/BasicSchemaRoundtrip"
run_test "Complex Schema Roundtrip" "./arrow_ipc_test.go" "TestArrowIPCSerialization/ComplexSchemaRoundtrip"
run_test "Schema with Metadata" "./arrow_ipc_test.go" "TestArrowIPCSerialization/SchemaWithMetadata"
run_test "Timestamp Types" "./arrow_ipc_test.go" "TestArrowIPCSerialization/TimestampTypes"
run_test "Date Types" "./arrow_ipc_test.go" "TestArrowIPCSerialization/DateTypes"
run_test "Multiple Schema Versions" "./arrow_ipc_test.go" "TestArrowIPCSerialization/MultipleSchemaVersions"
run_test "Arrow IPC Performance" "./arrow_ipc_test.go" "TestArrowIPCPerformance"

# Summary
echo ""
echo "====================================="
echo "Test Summary"
echo "====================================="
echo -e "Total:  $TOTAL"
echo -e "${GREEN}Passed: $PASSED${NC}"
if [ $FAILED -gt 0 ]; then
    echo -e "${RED}Failed: $FAILED${NC}"
else
    echo -e "Failed: $FAILED"
fi
echo "====================================="
echo ""

# Calculate success rate
if [ $TOTAL -gt 0 ]; then
    SUCCESS_RATE=$(awk "BEGIN {printf \"%.1f\", ($PASSED/$TOTAL)*100}")
    echo "Success Rate: $SUCCESS_RATE%"
fi

# Exit with error if any tests failed
if [ $FAILED -gt 0 ]; then
    exit 1
fi

echo -e "${GREEN}All Lakehouse v2.0 tests passed!${NC}"
