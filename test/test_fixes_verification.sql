-- Test Fix 1: ProjectionPlan support in vectorized executor
-- This should work without "unsupported plan type" error

CREATE DATABASE test_fixes;
USE test_fixes;

CREATE TABLE products (
    id INTEGER,
    name VARCHAR(100),
    price DOUBLE,
    quantity INTEGER,
    in_stock BOOLEAN,
    created_at TIMESTAMP,
    category VARCHAR
);

-- Insert test data
INSERT INTO products VALUES (1, 'Laptop', 999.99, 10, 1, '2024-01-01', 'Electronics');
INSERT INTO products VALUES (2, 'Mouse', 29.99, 50, 1, '2024-01-02', 'Electronics');
INSERT INTO products VALUES (3, 'Desk', 299.99, 15, 1, '2024-01-03', 'Furniture');
INSERT INTO products VALUES (4, 'Chair', 199.99, 20, 1, '2024-01-04', 'Furniture');
INSERT INTO products VALUES (5, 'Monitor', 399.99, 8, 1, '2024-01-05', 'Electronics');

-- Test projection queries (should work with vectorized executor now)
SELECT * FROM products;
SELECT name, price FROM products;
SELECT id, name, category FROM products;

-- Test Fix 2: Transaction support
-- This should execute all statements within transaction

START TRANSACTION;
INSERT INTO products VALUES (10, 'Keyboard', 79.99, 30, 1, '2024-01-10', 'Electronics');
UPDATE products SET price = price * 0.9 WHERE category = 'Electronics';
DELETE FROM products WHERE quantity = 0;
COMMIT;

-- Verify transaction results
SELECT * FROM products WHERE id = 10;
SELECT name, price FROM products WHERE category = 'Electronics';

-- Test rollback
START TRANSACTION;
INSERT INTO products VALUES (20, 'Test Product', 99.99, 5, 1, '2024-01-20', 'Test');
ROLLBACK;

-- This should not exist (rollback test)
SELECT * FROM products WHERE id = 20;

-- Cleanup
DROP TABLE products;
DROP DATABASE test_fixes;
