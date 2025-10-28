-- Manual Verification Script for Aggregate Function Fix
-- This script verifies that aggregate functions behave correctly on empty tables
-- per SQL:2023 standard requirements

-- Setup test database and table
CREATE DATABASE test_agg;
USE test_agg;

CREATE TABLE products (
    id INTEGER,
    name VARCHAR(100),
    price DOUBLE,
    quantity INTEGER,
    category VARCHAR
);

-- Test 1: Aggregate functions on initially empty table
-- Expected: All queries should return exactly 1 row

-- Test 1.1: COUNT(*) should return 0
SELECT COUNT(*) FROM products;
-- Expected output: 1 row with value 0

-- Test 1.2: COUNT(column) should return 0
SELECT COUNT(name) FROM products;
-- Expected output: 1 row with value 0

-- Test 1.3: SUM should return NULL
SELECT SUM(price) FROM products;
-- Expected output: 1 row with NULL

-- Test 1.4: AVG should return NULL
SELECT AVG(price) FROM products;
-- Expected output: 1 row with NULL

-- Test 1.5: MIN should return NULL
SELECT MIN(price) FROM products;
-- Expected output: 1 row with NULL

-- Test 1.6: MAX should return NULL
SELECT MAX(price) FROM products;
-- Expected output: 1 row with NULL

-- Test 1.7: Multiple aggregates with aliases
SELECT
    COUNT(*) AS total_products,
    SUM(price) AS total_value,
    AVG(price) AS avg_price,
    MIN(price) AS min_price,
    MAX(price) AS max_price
FROM products;
-- Expected output: 1 row with (0, NULL, NULL, NULL, NULL)

-- Test 2: Insert data, then delete all, then aggregate
-- Expected: Same behavior as empty table

INSERT INTO products VALUES
    (1, 'Laptop', 1099.00, 20, 'Electronics'),
    (2, 'Mouse', 29.99, 60, 'Electronics'),
    (3, 'Desk', 349.99, 25, 'Furniture'),
    (4, 'Chair', 199.99, 30, 'Furniture');

-- Verify data exists
SELECT COUNT(*) FROM products;
-- Expected output: 1 row with value 4

-- Delete all data
DELETE FROM products;

-- Verify empty
SELECT * FROM products;
-- Expected output: Empty set

-- Test 2.1: COUNT(*) after DELETE should return 0
SELECT COUNT(*) FROM products;
-- Expected output: 1 row with value 0

-- Test 2.2: SUM after DELETE should return NULL
SELECT SUM(price) FROM products;
-- Expected output: 1 row with NULL

-- Test 2.3: AVG after DELETE should return NULL
SELECT AVG(price) FROM products;
-- Expected output: 1 row with NULL

-- Test 3: GROUP BY on empty table
-- Expected: Empty result set (different from global aggregation!)

-- Test 3.1: GROUP BY on empty table
SELECT category, COUNT(*) FROM products GROUP BY category;
-- Expected output: Empty set

-- Test 3.2: GROUP BY with multiple aggregates
SELECT
    category,
    COUNT(*) AS product_count,
    SUM(price) AS total_value,
    AVG(price) AS avg_price
FROM products
GROUP BY category;
-- Expected output: Empty set

-- Test 3.3: GROUP BY with HAVING on empty table
SELECT category, COUNT(*) AS cnt
FROM products
GROUP BY category
HAVING cnt > 1;
-- Expected output: Empty set

-- Test 4: Edge cases

-- Test 4.1: Aggregate with WHERE clause (no matches)
INSERT INTO products VALUES (5, 'Keyboard', 79.99, 15, 'Electronics');
SELECT COUNT(*) FROM products WHERE price > 1000000;
-- Expected output: 1 row with value 0

-- Test 4.2: Multiple conditions
SELECT COUNT(*), SUM(price), AVG(quantity)
FROM products
WHERE category = 'NonExistent';
-- Expected output: 1 row with (0, NULL, NULL)

-- Clean up
DELETE FROM products;

-- Test 5: Verify behavior consistency across different aggregate functions

-- Test 5.1: COUNT variants
SELECT COUNT(*) FROM products;           -- Expected: 0
SELECT COUNT(id) FROM products;          -- Expected: 0
SELECT COUNT(name) FROM products;        -- Expected: 0
SELECT COUNT(price) FROM products;       -- Expected: 0

-- Test 5.2: Numeric aggregates
SELECT SUM(price) FROM products;         -- Expected: NULL
SELECT SUM(quantity) FROM products;      -- Expected: NULL
SELECT AVG(price) FROM products;         -- Expected: NULL
SELECT AVG(quantity) FROM products;      -- Expected: NULL
SELECT MIN(price) FROM products;         -- Expected: NULL
SELECT MAX(price) FROM products;         -- Expected: NULL

-- All tests completed
-- Summary:
-- ✅ Global aggregations (no GROUP BY) should return 1 row
-- ✅ COUNT returns 0, other aggregates return NULL
-- ✅ GROUP BY aggregations should return empty set
-- ✅ Behavior consistent after DELETE operations
