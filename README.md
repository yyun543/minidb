# MiniDB

MiniDB is a lightweight HTAP (Hybrid Transactional/Analytical Processing) database system implemented in Go. It aims to support both OLTP and OLAP workloads efficiently and may be extended to a distributed database in the future.

## Features

### Storage Engine

- Hybrid storage architecture: Currently leveraging in-memory row store with future consideration for column store integration for optimized analytical queries.
- Thread-safe concurrent access managed with appropriate synchronization primitives.
- WAL (Write-Ahead Logging) support for basic transaction durability.

### Query Processing

- SQL Query Support:
  - DDL (Data Definition Language):
    - `CREATE TABLE` - Create tables with schema definition
    - `DROP TABLE` - Remove existing tables
  - DML (Data Manipulation Language):
    - `SELECT` - Query data with `WHERE`, `JOIN`, `GROUP BY` clauses
    - `INSERT` - Add new records
    - `UPDATE` - Modify existing records with `WHERE` clause
    - `DELETE` - Remove records with `WHERE` clause
- Basic index support for faster lookups on indexed columns.

### Network Layer

- TCP server architecture to handle client connections.
- Connection handling with basic timeout management.

### Query Features

- `WHERE` clause with comparison operators (`=`, `>`, `<`, `>=`, `<=`, `<>`, `LIKE`, `IN`)
- `JOIN` operations (`INNER JOIN`)
- `GROUP BY` with basic aggregation functions (`COUNT`, `AVG`, `SUM`, `MIN`, `MAX`).
- Column aliases (`AS`)

## Project Structure

```bash
minidb/
├── cmd/
│   └── server/                 # Application entry point
│       └── main.go             # Starts the server, handles client connections
│       └── handler.go          # Handles query requests, calls Parser, Optimizer, Executor
├── internal/
│   ├── catalog/                # Metadata management
│   │   ├── catalog.go          # Catalog struct, database/table management
│   │   ├── metadata.go         # TableMeta, ColumnMeta definitions
│   ├── executor/               # SQL execution engine
│   │   ├── executor.go         # Receives query plan, drives execution
│   │   ├── operators/          # Execution operators
│   │   │   ├── table_scan.go    # Table scan
│   │   │   ├── filter.go        # Filter
│   │   │   ├── join.go          # Join (Nested Loop)
│   │   │   └── aggregate.go     # Aggregate
│   │   ├── interface.go        # Defines Executor, Operator interfaces
│   │   └── context.go          # Execution context
│   ├── optimizer/              # Query optimizer
│   │   ├── optimizer.go        # Receives AST, drives optimization, generates query plan
│   │   ├── plan.go             # Query plan (operator tree) definition
│   ├── parser/                 # SQL parser
│   │   ├── MiniQL.g4           # ANTLR4 grammar definition
│   │   ├── gen/                  # ANTLR-generated code
│   │   │   ├── MiniQLLexer.go
│   │   │   └── MiniQLParser.go
│   │   ├── parser.go           # Encapsulates ANTLR parsing, builds AST
│   │   └── visitor.go          # AST Visitor implementation (syntax checking, etc.)
│   ├── storage/                # Storage engine
│   │   ├── memtable.go         # In-memory table
│   │   ├── wal.go              # WAL (Write-Ahead Log)
│   │   ├── storage.go          # Defines storage engine interface
│   │   └── index.go            # Index (BTree)
│   └── types/                  # Data type system
│       └── types.go            # Defines data types, conversions, etc.
├── proto/                      # (Optional) Protobuf definitions (distributed/RPC)
│   └── minidb.proto
└── test/
    ├── catalog_test.go       # Catalog unit tests
    ├── executor_test.go      # Executor unit tests
    ├── parser_test.go        # Parser unit tests
    └── storage_test.go       # Storage unit tests
```

## Current Limitations

- Primarily in-memory storage. Persistence is achieved through WAL but requires explicit recovery.
- Basic transaction support (no MVCC or snapshot isolation).
- Limited JOIN support (Nested Loop INNER JOIN only).
- Basic `GROUP BY` support with limited aggregation functions.
- No cost-based query optimizer; uses rule-based optimizations.
- No support for foreign keys or constraints.
- No support for prepared statements.
- No built-in authentication/authorization.

## Future Improvements

The roadmap for MiniDB includes several exciting enhancements, aiming to improve its functionality, performance, and scalability.  A key consideration is the potential to evolve MiniDB into a distributed database system.

1. **Storage:**
   - Full persistence with snapshotting and recovery mechanisms.
   - Write-ahead logging enhancements for robust transaction support.
   - MVCC transaction support for concurrency and isolation.
   - Buffer pool management for efficient memory utilization.
   - Columnar storage option for analytic workloads.

2. **Query Processing:**
   - Cost-based query optimizer for intelligent query plan selection.
   - Support for more JOIN types (LEFT, RIGHT, FULL).
   - Advanced aggregation functions (e.g., window functions, percentiles).
   - Subquery support.

3. **Features:**
   - Authentication and authorization mechanisms for secure access.
   - Prepared statements for parameterized queries.
   - Foreign key constraints for data integrity.
   - Triggers for event-driven actions.
   - Views for simplified query interfaces.

4. **Performance:**
   - Query plan caching for repeated queries.
   - Enhanced index structures (e.g., B+tree) for efficient lookups.
   - Statistics collection for optimizer hints.
   - Query parallelization for improved throughput.

5. **Distribution:**
   - Sharding of data across multiple nodes.
   - Distributed query execution.
   - Consensus algorithms for data consistency and fault tolerance.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request following established coding conventions and testing guidelines.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Usage Examples

### Starting the Server

```bash
# Start server on default port 7205
./minidb

# Start server on custom port
./minidb -port 7205
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

### Using Indexes

```sql
-- Queries on indexed columns will be automatically optimized
SELECT * FROM users WHERE id = 1;
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
telnet localhost 7205

# Using netcat
nc localhost 7205
```

