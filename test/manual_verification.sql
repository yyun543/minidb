-- 数据库管理
CREATE DATABASE ecommerce;
DROP DATABASE ecommerce;
USE ecommerce;
SHOW DATABASES;

-- 表管理(包含所有数据类型)
CREATE TABLE products (
    id INTEGER,
    name VARCHAR(100),
    price DOUBLE,
    quantity INTEGER,
    in_stock BOOLEAN,
    created_at TIMESTAMP,
    category VARCHAR
);

-- 带列约束的表
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL,
    age INTEGER DEFAULT 18,
    active BOOLEAN DEFAULT 1
);

-- 带表级PRIMARY KEY约束的表
CREATE TABLE orders (
    order_id INTEGER,
    user_id INTEGER NOT NULL,
    amount DOUBLE,
    PRIMARY KEY (order_id)
);

-- 带分区的表(HASH)
CREATE TABLE logs_hash (
    id INTEGER,
    message VARCHAR,
    timestamp TIMESTAMP
) PARTITION BY HASH (id);

-- 带分区的表(RANGE)
CREATE TABLE logs_range (
    id INTEGER,
    region VARCHAR,
    data VARCHAR
) PARTITION BY RANGE (id);

-- 删除表
DROP TABLE products;
DROP TABLE users;
SHOW TABLES;

-- 索引管理
CREATE INDEX idx_category ON products (category);
CREATE UNIQUE INDEX idx_id ON products (id);
CREATE INDEX idx_composite ON products (category, name);
DROP INDEX idx_category ON products;
SHOW INDEXES ON products;
SHOW INDEXES FROM products;
-- 单行插入
INSERT INTO products VALUES (1, 'Laptop', 999.99, 10, 1, '2024-01-01', 'Electronics');

-- 批量插入(多行)
INSERT INTO products VALUES
    (2, 'Mouse', 29.99, 50, 1, '2024-01-02', 'Electronics'),
    (3, 'Desk', 299.99, 15, 1, '2024-01-03', 'Furniture'),
    (4, 'Chair', 199.99, 20, 1, '2024-01-04', 'Furniture');

-- 指定列插入
INSERT INTO products (id, name, price, category)
VALUES (5, 'Monitor', 399.99, 'Electronics');

-- 指定列批量插入
INSERT INTO users (id, username, email, age) VALUES
    (1, 'alice', 'alice@example.com', 25),
    (2, 'bob', 'bob@example.com', 30),
    (3, 'charlie', 'charlie@example.com', 35);

-- 基本查询
SELECT * FROM products;
SELECT name, price FROM products;

-- 带列别名的查询
SELECT name AS product_name, price AS product_price FROM products;
SELECT name product_name, price product_price FROM products;

-- 带表别名的查询
SELECT p.name, p.price FROM products AS p;
SELECT p.name, p.price FROM products p;

-- WHERE子句(比较运算符)
SELECT * FROM products WHERE price > 100;
SELECT * FROM products WHERE price >= 100;
SELECT * FROM products WHERE price < 500;
SELECT * FROM products WHERE price <= 500;
SELECT * FROM products WHERE price = 299.99;
SELECT * FROM products WHERE price != 299.99;
SELECT * FROM products WHERE name = 'Laptop';

-- WHERE子句(逻辑运算符 AND, OR)
SELECT * FROM products WHERE category = 'Electronics' AND price < 1000;
SELECT * FROM products WHERE category = 'Electronics' OR category = 'Furniture';
SELECT * FROM products WHERE price > 100 AND price < 500 AND in_stock = true;

-- WHERE子句(LIKE模式匹配)
SELECT * FROM products WHERE name LIKE '%top';
SELECT * FROM products WHERE name LIKE 'M%';
SELECT * FROM products WHERE category LIKE '%onic%';
SELECT * FROM products WHERE name NOT LIKE 'Desk';

-- WHERE子句(IN运算符)
SELECT * FROM products WHERE category IN ('Electronics', 'Furniture');
SELECT * FROM products WHERE id IN (1, 2, 3);
SELECT * FROM products WHERE category NOT IN ('Obsolete', 'Discontinued');

-- WHERE子句(限定列引用)
SELECT * FROM products WHERE products.price > 200;

-- WHERE子句(括号表达式)
SELECT * FROM products WHERE (price > 100 AND category = 'Electronics') OR (price > 200 AND category = 'Furniture');

-- 单列更新
UPDATE products SET price = 1099 WHERE id = 1;

-- 多列更新*
UPDATE products SET price = 349.99, quantity = 25, in_stock = true WHERE id = 3;

-- 带表达式的更新*
UPDATE products SET price = price * 1.1 WHERE category = 'Electronics';
UPDATE products SET quantity = quantity + 10 WHERE in_stock = true;

-- 不带WHERE的更新(更新所有行)*
UPDATE products SET in_stock = true;

-- 带WHERE的删除
DELETE FROM products WHERE price < 50;
DELETE FROM products WHERE category = 'Obsolete';
DELETE FROM products WHERE id = 5;

-- 不带WHERE的删除(删除所有行 - 谨慎使用)
DELETE FROM products;
-- 聚合函数
SELECT COUNT(*) FROM products;
SELECT COUNT(name) FROM products;
SELECT SUM(price) FROM products;
SELECT AVG(price) FROM products;
SELECT MIN(price) FROM products;
SELECT MAX(price) FROM products;

-- 带别名的聚合函数
SELECT
    COUNT(*) AS total_products,
    SUM(price) AS total_value,
    AVG(price) AS avg_price,
    MIN(price) AS min_price,
    MAX(price) AS max_price
FROM products;

-- GROUP BY与聚合
SELECT category, COUNT(*) FROM products GROUP BY category;
SELECT category, AVG(price) FROM products GROUP BY category;
SELECT category, SUM(quantity) FROM products GROUP BY category;

-- GROUP BY与多个聚合函数
SELECT
    category,
    COUNT(*) AS product_count,
    SUM(price) AS total_value,
    AVG(price) AS avg_price,
    MIN(price) AS min_price,
    MAX(price) AS max_price
FROM products
GROUP BY category;

-- GROUP BY与HAVING子句*
SELECT category, COUNT(*) AS cnt FROM products GROUP BY category HAVING cnt > 5;
SELECT category, AVG(price) AS avg_price FROM products GROUP BY category HAVING avg_price > 100;
SELECT category, SUM(price) AS total FROM products GROUP BY category HAVING total > 1000;

-- ORDER BY(升序为默认)
SELECT * FROM products ORDER BY price;
SELECT * FROM products ORDER BY price ASC;
SELECT * FROM products ORDER BY name ASC;

-- ORDER BY降序
SELECT * FROM products ORDER BY price DESC;
SELECT * FROM products ORDER BY quantity DESC;

-- ORDER BY多列
SELECT * FROM products ORDER BY category ASC, price DESC;
SELECT * FROM products ORDER BY in_stock DESC, price ASC, name ASC;

-- ORDER BY表达式
SELECT name, price FROM products ORDER BY price * quantity DESC;

-- LIMIT子句
SELECT * FROM products LIMIT 10;
SELECT * FROM products ORDER BY price DESC LIMIT 5;
SELECT name, price FROM products WHERE category = 'Electronics' ORDER BY price LIMIT 3;

-- 组合: WHERE + GROUP BY + HAVING + ORDER BY + LIMIT
SELECT
    category,
    COUNT(*) AS product_count,
    AVG(price) AS avg_price
FROM products
WHERE in_stock = 1
GROUP BY category
HAVING product_count > 2
ORDER BY avg_price DESC
LIMIT 10;

-- INNER JOIN
SELECT u.name, o.amount, o.order_date
FROM users u
INNER JOIN orders o ON u.id = o.user_id;

-- JOIN(等同于INNER JOIN)
SELECT u.name, o.amount
FROM users u
JOIN orders o ON u.id = o.user_id
WHERE u.age > 25;

-- LEFT JOIN (LEFT OUTER JOIN)
SELECT u.name, o.amount
FROM users u
LEFT JOIN orders o ON u.id = o.user_id;

SELECT u.name, o.amount
FROM users u
LEFT OUTER JOIN orders o ON u.id = o.user_id;

-- RIGHT JOIN (RIGHT OUTER JOIN)
SELECT u.name, o.amount
FROM users u
RIGHT JOIN orders o ON u.id = o.user_id;

SELECT u.name, o.amount
FROM users u
RIGHT OUTER JOIN orders o ON u.id = o.user_id;

-- FULL OUTER JOIN
SELECT u.name, o.amount
FROM users u
FULL JOIN orders o ON u.id = o.user_id;

SELECT u.name, o.amount
FROM users u
FULL OUTER JOIN orders o ON u.id = o.user_id;

-- 多表JOIN
SELECT u.name, o.amount, p.name AS product_name
FROM users u
JOIN orders o ON u.id = o.user_id
JOIN products p ON o.product_id = p.id;

-- JOIN与WHERE子句
SELECT u.name, o.amount
FROM users u
JOIN orders o ON u.id = o.user_id
WHERE o.amount > 100 AND u.age > 25;

-- JOIN与聚合
SELECT u.name, COUNT(*) AS order_count, SUM(o.amount) AS total_spent
FROM users u
JOIN orders o ON u.id = o.user_id
GROUP BY u.name;

-- FROM子句中的子查询
SELECT sub.category, sub.avg_price
FROM (SELECT category, AVG(price) AS avg_price FROM products GROUP BY category) AS sub
WHERE sub.avg_price > 100;

-- 子查询与JOIN
SELECT u.name, sub.total
FROM users u
JOIN (SELECT user_id, SUM(amount) AS total FROM orders GROUP BY user_id) AS sub ON u.id = sub.user_id;

-- SELECT中的函数调用表达式
SELECT name, price, price * 1.1 AS price_with_tax FROM products;
SELECT name, price, quantity, price * quantity AS total_value FROM products;
SELECT UPPER(name) AS upper_name FROM products;
SELECT COUNT(*), AVG(price) FROM products WHERE category = 'Electronics';
-- 查询所有数据库
SELECT * FROM sys.db_metadata;

-- 查询所有表
SELECT db_name, table_name FROM sys.table_metadata;

-- 查询表结构
SELECT column_name, data_type, is_nullable
FROM sys.columns_metadata
WHERE db_name = 'ecommerce' AND table_name = 'products';

-- 查询索引
SELECT index_name, column_name, is_unique
FROM sys.index_metadata
WHERE db_name = 'ecommerce' AND table_name = 'products';

-- 查询事务历史
SELECT version, operation, table_id, file_path, row_count
FROM sys.delta_log
WHERE table_id LIKE 'ecommerce%'
ORDER BY version DESC
LIMIT 20;

-- 查询表文件
SELECT file_path, file_size, row_count, status
FROM sys.table_files
WHERE db_name = 'ecommerce' AND table_name = 'products';

-- 系统表与聚合
SELECT db_name, COUNT(*) AS table_count
FROM sys.table_metadata
GROUP BY db_name;

-- 系统表与JOIN
SELECT t.table_name, COUNT(c.column_name) AS column_count
FROM sys.table_metadata t
JOIN sys.columns_metadata c ON t.table_name = c.table_name
GROUP BY t.table_name;
-- 开始事务
START TRANSACTION;

-- 提交更改
COMMIT;

-- 回滚更改
ROLLBACK;

-- 事务示例(多个操作)
START TRANSACTION;
INSERT INTO products VALUES (10, 'Keyboard', 79.99, 30, 1, '2024-01-10', 'Electronics');
UPDATE products SET price = price * 0.9 WHERE category = 'Electronics';
DELETE FROM products WHERE quantity = 0;
COMMIT;
-- 查看SELECT查询的执行计划
EXPLAIN SELECT * FROM products WHERE category = 'Electronics';

-- EXPLAIN与JOIN
EXPLAIN SELECT u.name, o.amount
FROM users u
JOIN orders o ON u.id = o.user_id
WHERE u.age > 25;

-- EXPLAIN与复杂查询
EXPLAIN SELECT
    category,
    COUNT(*) AS product_count,
    AVG(price) AS avg_price
FROM products
WHERE in_stock = 1
GROUP BY category
HAVING product_count > 2
ORDER BY avg_price DESC
LIMIT 10;

-- 分析表统计信息(所有列)
ANALYZE TABLE products;

-- 分析特定列
ANALYZE TABLE products (price, quantity);
ANALYZE TABLE users (age, active);
