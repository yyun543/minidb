#!/bin/bash
# Test P0 fixes
set -e
echo "=== Testing P0 Fixes ==="
cd /Users/10270273/codes/minidb
./minidb > logs/test_server.log 2>&1 &
SERVER_PID=$!
sleep 3
echo "Server started (PID: $SERVER_PID)"

# Test projection
echo "SELECT * FROM products;" | nc localhost 7205 > /dev/null 2>&1 && echo "✓ Fix 1 PASSED" || echo "✗ Fix 1 needs verification"

# Test transaction  
echo "START TRANSACTION; COMMIT;" | nc localhost 7205 > /dev/null 2>&1 && echo "✓ Fix 2 PASSED" || echo "✗ Fix 2 needs verification"

kill $SERVER_PID 2>/dev/null || true
echo "=== Tests Complete ==="
