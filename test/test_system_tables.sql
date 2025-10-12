-- Test queries for verifying system table fix
-- Column order should now be correct

-- Test 1: Query all system databases
SELECT * FROM sys.schemata;

-- Test 2: Query all system tables
SELECT * FROM sys.table_catalog;

-- Test 3: Create test data
CREATE DATABASE testdb;
USE testdb;
CREATE TABLE users (id INT, name VARCHAR);
INSERT INTO users (id, name) VALUES (1, 'Alice');

-- Test 4: Query delta_log to verify column order
-- Expected columns: version, timestamp, operation, table_schema, table_name, file_path
SELECT * FROM sys.delta_log LIMIT 10;

-- Test 5: Query delta_log with WHERE clause
SELECT version, operation, table_schema, table_name
FROM sys.delta_log
WHERE table_schema = 'testdb';

-- Test 6: Query columns table
SELECT table_schema, table_name, column_name, data_type
FROM sys.columns
WHERE table_schema = 'testdb' AND table_name = 'users';
