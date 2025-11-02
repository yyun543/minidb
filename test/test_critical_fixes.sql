-- Critical Bugs Fix Test Script
-- Test all the major fixes

-- Setup: Create database and tables
CREATE DATABASE test_fixes;
USE test_fixes;

-- Create products table
CREATE TABLE products (
    id INT,
    name VARCHAR,
    price FLOAT,
    quantity INT,
    in_stock BOOLEAN,
    created_at VARCHAR,
    category VARCHAR
);

-- Insert test data
INSERT INTO products VALUES (1, 'Laptop', 999.99, 10, 1, '2024-01-01', 'Electronics');
INSERT INTO products VALUES (2, 'Mouse', 29.99, 50, 1, '2024-01-02', 'Electronics');
INSERT INTO products VALUES (3, 'Desk', 299.99, 15, 1, '2024-01-03', 'Furniture');
INSERT INTO products VALUES (4, 'Chair', 199.99, 20, 1, '2024-01-04', 'Furniture');
INSERT INTO products VALUES (5, 'Monitor', 399.99, 0, 0, '', 'Electronics');
INSERT INTO products VALUES (10, 'Keyboard', 79.99, 30, 1, '2024-01-10', 'Electronics');

-- Test 1: IN operator with strings
SELECT * FROM products WHERE category IN ('Electronics', 'Furniture');

-- Test 2: IN operator with integers
SELECT * FROM products WHERE id IN (1, 2, 3);

-- Test 3: NOT IN operator
SELECT * FROM products WHERE category NOT IN ('Obsolete', 'Discontinued');

-- Create users and orders tables for JOIN tests
CREATE TABLE users (
    id INT,
    username VARCHAR,
    email VARCHAR,
    age INT,
    active BOOLEAN
);

INSERT INTO users VALUES (1, 'alice', 'alice@example.com', 25, 0);
INSERT INTO users VALUES (2, 'bob', 'bob@example.com', 30, 0);
INSERT INTO users VALUES (3, 'charlie', 'charlie@example.com', 35, 0);

CREATE TABLE orders (
    order_id INT,
    product_id INT,
    user_id INT,
    amount FLOAT
);

INSERT INTO orders VALUES (1, 1, 2, 19.99);
INSERT INTO orders VALUES (2, 2, 3, 29.99);
INSERT INTO orders VALUES (3, 3, 1, 38.88);

-- Test 4: INNER JOIN
SELECT u.username, o.amount, o.order_id
FROM users u
INNER JOIN orders o ON u.id = o.user_id;

-- Test 5: LEFT JOIN (should show all users, NULL for users without orders)
SELECT u.username, o.amount
FROM users u
LEFT JOIN orders o ON u.id = o.user_id;

-- Test 6: Multi-table JOIN
SELECT u.username, o.amount, p.name AS product_name
FROM users u
JOIN orders o ON u.id = o.user_id
JOIN products p ON o.product_id = p.id;

-- Test 7: JOIN with WHERE clause
SELECT u.username, o.amount
FROM users u
JOIN orders o ON u.id = o.user_id
WHERE o.amount > 10 AND u.age > 25;

-- Test 8: JOIN with aggregation
SELECT u.username, COUNT(*) AS order_count, SUM(o.amount) AS total_spent
FROM users u
JOIN orders o ON u.id = o.user_id
GROUP BY u.username;

-- Test 9: Expression evaluation - multiplication
SELECT name, price, quantity, price * quantity AS total_value
FROM products
WHERE quantity > 0;

-- Test 10: Expression evaluation - arithmetic with constant
SELECT name, price, price * 1.1 AS price_with_tax
FROM products
WHERE in_stock = 1;

-- Clean up
-- DROP TABLE orders;
-- DROP TABLE users;
-- DROP TABLE products;
-- DROP DATABASE test_fixes;
