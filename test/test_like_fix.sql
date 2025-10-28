-- Test LIKE query fix after adding Boolean type support to Filter operator
CREATE DATABASE testdb;
USE testdb;

-- Create table with BOOLEAN column (this was causing the panic)
CREATE TABLE products (
    id INTEGER,
    name VARCHAR,
    price DOUBLE,
    quantity INTEGER,
    in_stock BOOLEAN
);

-- Insert test data
INSERT INTO products VALUES (1, 'Laptop', 999.99, 10, 1);
INSERT INTO products VALUES (2, 'Mouse', 29.99, 50, 1);
INSERT INTO products VALUES (3, 'Monitor', 299.99, 20, 1);
INSERT INTO products VALUES (4, 'Keyboard', 79.99, 30, 0);

-- Basic SELECT to verify data
SELECT * FROM products;

-- Test LIKE queries that were causing panic
SELECT * FROM products WHERE name LIKE '%top';
SELECT * FROM products WHERE name LIKE 'M%';
SELECT * FROM products WHERE name LIKE '%o%';

-- Test other filters with BOOLEAN column present
SELECT * FROM products WHERE price > 100;
SELECT * FROM products WHERE quantity > 20;
SELECT * FROM products WHERE in_stock = 1;
