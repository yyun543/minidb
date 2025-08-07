
# MiniDB

MiniDB is a distributed MPP (Massively Parallel Processing) database system built in Go, designed for analytical workloads. Currently implemented as a single-node prototype with vectorized execution and cost-based optimization, it provides the foundation for future distributed parallel processing capabilities.

## Current Implementation

### Core Database Engine ✅

- **In-Memory Storage**: MemTable-based storage with WAL persistence
- **Multi-Client Support**: TCP server with session management  
- **SQL Parser**: ANTLR4-based parser supporting DDL, DML, and query operations
- **Type System**: Basic data types (INT, VARCHAR) with schema validation

### Query Processing ✅

- **Dual Execution Engines**: Vectorized (Apache Arrow) and regular execution engines
- **Cost-Based Optimization**: Statistics-driven query plan selection
- **Basic Operations**: SELECT, INSERT, UPDATE, DELETE with WHERE clauses
- **Aggregations**: GROUP BY with COUNT, SUM, AVG, MIN, MAX and HAVING clauses

## MPP Architecture Goals 🚧

### Distributed Processing (Planned)

- **Query Coordinator**: Distributed query planning and execution coordination
- **Compute Nodes**: Parallel execution across multiple compute nodes
- **Data Distribution**: Automatic data partitioning and distribution strategies
- **Inter-Node Communication**: Efficient data transfer protocols between nodes

### Lakehouse Integration (Planned)

- **Object Storage**: S3, GCS, Azure Blob storage connectors
- **Multi-Format Support**: Parquet, ORC, Delta Lake, Iceberg readers
- **Schema Evolution**: Dynamic schema changes without data migration
- **Metadata Service**: Distributed catalog for transaction coordination

## Supported SQL Features ✅

### Currently Working
- **DDL**: `CREATE/DROP DATABASE`, `CREATE/DROP TABLE`
- **DML**: `INSERT`, `SELECT`, `UPDATE`, `DELETE` 
- **Queries**: `WHERE` clauses (=, >, <, >=, <=, AND, OR)
- **Aggregation**: `GROUP BY`, `HAVING` with COUNT, SUM, AVG, MIN, MAX
- **Utilities**: `USE database`, `SHOW TABLES/DATABASES`, `EXPLAIN`

### Limited Support ⚠️
- **JOIN operations** (basic implementation)
- **WHERE operators**: LIKE, IN, BETWEEN (fallback to regular engine)
- **ORDER BY** (basic sorting)

### Planned Enhancements 🔄
- **Advanced JOINs**: Hash join, sort-merge join algorithms  
- **Window Functions**: ROW_NUMBER, RANK, analytical functions
- **Complex Expressions**: Nested queries, CTEs, advanced operators

## Architecture & Performance

### Current Architecture ✅

1. **Single-Node Design**: TCP server with multi-client session support
2. **Dual Execution Engines**: Vectorized (Arrow) and regular execution engines
3. **Statistics Collection**: Background statistics for cost-based optimization  
4. **Modular Design**: Clean separation of parser, optimizer, executor, storage layers

### MPP Design Principles 🎯

1. **Distributed-First**: Architecture designed for horizontal scaling
2. **Compute-Storage Separation**: Independent scaling of processing and storage
3. **Parallel Processing**: Query parallelization across multiple nodes
4. **Elastic Compute**: Dynamic resource allocation based on workload

### Performance Characteristics

- **Current Prototype**: Single-node analytical query processing
- **Vectorized Operations**: 10-100x speedup for compatible analytical queries  
- **Session Management**: Support for multiple concurrent connections
- **Memory Efficiency**: Arrow-based columnar processing with efficient allocators

## Project Structure

```bash
minidb/
├── cmd/
│   └── server/                    # Application entry point
│       ├── main.go                # Server startup with CLI flags and signal handling
│       └── handler.go             # Enhanced query handling with dual execution engines
├── internal/
│   ├── catalog/                   # Metadata management
│   │   ├── catalog.go             # Database/table management with type system
│   │   ├── metadata.go            # Enhanced metadata with Arrow schema support
│   │   └── system_tables.go       # System catalog tables
│   ├── executor/                  # Dual execution engines
│   │   ├── executor.go            # Regular execution engine
│   │   ├── vectorized_executor.go # Apache Arrow vectorized execution engine
│   │   ├── cost_optimizer.go      # Cost-based query optimization
│   │   ├── data_manager.go        # Data access layer
│   │   └── operators/             # Execution operators
│   │       ├── table_scan.go      # Optimized table scanning
│   │       ├── filter.go          # Vectorized filtering
│   │       ├── join.go            # Cost-optimized joins
│   │       └── aggregate.go       # Vectorized aggregations
│   ├── optimizer/                 # Advanced query optimizer
│   │   ├── optimizer.go           # Rule-based and cost-based optimization
│   │   ├── plan.go                # Enhanced query plan representation
│   │   └── rules.go               # Optimization rules (predicate pushdown, etc.)
│   ├── parser/                    # SQL parser
│   │   ├── MiniQL.g4              # Comprehensive ANTLR4 grammar
│   │   ├── gen/                   # ANTLR-generated code
│   │   ├── parser.go              # SQL parsing with enhanced error handling
│   │   ├── visitor.go             # AST visitor implementation
│   │   └── ast.go                 # Complete AST node definitions
│   ├── storage/                   # Advanced storage engine
│   │   ├── memtable.go            # Enhanced in-memory table
│   │   ├── distributed.go         # Distributed storage foundations
│   │   ├── wal.go                 # Write-Ahead Logging
│   │   ├── storage.go             # Storage engine interfaces
│   │   └── index.go               # Indexing support (BTree)
│   ├── types/                     # Enhanced type system
│   │   ├── schema.go              # Strong type system with Arrow integration
│   │   ├── partition.go           # Partitioning strategies for distribution
│   │   ├── vectorized.go          # Vectorized batch processing
│   │   └── types.go               # Data type definitions and conversions
│   ├── statistics/                # Statistics collection system
│   │   └── statistics.go          # Table and column statistics management
│   └── session/                   # Session management
│       └── session.go             # Session lifecycle and cleanup
└── test/                          # Comprehensive test suite
    ├── catalog_test.go            # Catalog functionality tests
    ├── executor_test.go           # Execution engine tests
    ├── optimizer_test.go          # Query optimization tests
    ├── parser_test.go             # SQL parsing tests
    └── storage_test.go            # Storage engine tests
```

## Current Performance Status

### Prototype Metrics ✅
- **Test Coverage**: ~77% integration test success rate  
- **Vectorized Execution**: Automatic selection for compatible analytical queries
- **Connection Handling**: Multi-client TCP server with session isolation
- **Query Processing**: Basic analytical operations (GROUP BY, aggregations)

### Target MPP Performance 🎯  
- **Distributed Processing**: Linear scalability across compute clusters
- **Query Throughput**: Thousands of concurrent analytical queries
- **Data Volume**: Petabyte-scale data processing capabilities
- **Fault Tolerance**: Automatic failure recovery and query restart

## Installation & Usage

### Building MiniDB

```bash
# Clone the repository
git clone <repository-url>
cd minidb

# Build the optimized server
go build -o minidb ./cmd/server

# Run tests to verify installation
go test ./test/... -v
```

### Starting the Server

```bash
# Start single-node prototype (localhost:7205)
./minidb

# Start with custom configuration  
./minidb -host 0.0.0.0 -port 8080

# Show available options
./minidb -h
```

### Current Server Output

```
=== MiniDB Server ===
Version: 1.0 (MPP Prototype)
Listening on: localhost:7205
Features: Vectorized Execution, Cost-based Optimization, Statistics Collection
Ready for connections...
```

### Future MPP Cluster (Planned)

```bash  
# Start coordinator node
./minidb coordinator --port 7205

# Start compute nodes
./minidb compute --coordinator localhost:7205 --port 8001
./minidb compute --coordinator localhost:7205 --port 8002
```

## SQL Usage Examples

### Database Operations

```sql
-- Create and manage databases
CREATE DATABASE ecommerce;
USE ecommerce;
SHOW DATABASES;
```

### Basic DDL Operations ✅

```sql
-- Create database and tables (currently supported)
CREATE DATABASE analytics_demo;
USE analytics_demo;

CREATE TABLE sales (
    region VARCHAR,
    amount INT,
    product VARCHAR
);

SHOW DATABASES;
SHOW TABLES;
```

### Working DML Examples ✅

```sql
-- Insert sample data
INSERT INTO sales VALUES ('North', 100, 'ProductA');
INSERT INTO sales VALUES ('South', 150, 'ProductB');
INSERT INTO sales VALUES ('North', 200, 'ProductC');
INSERT INTO sales VALUES ('East', 300, 'ProductD');

-- Basic SELECT with WHERE (working)
SELECT * FROM sales;
SELECT region, amount FROM sales WHERE amount > 150;

-- GROUP BY aggregations (working with vectorized execution)  
SELECT region, COUNT(*) as orders, SUM(amount) as total
FROM sales GROUP BY region;

-- HAVING clauses (working)
SELECT region, COUNT(*) as cnt 
FROM sales GROUP BY region 
HAVING cnt >= 2;
```

### Query Optimization ✅

```sql
-- View execution plans (currently working)
EXPLAIN SELECT region, SUM(amount) as total_sales
FROM sales WHERE amount > 100
GROUP BY region;

-- Output shows:
-- Query Execution Plan:
--------------------
-- Select
--   GroupBy
--     Filter
--       TableScan
```

### Limited Features ⚠️

```sql
-- UPDATE/DELETE (basic support, may have limitations)
UPDATE sales SET amount = 250 WHERE region = 'North';
DELETE FROM sales WHERE amount < 50;

-- JOIN operations (basic implementation, may not work in all cases)  
-- CREATE TABLE customers (id INT, name VARCHAR);
-- SELECT s.region, c.name FROM sales s JOIN customers c ON ...;
-- Note: Complex JOINs may fail or fall back to regular engine
```

### Future MPP Features 🔮

```sql
-- Planned: Advanced analytical queries
SELECT 
    region,
    amount,
    SUM(amount) OVER (PARTITION BY region ORDER BY amount) as running_total,
    ROW_NUMBER() OVER (PARTITION BY region ORDER BY amount DESC) as rank
FROM sales;

-- Planned: Complex multi-table operations with distributed execution
SELECT 
    region,
    COUNT(DISTINCT product) as product_variety,
    AVG(amount) as avg_sale,
    PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY amount) as median_sale
FROM sales s
JOIN product_catalog p ON s.product = p.name
WHERE s.date >= '2024-01-01'
GROUP BY region
ORDER BY avg_sale DESC;
```

### Result Formatting

```sql
-- Formatted table output with row counts
SELECT name, age FROM users WHERE age > 25;

| name            | age            |
|-----------------+----------------|
| Jane Smith      | 30             |
| Bob Wilson      | 35             |
|-----------------+----------------|
2 rows in set

-- Empty result handling
SELECT * FROM users WHERE age > 100;
Empty set
```

### Error Handling

```sql
-- Comprehensive error messages
CREATE TABLE users (...);
Error: table users already exists

SELECT nonexistent_column FROM users;
Error: column nonexistent_column does not exist

SELECT FROM users WHERE;
Error: parsing error: syntax error near 'WHERE'
```

## Connection Examples

### Using Network Clients

```bash
# Connect using netcat
nc localhost 7205

# Connect using telnet
telnet localhost 7205

# Example session
Welcome to MiniDB v1.0!
Session ID: 1234567890
Type 'exit;' or 'quit;' to disconnect
------------------------------------
minidb> CREATE TABLE test (id INT, name VARCHAR);
OK

minidb> INSERT INTO test VALUES (1, 'Hello');
OK

minidb> SELECT * FROM test;
| id              | name           |
|-----------------+----------------|
| 1               | Hello          |
|-----------------+----------------|
1 rows in set

minidb> exit;
Goodbye!
```

## Development Status & Advantages

### Current Prototype Benefits ✅

1. **Vectorized Analytics**: Significant performance improvements for GROUP BY and aggregations
2. **Cost-Based Optimization**: Intelligent query plan selection using table statistics  
3. **Modular Architecture**: Clean separation enabling easy distributed expansion
4. **Arrow Integration**: Industry-standard columnar processing for analytical workloads

### MPP Architecture Advantages 🎯

1. **Linear Scalability**: Designed for horizontal scaling across compute clusters
2. **Compute-Storage Separation**: Independent scaling of processing and storage resources
3. **Fault Tolerance**: Automatic failure recovery and query restart capabilities  
4. **Elastic Resource Management**: Dynamic compute allocation based on workload patterns

### Developer Experience

1. **Simple Deployment**: Single binary with no external dependencies (current)
2. **Comprehensive Testing**: Integration test framework with ~77% success rate
3. **Clear Documentation**: Honest status reporting of working vs planned features
4. **MPP-Ready Design**: Minimal changes needed for distributed deployment

## MPP Roadmap

### Phase 1: MPP Foundation 🚧
- [ ] **Distributed Query Coordinator**: Central query planning and execution coordination
- [ ] **Compute Node Management**: Automatic node discovery and health monitoring  
- [ ] **Inter-Node Communication**: Efficient data transfer protocols between nodes
- [ ] **Query Distribution**: Automatic query parallelization across compute clusters
- [ ] **Resource Management**: Intelligent workload scheduling and resource allocation

### Phase 2: Lakehouse Integration 🔮
- [ ] **Object Storage Connectors**: S3, GCS, Azure Blob storage integration
- [ ] **Multi-Format Support**: Native Parquet, ORC, Delta Lake, Iceberg readers
- [ ] **Distributed Metadata Service**: Schema evolution and transaction coordination
- [ ] **Data Distribution**: Automatic partitioning and pruning for optimal performance
- [ ] **Elastic Compute**: Dynamic scaling based on workload demands

### Phase 3: Advanced Analytics 🌟  
- [ ] **Window Functions**: ROW_NUMBER, RANK, advanced analytical functions
- [ ] **Machine Learning Integration**: SQL-based ML algorithms
- [ ] **Real-time Streaming**: Live data ingestion and processing
- [ ] **Advanced Optimization**: Adaptive query execution and auto-tuning
- [ ] **Multi-tenant Support**: Resource isolation and security

## Contributing

We welcome contributions! Please follow these guidelines:

1. Ensure all tests pass: `go test ./test/... -v`
2. Follow the existing code architecture and patterns
3. Add appropriate unit tests for new features
4. Update documentation for user-facing changes

## Testing & Validation

### Current Testing Status
- **Integration Tests**: ~77% success rate across test framework
- **Working Features**: Basic DDL, DML, GROUP BY, aggregations
- **Vectorized Queries**: Functional for compatible analytical operations
- **Connection Handling**: Multi-client TCP server with session management

### Target MPP Benchmarks 🎯
- **Distributed Processing**: Linear scalability across compute clusters
- **Query Throughput**: Support for thousands of concurrent analytical queries  
- **Data Volume**: Petabyte-scale processing capabilities
- **Fault Tolerance**: Sub-second failure detection and recovery

## License

This project is licensed under the GPL License - see the LICENSE file for details.

---

**MiniDB v1.0** - MPP Database Prototype with Vectorized Execution and Distributed Architecture Foundations