CREATE DATABASE ecommerce;
USE ecommerce;
CREATE TABLE products (id INT, name VARCHAR);
CREATE TABLE orders (id INT, product_id INT);
INSERT INTO products (id, name) VALUES (1, 'Laptop');
INSERT INTO orders (id, product_id) VALUES (100, 1);
SELECT * FROM products;
SELECT * FROM orders;
SHOW TABLES;
