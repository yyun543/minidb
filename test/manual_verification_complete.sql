-- ============================================================================
-- MiniDB Complete Manual Verification Script
-- This script tests ALL SQL features supported by MiniDB
-- Based on: README.md SQL examples + MiniQL.g4 grammar
-- ============================================================================

-- ============================================================================
-- SECTION 1: DATABASE MANAGEMENT (DDL)
-- ============================================================================

-- 1.1 Create and use databases
CREATE DATABASE testdb;
CREATE DATABASE ecommerce;
SHOW DATABASES;
USE testdb;
USE ecommerce;

-- ============================================================================
-- SECTION 2: TABLE MANAGEMENT (DDL)
-- ============================================================================

USE testdb;

-- 2.1 Basic table creation with all data types
CREATE TABLE products (
    id INTEGER,
    name VARCHAR,
    price DOUBLE,
    quantity INTEGER,
    in_stock BOOLEAN,
    created_at TIMESTAMP,
    category VARCHAR
);

-- 2.2 Table with column constraints
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    username VARCHAR NOT NULL UNIQUE,
    email VARCHAR NOT NULL,
    age INTEGER DEFAULT 18,
    active BOOLEAN DEFAULT 1
);

-- 2.3 Table with table-level PRIMARY KEY constraint
CREATE TABLE orders (
    order_id INTEGER,
    user_id INTEGER NOT NULL,
    amount DOUBLE,
    PRIMARY KEY (order_id)
);

-- 2.4 Table with HASH partitioning
CREATE TABLE logs_hash (
    id INTEGER,
    message VARCHAR,
    timestamp TIMESTAMP
) PARTITION BY HASH (id);

-- 2.5 Table with RANGE partitioning
CREATE TABLE logs_range (
    id INTEGER,
    region VARCHAR,
    data VARCHAR
) PARTITION BY RANGE (id);

-- 2.6 Show tables
SHOW TABLES;

-- ============================================================================
-- SECTION 3: INDEX MANAGEMENT (DDL)
-- ============================================================================

-- 3.1 Create various indexes
CREATE INDEX idx_category ON products (category);
CREATE UNIQUE INDEX idx_id ON products (id);
CREATE INDEX idx_composite ON products (category, name);

-- 3.2 Show indexes
SHOW INDEXES ON products;
SHOW INDEXES FROM products;

-- ============================================================================
-- SECTION 4: DATA INSERTION (DML)
-- ============================================================================

-- 4.1 Single row insertion
INSERT INTO products VALUES (1, 'Laptop', 999.99, 10, 1, '2024-01-01', 'Electronics');

-- 4.2 Multi-row insertion (批量插入)
INSERT INTO products VALUES
    (2, 'Mouse', 29.99, 50, 1, '2024-01-02', 'Electronics'),
    (3, 'Desk', 299.99, 15, 1, '2024-01-03', 'Furniture'),
    (4, 'Chair', 199.99, 20, 1, '2024-01-04', 'Furniture');

-- 4.3 Insert with column list
INSERT INTO products (id, name, price, category)
VALUES (5, 'Monitor', 399.99, 'Electronics');

-- 4.4 Multi-row insert with column list
INSERT INTO users (id, username, email, age) VALUES
    (1, 'alice', 'alice@example.com', 25),
    (2, 'bob', 'bob@example.com', 30),
    (3, 'charlie', 'charlie@example.com', 35);

-- 4.5 Insert into orders
INSERT INTO orders VALUES (100, 1, 150.50);
INSERT INTO orders VALUES (101, 2, 250.75);
INSERT INTO orders VALUES (102, 3, 99.99);

-- ============================================================================
-- SECTION 5: BASIC QUERIES (DQL)
-- ============================================================================

-- 5.1 Select all columns
SELECT * FROM products;
SELECT * FROM users;
SELECT * FROM orders;

-- 5.2 Select specific columns
SELECT name, price FROM products;
SELECT username, email FROM users;

-- 5.3 Column aliases (explicit with AS)
SELECT name AS product_name, price AS product_price FROM products;

-- 5.4 Column aliases (implicit without AS)
SELECT name product_name, price product_price FROM products;

-- 5.5 Table aliases (explicit with AS)
SELECT p.name, p.price FROM products AS p;

-- 5.6 Table aliases (implicit without AS)
SELECT p.name, p.price FROM products p;

-- ============================================================================
-- SECTION 6: WHERE CLAUSE (DQL)
-- ============================================================================

-- 6.1 Comparison operators
SELECT * FROM products WHERE price > 100;
SELECT * FROM products WHERE price >= 100;
SELECT * FROM products WHERE price < 500;
SELECT * FROM products WHERE price <= 500;
SELECT * FROM products WHERE price = 299.99;
SELECT * FROM products WHERE price != 299.99;
SELECT * FROM products WHERE name = 'Laptop';

-- 6.2 Logical operators (AND, OR)
SELECT * FROM products WHERE category = 'Electronics' AND price < 1000;
SELECT * FROM products WHERE category = 'Electronics' OR category = 'Furniture';
SELECT * FROM products WHERE price > 100 AND price < 500 AND in_stock = 1;

-- 6.3 LIKE operator (pattern matching)
SELECT * FROM products WHERE name LIKE '%top';
SELECT * FROM products WHERE name LIKE 'M%';
SELECT * FROM products WHERE category LIKE '%onic%';
SELECT * FROM products WHERE name NOT LIKE 'Desk';

-- 6.4 IN operator
SELECT * FROM products WHERE category IN ('Electronics', 'Furniture');
SELECT * FROM products WHERE id IN (1, 2, 3);
SELECT * FROM products WHERE category NOT IN ('Obsolete', 'Discontinued');

-- 6.5 Qualified column references
SELECT * FROM products WHERE products.price > 200;

-- 6.6 Parenthesized expressions
SELECT * FROM products WHERE (price > 100 AND category = 'Electronics') OR (price > 200 AND category = 'Furniture');

-- ============================================================================
-- SECTION 7: AGGREGATE FUNCTIONS (DQL)
-- ============================================================================

-- 7.1 Basic aggregate functions
SELECT COUNT(*) FROM products;
SELECT COUNT(name) FROM products;
SELECT SUM(price) FROM products;
SELECT AVG(price) FROM products;
SELECT MIN(price) FROM products;
SELECT MAX(price) FROM products;

-- 7.2 Aggregate functions with aliases
SELECT
    COUNT(*) AS total_products,
    SUM(price) AS total_value,
    AVG(price) AS avg_price,
    MIN(price) AS min_price,
    MAX(price) AS max_price
FROM products;

-- ============================================================================
-- SECTION 8: GROUP BY (DQL)
-- ============================================================================

-- 8.1 GROUP BY with single aggregation
SELECT category, COUNT(*) FROM products GROUP BY category;
SELECT category, AVG(price) FROM products GROUP BY category;
SELECT category, SUM(quantity) FROM products GROUP BY category;

-- 8.2 GROUP BY with multiple aggregations
SELECT
    category,
    COUNT(*) AS product_count,
    SUM(price) AS total_value,
    AVG(price) AS avg_price,
    MIN(price) AS min_price,
    MAX(price) AS max_price
FROM products
GROUP BY category;

-- ============================================================================
-- SECTION 9: HAVING CLAUSE (DQL)
-- ============================================================================

-- 9.1 HAVING with COUNT
SELECT category, COUNT(*) AS cnt FROM products GROUP BY category HAVING cnt > 1;

-- 9.2 HAVING with AVG
SELECT category, AVG(price) AS avg_price FROM products GROUP BY category HAVING avg_price > 100;

-- 9.3 HAVING with SUM
SELECT category, SUM(price) AS total FROM products GROUP BY category HAVING total > 500;

-- ============================================================================
-- SECTION 10: ORDER BY (DQL)
-- ============================================================================

-- 10.1 ORDER BY ascending (default)
SELECT * FROM products ORDER BY price;
SELECT * FROM products ORDER BY price ASC;
SELECT * FROM products ORDER BY name ASC;

-- 10.2 ORDER BY descending
SELECT * FROM products ORDER BY price DESC;
SELECT * FROM products ORDER BY quantity DESC;

-- 10.3 ORDER BY multiple columns
SELECT * FROM products ORDER BY category ASC, price DESC;
SELECT * FROM products ORDER BY in_stock DESC, price ASC, name ASC;

-- 10.4 ORDER BY with expressions
SELECT name, price FROM products ORDER BY price * quantity DESC;

-- ============================================================================
-- SECTION 11: LIMIT CLAUSE (DQL)
-- ============================================================================

-- 11.1 Basic LIMIT
SELECT * FROM products LIMIT 10;
SELECT * FROM products LIMIT 3;

-- 11.2 LIMIT with ORDER BY
SELECT * FROM products ORDER BY price DESC LIMIT 5;
SELECT * FROM products ORDER BY price DESC LIMIT 2;

-- 11.3 LIMIT with WHERE and ORDER BY
SELECT name, price FROM products WHERE category = 'Electronics' ORDER BY price LIMIT 3;

-- ============================================================================
-- SECTION 12: COMBINED QUERIES (DQL)
-- ============================================================================

-- 12.1 WHERE + GROUP BY + HAVING + ORDER BY + LIMIT
SELECT
    category,
    COUNT(*) AS product_count,
    AVG(price) AS avg_price
FROM products
WHERE in_stock = 1
GROUP BY category
HAVING product_count > 0
ORDER BY avg_price DESC
LIMIT 10;

-- ============================================================================
-- SECTION 13: JOIN OPERATIONS (DQL)
-- ============================================================================

-- 13.1 INNER JOIN
SELECT u.username, o.amount, o.order_id
FROM users u
INNER JOIN orders o ON u.id = o.user_id;

-- 13.2 JOIN (equivalent to INNER JOIN)
SELECT u.username, o.amount
FROM users u
JOIN orders o ON u.id = o.user_id;

-- 13.3 JOIN with WHERE clause
SELECT u.username, o.amount
FROM users u
JOIN orders o ON u.id = o.user_id
WHERE u.age > 25;

-- 13.4 LEFT JOIN (LEFT OUTER JOIN)
SELECT u.username, o.amount
FROM users u
LEFT JOIN orders o ON u.id = o.user_id;

SELECT u.username, o.amount
FROM users u
LEFT OUTER JOIN orders o ON u.id = o.user_id;

-- 13.5 RIGHT JOIN (RIGHT OUTER JOIN)
SELECT u.username, o.amount
FROM users u
RIGHT JOIN orders o ON u.id = o.user_id;

SELECT u.username, o.amount
FROM users u
RIGHT OUTER JOIN orders o ON u.id = o.user_id;

-- 13.6 FULL OUTER JOIN
SELECT u.username, o.amount
FROM users u
FULL JOIN orders o ON u.id = o.user_id;

SELECT u.username, o.amount
FROM users u
FULL OUTER JOIN orders o ON u.id = o.user_id;

-- 13.7 Multiple JOINs
INSERT INTO products (id, name, price, category) VALUES (6, 'Book', 19.99, 'Books');
SELECT u.username, o.amount, p.name AS product_name
FROM users u
JOIN orders o ON u.id = o.user_id
JOIN products p ON o.order_id = p.id;

-- 13.8 JOIN with aggregations
SELECT u.username, COUNT(*) AS order_count, SUM(o.amount) AS total_spent
FROM users u
JOIN orders o ON u.id = o.user_id
GROUP BY u.username;

-- ============================================================================
-- SECTION 13B: SUBQUERIES AND ADVANCED SELECT (DQL)
-- ============================================================================

-- NOTE: The following features may not be fully implemented yet
-- These are included for completeness based on README.md

-- 13B.1 Subquery in FROM clause (derived table)
-- Expected: Should return categories with average price > 100
-- SELECT sub.category, sub.avg_price
-- FROM (SELECT category, AVG(price) AS avg_price FROM products GROUP BY category) AS sub
-- WHERE sub.avg_price > 100;

-- 13B.2 Subquery with JOIN
-- Expected: Show users with their total order amounts
-- SELECT u.username, sub.total
-- FROM users u
-- JOIN (SELECT user_id, SUM(amount) AS total FROM orders GROUP BY user_id) AS sub
-- ON u.id = sub.user_id;

-- 13B.3 SELECT with function calls (if supported)
-- Expected: Convert product names to uppercase
-- SELECT UPPER(name) AS upper_name FROM products;

-- 13B.4 SELECT with complex expressions
-- Expected: Calculate price with 10% tax
-- SELECT name, price, price * 1.1 AS price_with_tax FROM products;

-- 13B.5 SELECT with multiple expressions
-- Expected: Calculate total value per product
-- SELECT name, price, quantity, price * quantity AS total_value FROM products;

-- ============================================================================
-- SECTION 14: UPDATE OPERATIONS (DML)
-- ============================================================================

-- 14.1 Single column update
UPDATE products SET price = 1099 WHERE id = 1;

-- 14.2 Multiple column update
UPDATE products SET price = 349.99, quantity = 25, in_stock = 1 WHERE id = 3;

-- 14.3 Update with expressions
UPDATE products SET price = price * 1.1 WHERE category = 'Electronics';
UPDATE products SET quantity = quantity + 10 WHERE in_stock = 1;

-- 14.4 Update without WHERE (updates all rows)
UPDATE products SET in_stock = 1;

-- ============================================================================
-- SECTION 15: DELETE OPERATIONS (DML)
-- ============================================================================

-- 15.1 Create test table for deletion
CREATE TABLE delete_test (id INTEGER, name VARCHAR, value INTEGER);
INSERT INTO delete_test VALUES (1, 'Test1', 100);
INSERT INTO delete_test VALUES (2, 'Test2', 200);
INSERT INTO delete_test VALUES (3, 'Test3', 300);
INSERT INTO delete_test VALUES (4, 'Test4', 400);

-- 15.2 DELETE with WHERE
DELETE FROM delete_test WHERE value < 200;
SELECT * FROM delete_test;

-- 15.3 DELETE without WHERE (deletes all rows)
DELETE FROM delete_test;
SELECT * FROM delete_test;

-- ============================================================================
-- SECTION 16: SYSTEM TABLES (DQL)
-- ============================================================================

-- 16.1 Query all databases
SELECT * FROM sys.db_metadata;

-- 16.2 Query all tables
SELECT db_name, table_name FROM sys.table_metadata;

-- 16.3 Query table structure
SELECT column_name, data_type, is_nullable
FROM sys.columns_metadata
WHERE db_name = 'testdb' AND table_name = 'products';

-- 16.4 Query indexes
SELECT index_name, column_name, is_unique
FROM sys.index_metadata
WHERE db_name = 'testdb' AND table_name = 'products';

-- 16.5 Query transaction history
SELECT version, operation, table_id, file_path, row_count
FROM sys.delta_log
WHERE table_id LIKE 'testdb%'
ORDER BY version DESC
LIMIT 20;

-- 16.6 Query table files
SELECT file_path, file_size, row_count, status
FROM sys.table_files
WHERE db_name = 'testdb' AND table_name = 'products';

-- 16.7 System tables with aggregations
SELECT db_name, COUNT(*) AS table_count
FROM sys.table_metadata
GROUP BY db_name;

-- 16.8 System tables with JOINs
SELECT t.table_name, COUNT(c.column_name) AS column_count
FROM sys.table_metadata t
JOIN sys.columns_metadata c ON t.table_name = c.table_name
GROUP BY t.table_name;

-- ============================================================================
-- SECTION 17: TRANSACTION CONTROL (DCL)
-- ============================================================================

-- 17.1 Basic transaction
START TRANSACTION;
INSERT INTO products VALUES (7, 'Tablet', 599.99, 5, 1, '2024-01-07', 'Electronics');
COMMIT;

-- 17.2 Transaction with rollback
START TRANSACTION;
INSERT INTO products VALUES (8, 'Headphones', 149.99, 20, 1, '2024-01-08', 'Electronics');
ROLLBACK;

-- 17.3 Transaction with multiple operations
START TRANSACTION;
INSERT INTO products VALUES (9, 'Keyboard', 79.99, 30, 1, '2024-01-09', 'Electronics');
UPDATE products SET price = price * 0.9 WHERE category = 'Electronics';
DELETE FROM products WHERE quantity = 0;
COMMIT;

-- ============================================================================
-- SECTION 18: UTILITY COMMANDS
-- ============================================================================

-- 18.1 EXPLAIN queries
EXPLAIN SELECT * FROM products WHERE category = 'Electronics';

EXPLAIN SELECT u.username, o.amount
FROM users u
JOIN orders o ON u.id = o.user_id
WHERE u.age > 25;

EXPLAIN SELECT
    category,
    COUNT(*) AS product_count,
    AVG(price) AS avg_price
FROM products
WHERE in_stock = 1
GROUP BY category
HAVING product_count > 0
ORDER BY avg_price DESC
LIMIT 10;

-- 18.2 ANALYZE TABLE (collect statistics)
ANALYZE TABLE products;
ANALYZE TABLE users;
ANALYZE TABLE products (price, quantity);
ANALYZE TABLE users (age, active);

-- ============================================================================
-- SECTION 19: DROP OPERATIONS (DDL)
-- ============================================================================

-- 19.1 Drop indexes
DROP INDEX idx_composite ON products;
DROP INDEX idx_category ON products;

-- 19.2 Drop tables
DROP TABLE delete_test;
DROP TABLE logs_hash;
DROP TABLE logs_range;
DROP TABLE orders;
DROP TABLE users;
DROP TABLE products;

-- 19.3 Show tables after drops
SHOW TABLES;

-- 19.4 Drop databases
DROP DATABASE testdb;
DROP DATABASE ecommerce;
SHOW DATABASES;

-- ============================================================================
-- SECTION 20: EDGE CASES AND SPECIAL SCENARIOS
-- ============================================================================

-- 20.1 Recreate test environment
CREATE DATABASE edge_cases;
USE edge_cases;

-- 20.2 Empty table queries
CREATE TABLE empty_table (id INTEGER, name VARCHAR);
SELECT * FROM empty_table;
SELECT COUNT(*) FROM empty_table;
DELETE FROM empty_table;

-- 20.3 Single row table
CREATE TABLE single_row (id INTEGER, value INTEGER);
INSERT INTO single_row VALUES (1, 100);
SELECT * FROM single_row;
UPDATE single_row SET value = 200;
SELECT * FROM single_row;
DELETE FROM single_row WHERE id = 1;
SELECT * FROM single_row;

-- 20.4 Null values handling
CREATE TABLE null_test (id INTEGER, name VARCHAR, value INTEGER);
INSERT INTO null_test VALUES (1, 'Test1', 100);
INSERT INTO null_test (id, name) VALUES (2, 'Test2');
SELECT * FROM null_test;

-- 20.5 Boolean values
CREATE TABLE bool_test (id INTEGER, active BOOLEAN);
INSERT INTO bool_test VALUES (1, 1);
INSERT INTO bool_test VALUES (2, 0);
SELECT * FROM bool_test WHERE active = 1;
SELECT * FROM bool_test WHERE active = 0;

-- 20.6 Timestamp handling
CREATE TABLE time_test (id INTEGER, created_at TIMESTAMP);
INSERT INTO time_test VALUES (1, '2024-01-01');
INSERT INTO time_test VALUES (2, '2024-12-31');
SELECT * FROM time_test;

-- 20.7 VARCHAR with special characters
CREATE TABLE varchar_test (id INTEGER, text VARCHAR);
INSERT INTO varchar_test VALUES (1, 'Hello World');
INSERT INTO varchar_test VALUES (2, 'Test@123');
INSERT INTO varchar_test VALUES (3, 'Line1
Line2');
SELECT * FROM varchar_test;

-- 20.8 Expression in SELECT
CREATE TABLE expr_test (id INTEGER, a INTEGER, b INTEGER);
INSERT INTO expr_test VALUES (1, 10, 20);
INSERT INTO expr_test VALUES (2, 30, 40);
SELECT id, a, b, a + b AS sum, a * b AS product FROM expr_test;

-- 20.9 Cleanup
DROP TABLE empty_table;
DROP TABLE single_row;
DROP TABLE null_test;
DROP TABLE bool_test;
DROP TABLE time_test;
DROP TABLE varchar_test;
DROP TABLE expr_test;
DROP DATABASE edge_cases;

-- ============================================================================
-- VERIFICATION COMPLETE
-- ============================================================================

SHOW DATABASES;

-- Expected: Only 'sys' and 'default' databases should remain
-- All user-created databases should be cleaned up

-- ============================================================================
-- END OF COMPLETE MANUAL VERIFICATION SCRIPT
-- ============================================================================
