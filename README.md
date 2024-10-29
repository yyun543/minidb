# MiniDB

MiniDB is a lightweight HTAP (Hybrid Transactional/Analytical Processing) database system implemented in Go. It supports both OLTP and OLAP workloads through a hybrid storage engine.

## Features

### Storage Engine
- Hybrid storage architecture:
  - Row Store for OLTP workloads
  - Column Store for OLAP workloads
- Thread-safe concurrent access with RWMutex
- Basic transaction support

### Query Processing
- SQL Query Support:
  - DDL (Data Definition Language):
    - CREATE TABLE - Create tables with schema definition
    - DROP TABLE - Remove existing tables
    - SHOW TABLES - List all tables
  - DML (Data Manipulation Language):
    - SELECT - Query data with WHERE, JOIN, GROUP BY clauses
    - INSERT - Add new records
    - UPDATE - Modify existing records with WHERE clause
    - DELETE - Remove records with WHERE clause
- Query result caching with TTL
- Basic index support for faster lookups

### Network Layer
- TCP server/client architecture
- Connection pooling with max connections limit
- Timeout handling and retry mechanism
- Graceful shutdown support

### Query Features
- WHERE clause with comparison operators (=, >, <, >=, <=, <>, LIKE, IN)
- JOIN operations (INNER JOIN)
- GROUP BY with basic aggregation
- Column aliases (AS)
- LIMIT and OFFSET support
- Result formatting in table style

## Project Structure

```bash
minidb/
├── cmd/
│   └── minidb/
│       └── main.go           # Application entry point
├── internal/
│   ├── cache/               # Query result caching
│   │   └── cache.go
│   ├── executor/           # SQL execution engine
│   │   ├── executor.go
│   │   ├── formatter.go
│   │   └── visitor.go
│   ├── index/             # Index management
│   │   └── index.go
│   ├── network/           # Network layer
│   │   └── server.go
│   ├── parser/            # SQL parser
│   │   ├── ast.go
│   │   ├── lexer.go
│   │   ├── parser.go
│   │   └── visitor.go
│   └── storage/           # Storage engine
│       ├── column_store.go
│       ├── engine.go
│       └── row_store.go
```

## Current Limitations

- In-memory storage only (no persistence)
- Basic transaction support (no MVCC)
- Limited JOIN support (only INNER JOIN)
- Basic GROUP BY support (limited aggregation functions)
- No query optimizer
- No support for foreign keys or constraints
- No support for prepared statements
- No authentication/authorization

## Future Improvements

1. Storage
   - Persistent storage
   - Write-ahead logging
   - MVCC transaction support
   - Buffer pool management

2. Query Processing
   - Cost-based query optimizer
   - More JOIN types (LEFT, RIGHT, FULL)
   - Advanced aggregation functions
   - Window functions
   - Subqueries

3. Features
   - Authentication and authorization
   - Prepared statements
   - Foreign key constraints
   - Triggers
   - Views

4. Performance
   - Query plan caching
   - Better index structures (B+tree)
   - Statistics collection
   - Query parallelization

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Usage Examples

### Starting the Server
```bash
# Start server on default port 8086
./minidb

# Start server on custom port
./minidb -port 3306
```

### DDL Operations
```sql
-- Create a new table
CREATE TABLE users (
    id INT,
    name VARCHAR,
    email VARCHAR,
    age INT,
    created_at VARCHAR
);

-- Show all tables
SHOW TABLES;

-- Drop a table
DROP TABLE users;
```

### DML Operations
```sql
-- Insert data
INSERT INTO users VALUES (1, 'John Doe', 'john@example.com', 25, '2024-01-01');
INSERT INTO users VALUES (2, 'Jane Smith', 'jane@example.com', 30, '2024-01-02');
INSERT INTO users VALUES (3, 'Bob Wilson', 'bob@example.com', 35, '2024-01-03');

-- Basic SELECT
SELECT * FROM users;
SELECT name, email FROM users;

-- WHERE clause with different operators
SELECT * FROM users WHERE age > 25;
SELECT * FROM users WHERE name LIKE 'J%';
SELECT * FROM users WHERE age IN (25, 30);
SELECT * FROM users WHERE age >= 25 AND age <= 35;

-- JOIN operations
SELECT u.name, o.order_id, o.amount 
FROM users u 
JOIN orders o ON u.id = o.user_id;

-- GROUP BY with aggregation
SELECT age, COUNT(*) as count 
FROM users 
GROUP BY age;

-- GROUP BY with HAVING
SELECT age, COUNT(*) as count 
FROM users 
GROUP BY age 
HAVING count > 1;

-- Update data
UPDATE users 
SET email = 'john.doe@example.com' 
WHERE id = 1;

-- Delete data
DELETE FROM users 
WHERE age < 25;
```

### Using Cache
```sql
-- First query execution (stored in cache)
SELECT * FROM users WHERE age > 30;

-- Subsequent identical queries within TTL will be served from cache
SELECT * FROM users WHERE age > 30;
```

### Using Indexes
```sql
-- Queries on indexed columns will be automatically optimized
SELECT * FROM users WHERE id = 1;
SELECT * FROM users WHERE email = 'john@example.com';
```

### Formatting Examples
```sql
-- Table style output
SELECT name, age FROM users;
+------------+-----+
| name       | age |
+------------+-----+
| John Doe   | 25  |
| Jane Smith | 30  |
| Bob Wilson | 35  |
+------------+-----+
3 rows in set

-- Empty result
SELECT * FROM users WHERE age > 100;
Empty set
```

### Error Handling Examples
```sql
-- Table already exists
CREATE TABLE users (...);
Error: table users already exists

-- Invalid column
SELECT invalid_column FROM users;
Error: column invalid_column does not exist

-- Invalid syntax
SELECT FROM users WHERE;
Error: syntax error near 'WHERE'
```

### Client Connection Example
```bash
# Using telnet
telnet localhost 8086

# Using netcat
nc localhost 8086
```
