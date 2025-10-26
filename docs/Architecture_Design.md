# MiniDB Architecture Design Document

**Version:** 2.0 (Lakehouse Architecture)
**Last Updated:** 2025-10-06
**Authors:**  Yason Lee

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [System Overview](#2-system-overview)
3. [Architectural Layers](#3-architectural-layers)
4. [Core Components](#4-core-components)
5. [Execution Engines](#5-execution-engines)
6. [Storage Architecture](#6-storage-architecture)
7. [Transaction Management](#7-transaction-management)
8. [Query Processing Pipeline](#8-query-processing-pipeline)
9. [Concurrency Control](#9-concurrency-control)
10. [Performance Optimizations](#10-performance-optimizations)
11. [System Characteristics](#11-system-characteristics)
12. [Design Decisions](#12-design-decisions)
13. [Future Architecture Evolution](#13-future-architecture-evolution)

---

## 1. Executive Summary

MiniDB is a high-performance analytical database system built on **Lakehouse architecture** principles, combining the flexibility of data lakes with the performance and reliability of data warehouses. The system implements a dual-engine execution model, providing both traditional row-based processing and vectorized columnar execution for optimal query performance across diverse workloads.

### Key Architectural Highlights

- **Dual Execution Engines**: Automatic selection between regular and vectorized engines based on cost estimation
- **Lakehouse Storage**: Apache Parquet + Delta Log for ACID transactions over object storage
- **Cost-Based Optimization**: Statistics-driven query optimization with join reordering and predicate pushdown
- **Multi-Concurrency Models**: Both pessimistic (sync.RWMutex) and optimistic (PutIfNotExists) concurrency control
- **Merge-on-Read**: 1000x write amplification reduction for UPDATE/DELETE operations
- **Columnar Analytics**: Apache Arrow integration for 10-100x speedup on analytical queries

---

## 2. System Overview

### 2.1 Architecture Philosophy

MiniDB follows a **layered modular architecture** with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────────┐
│                    SQL Interface Layer                       │
│                  (TCP Server + Session Mgmt)                 │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                      Parser Layer                            │
│            (ANTLR4-based SQL Parser → AST)                   │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                    Optimizer Layer                           │
│      (Rule-Based + Cost-Based Query Optimization)            │
└─────────────────────────────────────────────────────────────┘
                              │
                    ┌─────────┴─────────┐
                    │                   │
┌───────────────────▼──────┐  ┌────────▼────────────────────┐
│  Regular Executor        │  │  Vectorized Executor        │
│  (Row-based Processing)  │  │  (Apache Arrow Batches)     │
└───────────────────────────┘  └─────────────────────────────┘
                    │                   │
                    └─────────┬─────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                   Storage Layer (Lakehouse)                  │
│  ┌───────────────┐  ┌──────────────┐  ┌──────────────────┐ │
│  │ Parquet Engine│  │  Delta Log   │  │  Object Store    │ │
│  │   (Columnar)  │  │  (ACID Txn)  │  │  (Local/Cloud)   │ │
│  └───────────────┘  └──────────────┘  └──────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 Design Principles

1. **Separation of Concerns**: Each layer has well-defined responsibilities
2. **Modularity**: Components can be replaced or extended independently
3. **Performance First**: Vectorization and zero-copy data transfers where possible
4. **ACID Guarantees**: Transaction log ensures data consistency
5. **Cloud-Native Ready**: Object storage abstraction for S3/Azure Blob compatibility
6. **Observability**: Structured logging (zap) with timing metrics at every layer

---

## 3. Architectural Layers

### 3.1 Network Layer

**Component**: `cmd/server/main.go`

**Responsibilities**:
- TCP server listening on configurable host:port (default: localhost:7205)
- Connection management with goroutine-per-connection model
- Session lifecycle management
- Protocol handling (SQL text over TCP)

**Implementation Details**:
```go
// Key structures
- listener: net.Listener        // TCP listener
- handler: *QueryHandler         // Global query processor
- sessionManager: *SessionManager // Session tracking
```

**Session Management** (`internal/session/session.go`):
- Snowflake ID generation for unique session IDs
- Session-scoped variables (current database, transaction state)
- Automatic session cleanup with configurable timeout
- Thread-safe session storage using `sync.Map`

---

### 3.2 Parser Layer

**Component**: `internal/parser/`

**Grammar**: ANTLR4-based (`MiniQL.g4`)

**Responsibilities**:
- Lexical analysis: SQL text → tokens
- Syntax analysis: tokens → parse tree
- Semantic analysis: parse tree → AST (Abstract Syntax Tree)

**Supported SQL Syntax**:
```sql
-- DDL
CREATE/DROP DATABASE
CREATE/DROP TABLE (INT, VARCHAR types)
CREATE/DROP INDEX (BTREE)

-- DML
INSERT INTO table VALUES (...)
UPDATE table SET ... WHERE ...
DELETE FROM table WHERE ...

-- DQL
SELECT [DISTINCT] columns FROM table
  [JOIN table ON condition]
  [WHERE condition]
  [GROUP BY columns [HAVING condition]]
  [ORDER BY columns [ASC|DESC]]
  [LIMIT n]

-- System Commands
SHOW DATABASES | TABLES | INDEXES
USE database
ANALYZE TABLE table [columns]
EXPLAIN query
```

**AST Node Types** (`internal/parser/ast.go`):
```go
- SelectStmt: SELECT queries with full clause support
- InsertStmt: INSERT operations
- UpdateStmt: UPDATE with assignments and filters
- DeleteStmt: DELETE with WHERE conditions
- CreateTableStmt: Table definitions
- JoinClause: JOIN operations with conditions
- WhereClause: Filter predicates
- GroupByClause: Aggregation grouping
- OrderByClause: Result sorting
```

**Parser Pipeline**:
1. **Lexer** (`miniql_lexer.go`): Tokenizes SQL string
2. **Parser** (`miniql_parser.go`): Generates parse tree
3. **Visitor** (`parser.go`): Transforms parse tree to AST nodes
4. **Error Handling**: ANTLR diagnostic error listeners

---

### 3.3 Optimizer Layer

**Component**: `internal/optimizer/`

**Optimizer Types**:

#### 3.3.1 Rule-Based Optimizer (`optimizer.go`)

Applies transformation rules in sequence:

```go
type Optimizer struct {
    rules []Rule  // Optimization rules
}

// Applied rules:
1. PredicatePushDownRule    // Push WHERE to data sources
2. JoinReorderRule          // Reorder joins by cardinality
3. ProjectionPruningRule    // Eliminate unused columns
```

**Rule Application Process**:
```
AST → buildPlan() → Initial Plan Tree
  ↓
Apply Rule 1: Predicate Pushdown
  ↓
Apply Rule 2: Join Reordering
  ↓
Apply Rule 3: Projection Pruning
  ↓
Optimized Plan Tree
```

#### 3.3.2 Cost-Based Optimizer (`cost_optimizer.go`)

Uses table statistics to make execution decisions:

```go
type CostBasedOptimizer struct {
    statsMgr *statistics.StatisticsManager
    config   *OptimizerConfig
}

// Cost factors
SeqScanCostFactor:    1.0
IndexScanCostFactor:  0.1
HashJoinCostFactor:   1.5
NestedLoopCostFactor: 2.0
```

**Cost Estimation**:
- **Table Scan**: `RowCount * SeqScanCostFactor`
- **Index Scan**: `RowCount * Selectivity * IndexScanCostFactor`
- **Nested Loop Join**: `LeftRows * RightRows * NLFactor`
- **Hash Join**: `(LeftRows + RightRows) * HashFactor`

**Statistics Used**:
- Table row counts
- Column cardinality (distinct values)
- Min/Max values for predicate pushdown
- Null counts

#### 3.3.3 Plan Tree Structure

```go
type Plan struct {
    Type       PlanType           // SelectPlan, FilterPlan, etc.
    Properties interface{}        // Type-specific properties
    Children   []*Plan            // Child plan nodes
}

// Plan types
const (
    SelectPlan        // Projection
    TableScanPlan     // Table scan
    FilterPlan        // WHERE clause
    JoinPlan          // JOIN operation
    GroupPlan         // GROUP BY
    OrderPlan         // ORDER BY
    LimitPlan         // LIMIT
    // ... DDL/DML plans
)
```

---

## 4. Core Components

### 4.1 Catalog System

**Component**: `internal/catalog/`

**Architecture**:

```go
type Catalog struct {
    *SimpleSQLCatalog
}

type SimpleSQLCatalog struct {
    storageEngine storage.Engine
    sqlRunner     SQLRunner        // For SQL-based catalog queries
    databases     sync.Map         // Database metadata cache
    tables        sync.Map         // Table metadata cache
    indexes       sync.Map         // Index metadata cache
}
```

**Metadata Storage**:

MiniDB v2.0 uses **SQL Bootstrap** approach:
- System catalog tables stored in `sys.*` tables
- Catalog queries use regular SQL execution path
- Metadata persisted in Parquet format with Delta Log

**Key System Tables**:
```sql
sys.databases           -- Database list
sys.tables              -- Table schemas
sys.indexes             -- Index definitions
sys.table_statistics    -- Table-level stats
sys.column_statistics   -- Column-level stats
sys.delta_log           -- Transaction log (SQL queryable!)
```

**Table Metadata** (`catalog.TableMeta`):
```go
type TableMeta struct {
    Database   string          // Database name
    Table      string          // Table name
    ChunkCount int64           // Number of data chunks
    Schema     *arrow.Schema   // Arrow schema definition
}
```

**Index Metadata** (`catalog.IndexMeta`):
```go
type IndexMeta struct {
    Database  string     // Owner database
    Table     string     // Indexed table
    Name      string     // Index name
    Columns   []string   // Indexed columns
    IsUnique  bool       // Uniqueness constraint
    IndexType string     // BTREE, HASH, etc.
}
```

---

### 4.2 Statistics Manager

**Component**: `internal/statistics/`

**Purpose**: Collect and maintain table/column statistics for cost-based optimization

**Statistics Collection**:

```go
type StatisticsManager struct {
    catalog      *catalog.Catalog
    tableStats   sync.Map  // table_id → TableStatistics
    columnStats  sync.Map  // table.column → ColumnStatistics
}

type TableStatistics struct {
    TableID    string
    RowCount   int64
    FileCount  int64
    TotalSize  int64
    Version    int64
}

type ColumnStatistics struct {
    TableID       string
    ColumnName    string
    DataType      string
    MinValue      interface{}
    MaxValue      interface{}
    NullCount     int64
    DistinctCount int64
}
```

**Collection Trigger**:
```sql
ANALYZE TABLE products;              -- All columns
ANALYZE TABLE products (id, name);   -- Specific columns
```

**Collection Process**:
1. **Scan Table**: Full table scan via Parquet engine
2. **Compute Stats**: Min/max/null/distinct for each column
3. **Persist**: Write to `sys.table_statistics` and `sys.column_statistics`
4. **Cache**: Update in-memory statistics cache

---

## 5. Execution Engines

### 5.1 Engine Selection Strategy

**Decision Point**: `internal/executor/cost_optimizer.go:ShouldUseVectorizedExecution()`

**Vectorized Engine Used When**:
- Query has aggregations (SUM, COUNT, AVG, MIN, MAX)
- Query has GROUP BY clause
- Table has sufficient statistics
- No complex WHERE predicates (e.g., LIKE, IN)

**Regular Engine Used When**:
- Complex predicates requiring row-by-row evaluation
- Transaction operations (UPDATE, DELETE with complex conditions)
- DDL operations

### 5.2 Regular Executor

**Component**: `internal/executor/executor.go`

**Architecture**:

```go
type ExecutorImpl struct {
    catalog     *catalog.Catalog
    dataManager *DataManager
}

// Operator-based execution model
type Operator interface {
    Init(ctx interface{}) error
    Next() (*types.Batch, error)
    Close() error
}
```

**Available Operators** (`internal/executor/operators/`):
- **TableScan**: Read data from storage
- **Filter**: Apply WHERE predicates
- **Projection**: Column selection
- **Join**: Nested loop join
- **GroupBy**: Aggregation with grouping
- **OrderBy**: Result sorting
- **Limit**: Result set limiting

**Execution Flow**:
```bash
1. buildOperator(plan) → Operator Tree
2. op.Init(ctx)        → Initialize resources
3. Loop:
     batch := op.Next() → Pull next batch
     if batch == nil: break
     collect(batch)
4. op.Close()          → Cleanup resources
5. Return ResultSet
```

**Example Operator Tree**:
```bash
SELECT name, COUNT(*) FROM users WHERE age > 18 GROUP BY name

Operator Tree:
  GroupBy(name, COUNT(*))
    ↓
  Filter(age > 18)
    ↓
  TableScan(users)
```

---

### 5.3 Vectorized Executor

**Component**: `internal/executor/vectorized_executor.go`

**Architecture**:

```go
type VectorizedExecutor struct {
    catalog     *catalog.Catalog
    dataManager *DataManager
    statsMgr    *statistics.StatisticsManager
    optimizer   *CostBasedOptimizer
}

type VectorizedPipeline struct {
    operations []types.VectorizedOperation
    schema     *arrow.Schema
}
```

**Vectorized Operations** (`internal/types/vectorized.go`):
```go
type VectorizedOperation interface {
    Execute(input *VectorizedBatch) (*VectorizedBatch, error)
    Name() string
}

// Available operations
- FilterOperation     // Vectorized filtering using Arrow compute
- ProjectOperation    // Column projection
- AggregateOperation  // SIMD aggregations
```

**Vectorized Batch Processing**:

```go
type VectorizedBatch struct {
    schema  *arrow.Schema
    columns []arrow.Array  // Columnar data
    rowCount int64
}

// Benefits of vectorization:
1. SIMD instructions for parallel computation
2. Better CPU cache utilization
3. Reduced function call overhead
4. Efficient memory access patterns
```

**Performance Characteristics**:
- **GROUP BY + Aggregation**: 10-100x faster than row-based
- **Large Table Scans**: 5-20x faster due to columnar layout
- **Simple Filters**: 3-10x faster with vectorized predicates

**Apache Arrow Integration**:
```go
// Data conversion pipeline
Parquet File → arrow.Record → VectorizedBatch → Processing → Results

// Zero-copy where possible
- Direct memory references to Arrow buffers
- No serialization/deserialization overhead
- Efficient inter-process communication (IPC)
```

---

## 6. Storage Architecture

### 6.1 Lakehouse Storage Model

**Component**: `internal/storage/parquet_engine.go`

**Architecture Layers**:

```
┌──────────────────────────────────────────────────┐
│          ParquetEngine (Unified API)             │
└──────────────────────────────────────────────────┘
         │                    │
         │                    │
    ┌────▼──────┐      ┌──────▼─────────┐
    │   Delta   │      │    Object      │
    │    Log    │      │     Store      │
    │ (ACID Tx) │      │  (Files/S3)    │
    └───────────┘      └────────────────┘
         │                    │
         └──────────┬─────────┘
                    │
            ┌───────▼────────┐
            │     Parquet    │
            │      Files     │
            │   (Columnar)   │
            └────────────────┘
```

**Parquet Engine**:

```go
type ParquetEngine struct {
    basePath           string
    objectStore        ObjectStore
    deltaLog           delta.LogInterface
    schemas            map[string]*arrow.Schema
    mu                 sync.RWMutex
    useOptimisticLock  bool
    maxRetries         int
}

// Core operations
- Write(ctx, db, table, record) → Append data
- Scan(ctx, db, table, filters) → Read with predicates
- Update(ctx, db, table, assignments, filters) → MoR update
- Delete(ctx, db, table, filters) → MoR delete
- Compact(tableID) → Merge small files
```

**File Organization**:
```bash
minidb_data/
├── database1/
│   ├── table1/
│   │   ├── data/
│   │   │   ├── part-00001-{uuid}.parquet  (1GB target)
│   │   │   ├── part-00002-{uuid}.parquet
│   │   │   └── delta-{uuid}.parquet       (MoR deltas)
│   │   └── _delta_log/
│   │       ├── 000001.json  (ADD part-00001)
│   │       ├── 000002.json  (REMOVE old, ADD new)
│   │       └── 000010.checkpoint.parquet  (Checkpoint)
└── sys/
    ├── delta_log/           (System transaction log)
    └── table_statistics/    (Statistics data)
```

---

### 6.2 Delta Log

**Purpose**: Provide ACID transaction guarantees over object storage

**Components**:

#### 6.2.1 Pessimistic Delta Log

**File**: `internal/delta/log.go`

**Concurrency Model**: Global lock (`sync.RWMutex`)

```go
type DeltaLog struct {
    entries             []LogEntry
    mu                  sync.RWMutex  // Serializes all commits
    currentVer          atomic.Int64
    tableName           string
    persistenceCallback PersistenceCallback
    checkpointCallback  CheckpointCallback
}

// Transaction workflow
1. Lock: dl.mu.Lock()
2. Append: entries = append(entries, newEntry)
3. Persist: callback.Persist(entry)
4. Unlock: dl.mu.Unlock()
```

**Advantages**:
- Simple to reason about
- Strong consistency guarantees
- No retry logic needed

**Disadvantages**:
- Lower write concurrency
- Lock contention on high write workloads

#### 6.2.2 Optimistic Delta Log

**File**: `internal/delta/optimistic_log.go`

**Concurrency Model**: Atomic version increment + Conditional PUT

```go
type OptimisticDeltaLog struct {
    objectStore objectstore.ConditionalObjectStore
    basePath    string
    currentVer  atomic.Int64  // Lock-free version counter
}

// Transaction workflow
1. version = currentVer.Add(1)           // Atomic increment
2. entry = createLogEntry(version)
3. err = store.PutIfNotExists(path, entry)
4. if conflict:
     currentVer.Add(-1)                  // Rollback
     return RetryableConflictError
```

**Advantages**:
- High write concurrency
- Lock-free design
- Scales with concurrent writers

**Disadvantages**:
- Requires retry logic
- Potential contention on version file creation

**Conflict Resolution**:
```go
// Client-side retry with exponential backoff
maxRetries := 3
for attempt := 0; attempt < maxRetries; attempt++ {
    err = deltaLog.AppendAdd(tableID, file)
    if err == nil {
        break  // Success
    }
    if isConflictError(err) {
        backoff := time.Duration(math.Pow(2, attempt)) * 100ms
        time.Sleep(backoff)
        continue  // Retry
    }
    return err  // Non-retryable error
}
```

---

### 6.3 Log Entry Format

```go
type LogEntry struct {
    Version    int64             // Transaction version
    Timestamp  int64             // Commit timestamp (milliseconds)
    TableID    string            // "database.table"
    Operation  string            // "ADD" or "REMOVE"

    // File metadata
    FilePath   string            // Relative path to data file
    FileSize   int64             // File size in bytes
    RowCount   int64             // Number of rows

    // Statistics for predicate pushdown
    MinValues  map[string]interface{}  // Column min values
    MaxValues  map[string]interface{}  // Column max values
    NullCounts map[string]int64        // Null counts

    // Merge-on-Read metadata
    IsDelta    bool              // Is this a delta file?
    DeltaType  string            // "UPDATE" or "DELETE"
    DataChange bool              // Does this change data?
}
```

**Serialization**: JSON format for human readability

**Example Log Entry**:
```json
{
  "version": 42,
  "timestamp": 1698765432000,
  "table_id": "ecommerce.orders",
  "operation": "ADD",
  "file_path": "ecommerce/orders/data/part-00042-abc123.parquet",
  "file_size": 1073741824,
  "row_count": 1000000,
  "min_values": {"order_id": 1, "amount": 10.50},
  "max_values": {"order_id": 1000000, "amount": 9999.99},
  "null_counts": {"shipping_address": 0},
  "is_delta": false,
  "data_change": true
}
```

---

### 6.4 Checkpoint Mechanism

**Trigger**: Every 10 transactions

**Purpose**:
- Compress transaction log history
- Accelerate recovery by avoiding full log replay
- Provide snapshot of table state

**Checkpoint File Format**: Apache Parquet (same as data files)

**Checkpoint Content**:
```go
type CheckpointData struct {
    Version      int64                    // Checkpoint version
    TableID      string                   // Target table
    ActiveFiles  []ParquetFile            // Current valid files
    Statistics   map[string]Statistics    // Aggregated stats
    Schema       *arrow.Schema            // Table schema
}
```

**Recovery Process**:
```bash
1. Find latest checkpoint: _delta_log/000100.checkpoint.parquet
2. Read checkpoint → Get file list at V100
3. Replay log entries V101..V150 → Get current state
4. Build table snapshot → Ready for queries
```

**Checkpoint vs Full Log Replay**:
- **Without Checkpoint**: Read 1000 JSON files (100ms)
- **With Checkpoint**: Read 1 Parquet file (10ms) + 10 JSON files (1ms) = 11ms
- **Speedup**: ~9x faster recovery

---

### 6.5 Merge-on-Read (MoR)

**Problem**: Traditional Copy-on-Write has high write amplification
- UPDATE 1 row in 1GB file → Rewrite entire 1GB file
- DELETE 1 row → Rewrite entire file
- Write amplification: 1GB / 1KB = **1,000,000x**

**MoR Solution**: Write deltas, merge on read

**Delta File Types**:
1. **DELETE Delta**: Contains row IDs to delete
2. **UPDATE Delta**: Contains (old_row_id, new_row_data) pairs

**Update Workflow**:
```go
// UPDATE users SET name='Alice' WHERE id=42

1. Scan table, find matching rows
2. Create delta file with updates:
   delta-update-{uuid}.parquet
   Content: [{row_id: 42, name: 'Alice'}]
3. Append to Delta Log:
   {operation: "ADD", is_delta: true, delta_type: "UPDATE"}
4. Done (no base file rewrite!)
```

**Read Workflow with MoR**:
```go
// SELECT * FROM users WHERE id=42

1. Get snapshot from Delta Log → List of files
2. Identify base files + delta files
3. Scan base file → Get row
4. Apply delta files → Merge updates/deletes
5. Return merged result
```

**Write Amplification Comparison**:
- **Copy-on-Write**: 1GB file → 1GB write (1,000,000x)
- **Merge-on-Read**: 1KB delta → 1KB write (1x)
- **Improvement**: **1000x reduction**

**Compaction**:
Periodically merge deltas into base files to maintain read performance:
```bash
Trigger: 10+ delta files or 10MB delta size
Process: Merge deltas + base → New base file
Result: Improved read performance, cleaned history
```

---

### 6.6 Predicate Pushdown

**Feature**: Evaluate filters at storage layer using statistics

**Statistics Used**:

- **Min/Max Values**: Skip files outside range
- **Null Counts**: Skip files for IS NOT NULL predicates

**Example**:
```sql
SELECT * FROM orders WHERE order_date = '2024-01-15'
```

**Without Pushdown**:
```bash
1. Read all 100 Parquet files (100GB)
2. Filter in executor
3. Return 100MB result
Total I/O: 100GB
```

**With Pushdown**:
```bash
1. Check Delta Log statistics:
   - File 1: min='2024-01-01', max='2024-01-10' → SKIP
   - File 2: min='2024-01-11', max='2024-01-20' → READ
   - File 3: min='2024-01-21', max='2024-01-31' → SKIP
2. Read only File 2 (1GB)
3. Filter and return result
Total I/O: 1GB
Speedup: 100x
```

**Implementation**:
```go
func (pe *ParquetEngine) Scan(
    ctx context.Context,
    dbName, tableName string,
    filters []Filter,
) (*RecordIterator, error) {
    snapshot := pe.deltaLog.GetSnapshot()

    // Apply predicate pushdown
    candidateFiles := []ParquetFile{}
    for _, file := range snapshot.Files {
        if canSkipFile(file.Stats, filters) {
            continue  // Skip based on statistics
        }
        candidateFiles = append(candidateFiles, file)
    }

    // Read only candidate files
    return pe.scanFiles(candidateFiles, filters)
}
```

---

## 7. Transaction Management

### 7.1 ACID Properties

**Atomicity**:
- All-or-nothing commits via Delta Log
- Version increments are atomic (atomic.Int64)

**Consistency**:
- Schema validation before writes
- Foreign key constraints (planned)

**Isolation**:
- **Snapshot Isolation**: Readers see consistent snapshot
- Reads at version V see all commits ≤ V
- No dirty reads, no phantom reads

**Durability**:
- Fsync after each Delta Log write
- Parquet files synced to disk/object storage
- Recovery from checkpoints + log replay

### 7.2 Transaction Lifecycle

```bash
BEGIN TRANSACTION (implicit)
  ↓
Execute SQL Statement
  ↓
Validation (schema, constraints)
  ↓
Write to Storage
  ↓
Append to Delta Log (version N+1)
  ↓
Fsync Delta Log
  ↓
COMMIT (implicit)
```

### 7.3 Snapshot Isolation

**Version-Based Reads**:
```go
// Reader gets snapshot at version V
snapshot := deltaLog.GetSnapshotAt(version)
files := snapshot.Files  // Immutable file list

// Writer commits new version V+1
deltaLog.AppendAdd(tableID, newFile)  // V+1

// Reader's snapshot unchanged
// No locks, no blocking
```

**Time Travel Queries** (Planned):
```sql
-- Query table as of specific version
SELECT * FROM orders VERSION AS OF 42

-- Query table as of timestamp
SELECT * FROM orders TIMESTAMP AS OF '2024-01-15 10:00:00'
```

---

## 8. Query Processing Pipeline

### 8.1 End-to-End Query Flow

```bash
SQL Text
  ↓
┌─────────────────────┐
│  1. Parser          │  ANTLR4 lexer/parser
│  "SELECT ..." →     │  Visitor pattern
│  AST                │  Syntax validation
└─────────────────────┘
  ↓
┌─────────────────────┐
│  2. Semantic        │  Resolve table/column names
│     Analysis        │  Type checking
│                     │  Alias resolution
└─────────────────────┘
  ↓
┌─────────────────────┐
│  3. Rule-Based      │  Predicate pushdown
│     Optimizer       │  Join reordering
│                     │  Projection pruning
└─────────────────────┘
  ↓
┌─────────────────────┐
│  4. Cost-Based      │  Collect statistics
│     Optimizer       │  Estimate costs
│                     │  Choose join algorithm
└─────────────────────┘
  ↓
┌─────────────────────┐
│  5. Executor        │  Regular or Vectorized?
│     Selection       │  Based on cost estimation
└─────────────────────┘
  ↓
┌──────────┬──────────┐
│ Regular  │Vectorized│
│ Executor │ Executor │
└──────────┴──────────┘
  ↓           ↓
┌─────────────────────┐
│  6. Storage Access  │  Read Parquet files
│                     │  Apply filters
│                     │  Merge MoR deltas
└─────────────────────┘
  ↓
┌─────────────────────┐
│  7. Result Set      │  Format output
│                     │  Return to client
└─────────────────────┘
```

### 8.2 Detailed Example

**Query**:
```sql
SELECT category, AVG(price) as avg_price
FROM products
WHERE stock > 0
GROUP BY category
```

**Step 1: Parser**
```go
AST: SelectStmt {
    Columns: [
        {Column: "category"},
        {Expr: FunctionCall("AVG", "price"), Alias: "avg_price"}
    ],
    From: "products",
    Where: BinaryExpr{Left: "stock", Op: ">", Right: 0},
    GroupBy: ["category"]
}
```

**Step 2: Optimizer - Build Initial Plan**
```bash
SelectPlan(category, AVG(price))
  ↓
GroupPlan(category, AVG(price))
  ↓
FilterPlan(stock > 0)
  ↓
TableScanPlan(products)
```

**Step 3: Optimizer - Apply Rules**
- **Predicate Pushdown**: Move `stock > 0` to TableScan properties
- **Projection Pruning**: Only read columns: category, price, stock

**Optimized Plan**:
```go
SelectPlan(category, avg_price)
  ↓
GroupPlan(category, AVG(price))
  ↓
TableScanPlan(products, filters=[stock > 0], columns=[category, price, stock])
```

**Step 4: Cost-Based Decisions**
```bash
- products table: 1M rows, 500MB
- Statistics: stock values [0..1000]
- Selectivity estimate: ~90% (stock > 0)
- Has aggregation → Use Vectorized Executor
```

**Step 5: Vectorized Execution**
```go
// Build vectorized pipeline
pipeline = [
    TableScanOperation(products),
    FilterOperation(stock > 0),
    AggregateOperation(GroupBy=category, Agg=AVG(price))
]

// Execute in batches (default: 10K rows per batch)
for batch := range pipeline.Execute() {
    results = append(results, batch)
}
```

**Step 6: Storage Access**
```bash
1. Get snapshot from Delta Log
2. List files: [part-00001.parquet, ..., part-00100.parquet]
3. Check statistics → All files may have stock > 0
4. Scan all files with filter
5. Apply MoR deltas (if any)
6. Return Arrow Record batches
```

**Step 7: Result**
```bash
category    | avg_price
------------|----------
Electronics | 599.99
Books       | 29.99
Clothing    | 49.99
```

---

## 9. Concurrency Control

### 9.1 Read-Write Concurrency

**Readers Never Block Writers**:
- Readers operate on snapshots
- Writers create new versions
- No lock contention between reads/writes

**Writers May Block Each Other**:
- **Pessimistic Mode**: Global lock serializes writes
- **Optimistic Mode**: Writers retry on conflict

### 9.2 Concurrency Model Comparison

| Aspect | Pessimistic (sync.RWMutex) | Optimistic (PutIfNotExists) |
|--------|----------------------------|------------------------------|
| **Write Concurrency** | Low (serialized) | High (parallel) |
| **Read Concurrency** | High (snapshot) | High (snapshot) |
| **Conflict Handling** | Block & wait | Detect & retry |
| **Latency** | Predictable | Variable (retries) |
| **Throughput** | Lower ceiling | Higher ceiling |
| **Best For** | OLTP, low contention | OLAP, high concurrency |

### 9.3 Lock-Free Data Structures

**Atomic Operations**:
```go
// Version counter
currentVer atomic.Int64
version := currentVer.Add(1)  // Thread-safe increment

// Session storage
sessions sync.Map               // Lock-free concurrent map
sessions.Store(key, value)
sessions.Load(key)
```

---

## 10. Performance Optimizations

### 10.1 Z-Order Clustering

**Component**: `internal/optimizer/zorder.go`

**Purpose**: Multi-dimensional data clustering for query acceleration

**Algorithm**:
```go
// Interleave bits from multiple dimensions
func computeZValue(record arrow.Record, columns []string) uint64 {
    var zValue uint64

    // 1. Normalize each dimension to 21-bit space
    dimValues := []uint64{}
    for _, col := range columns {
        rawValue := getColumnValue(record, col)
        normalized := normalize(rawValue, 0, 1<<21-1)
        dimValues = append(dimValues, normalized)
    }

    // 2. Bit interleaving (Z-order curve)
    for bitPos := 0; bitPos < 21; bitPos++ {
        for dimIdx, dimValue := range dimValues {
            bit := (dimValue >> bitPos) & 1
            zValue |= bit << (bitPos*len(columns) + dimIdx)
        }
    }

    return zValue  // 63-bit Z-order value
}
```

**Benefits**:
- Multi-dimensional range queries: 5-10x faster
- Better data locality for multi-column filters
- Example: `WHERE state='CA' AND age > 18` on Z-ordered (state, age)

**Space-Filling Curve**:
```
2D Example (x, y coordinates):
Linear:  (0,0),(0,1),(0,2),...(0,N),(1,0),(1,1),...
Z-Order: (0,0),(1,0),(0,1),(1,1),(2,0),(3,0),(2,1),(3,1),...

Z-Order preserves spatial locality!
```

### 10.2 Compaction

**Component**: `internal/optimizer/compaction.go`

**Trigger Conditions**:
- 10+ small files (<10MB each)
- Total delta file size > 100MB
- Manual trigger: `OPTIMIZE TABLE products`

**Compaction Process**:
```go
type CompactionConfig struct {
    TargetFileSize    int64  // 1GB optimal
    MinFileSize       int64  // 10MB threshold
    MaxFilesToCompact int    // Max 100 files per run
}

func Compact(tableID string) {
    // 1. Identify small files
    smallFiles := identifySmallFiles(snapshot.Files)

    // 2. Read and merge
    mergedData := mergeFiles(smallFiles)

    // 3. Write new file
    newFile := writeParquet(mergedData, targetSize=1GB)

    // 4. Atomic Delta Log update
    deltaLog.AppendRemove(tableID, smallFiles...)
    deltaLog.AppendAdd(tableID, newFile)
}
```

**Before Compaction**:
```bash
Table: orders
Files: 100 files × 10MB = 1GB
Delta Log: 100 entries
Read latency: 100 × 5ms = 500ms
```

**After Compaction**:
```bash
Table: orders
Files: 1 file × 1GB = 1GB
Delta Log: 1 entry
Read latency: 1 × 20ms = 20ms
Speedup: 25x
```

### 10.3 Data Skipping

**Techniques**:

1. **File-Level Skipping**: Min/Max statistics
2. **Row Group Skipping**: Parquet row group statistics
3. **Page-Level Skipping**: Parquet page index (planned)

**Example Performance**:
```sql
-- Query: SELECT * FROM logs WHERE timestamp = '2024-01-15'
-- Table: 365 files (one per day), 1TB total

Statistics Check:
- 364 files skipped (wrong date range)
- 1 file scanned (2.7GB)

I/O Reduction: 1TB → 2.7GB (370x less)
Query Time: 300s → 0.8s
```

---

## 11. System Characteristics

### 11.1 Performance Profile

**OLAP Workloads** (Analytical):
- Aggregations: ⚡ **Excellent** (10-100x speedup with vectorization)
- Large scans: ⚡ **Excellent** (columnar layout, predicate pushdown)
- Complex JOINs: ✅ **Good** (hash join for large tables)

**OLTP Workloads** (Transactional):
- Point queries: ✅ **Good** (with indexes)
- Small updates: ✅ **Good** (MoR, low latency)
- High concurrency writes: ⚠️ **Moderate** (depends on lock mode)

### 11.2 Scalability

**Vertical Scaling**:
- Memory: Linear scaling with data size
- CPU: Near-linear scaling with vectorization
- Storage: Tested up to 100GB+ on single node

**Horizontal Scaling** (Planned):
- Distributed query execution (MPP architecture)
- Data partitioning across nodes
- Coordinated transactions

### 11.3 Resource Usage

**Memory**:
- Base overhead: ~50MB
- Per-session: ~1MB
- Query execution: Configurable batch size (default: 10K rows)
- Arrow zero-copy: Minimal duplication

**CPU**:
- Vectorized operations: 80-90% CPU utilization (good!)
- Regular operations: 40-60% CPU utilization
- SIMD instructions: Auto-vectorization by compiler

**Disk I/O**:
- Sequential reads: Optimized (Parquet columnar)
- Random writes: Minimized (append-only Delta Log)
- Compression: Snappy (default), up to 5x reduction

---

## 12. Design Decisions

### 12.1 Why Dual Execution Engines?

**Rationale**:
- **Vectorized**: 10-100x faster for aggregations, but limited predicate support
- **Regular**: Universal predicate support, but slower for analytics
- **Solution**: Automatic selection based on query characteristics

**Trade-off**: Increased code complexity vs. optimal performance

### 12.2 Why Parquet + Arrow?

**Parquet (Storage)**:
- Columnar format: Read only needed columns
- Compression: 2-10x space savings
- Statistics: Built-in min/max for pushdown
- Industry standard: Ecosystem compatibility

**Arrow (In-Memory)**:
- Zero-copy transfers: Minimal serialization
- SIMD-friendly: Vectorized operations
- Cross-language: Share data with Python/R
- IPC: Efficient inter-process communication

### 12.3 Why Delta Log?

**Problem**: Object storage has no transactions

**Solution**: Transaction log provides:
- ACID guarantees
- Version history (time travel)
- Metadata separation (small files)
- Cloud-native (works with S3/Azure)

**Alternative Considered**: PostgreSQL as metadata store
- **Rejected**: Additional dependency, not cloud-native

### 12.4 Why MoR over CoW?

**Copy-on-Write** (Traditional):
- UPDATE 1 byte → Rewrite entire file
- High write latency
- High write amplification

**Merge-on-Read** (MiniDB):
- UPDATE 1 byte → Write small delta file
- Low write latency (1000x improvement)
- Slightly slower reads (mitigated by compaction)

**Trade-off**: Optimized for write-heavy analytical workloads

### 12.5 Why Go Language?

**Advantages**:
- High performance (compiled, static typing)
- Excellent concurrency (goroutines, channels)
- Memory safety (GC, no segfaults)
- Fast compilation (rapid iteration)
- Rich ecosystem (Arrow, Parquet libraries)

**Disadvantages**:
- GC pauses (mitigated by tuning)
- No SIMD intrinsics (rely on compiler auto-vectorization)

---

## 13. Future Architecture Evolution

### 13.1 Short-Term Roadmap (3-6 months)

**Distributed Query Execution**:
- Implement MPP architecture
- Worker node coordination
- Query fragment distribution

**Advanced Indexing**:
- B-Tree index integration
- Bloom filters for existence checks
- Zone maps for min/max filtering

**SQL Enhancement**:
- Window functions (ROW_NUMBER, RANK, LAG/LEAD)
- Common Table Expressions (CTEs)
- Subqueries (correlated and uncorrelated)

### 13.2 Mid-Term Roadmap (6-12 months)

**Cloud Storage Integration**:
- S3 backend (AWS)
- Azure Blob Storage backend
- GCS backend (Google Cloud)

**Query Optimization**:
- Adaptive query execution
- Runtime statistics feedback
- Query result caching

**Advanced ACID**:
- Multi-statement transactions
- Savepoints
- Isolation level configuration

### 13.3 Long-Term Vision (12+ months)

**Machine Learning Integration**:
- In-database ML model training
- Vectorized scoring
- GPU acceleration

**Advanced Analytics**:
- Graph query support
- Spatial data types and queries
- Time-series optimizations

**Enterprise Features**:
- Role-based access control (RBAC)
- Audit logging
- Data encryption at rest
- Backup and restore

---

## Appendix A: Key File Locations

```
minidb/
├── cmd/
│   └── server/
│       ├── main.go              # TCP server entry point
│       └── handler.go           # Query handler
├── internal/
│   ├── parser/
│   │   ├── MiniQL.g4            # ANTLR grammar
│   │   ├── parser.go            # Parser implementation
│   │   └── ast.go               # AST node definitions
│   ├── optimizer/
│   │   ├── optimizer.go         # Rule-based optimizer
│   │   ├── plan.go              # Plan tree structures
│   │   ├── predicate_push_down_rule.go
│   │   ├── join_reorder_rule.go
│   │   └── projection_pruning_rule.go
│   ├── executor/
│   │   ├── executor.go          # Regular executor
│   │   ├── vectorized_executor.go  # Vectorized executor
│   │   ├── cost_optimizer.go    # Cost-based optimizer
│   │   ├── data_manager.go      # Data access layer
│   │   └── operators/
│   │       ├── table_scan.go
│   │       ├── filter.go
│   │       ├── join.go
│   │       └── group_by.go
│   ├── storage/
│   │   └── parquet_engine.go    # Lakehouse storage engine
│   ├── delta/
│   │   ├── log.go               # Pessimistic Delta Log
│   │   ├── optimistic_log.go    # Optimistic Delta Log
│   │   └── types.go             # Delta Log types
│   ├── objectstore/
│   │   └── local.go             # Local file system store
│   ├── catalog/
│   │   ├── catalog.go           # Metadata catalog
│   │   └── simple_sql_catalog.go
│   ├── statistics/
│   │   └── statistics.go        # Statistics manager
│   ├── session/
│   │   └── session.go           # Session management
│   ├── logger/
│   │   └── logger.go            # Structured logging
│   └── types/
│       ├── batch.go             # Batch types
│       └── vectorized.go        # Vectorized operations
└── test/
    ├── executor_test.go
    ├── optimizer_test.go
    ├── delta_acid_test.go
    └── ...
```

---

## Appendix B: Configuration Parameters

**Optimizer Configuration**:
```go
SeqScanCostFactor:    1.0      // Sequential scan cost
IndexScanCostFactor:  0.1      // Index scan cost (10x cheaper)
HashJoinCostFactor:   1.5      // Hash join cost
NestedLoopCostFactor: 2.0      // Nested loop cost
WorkMemSize:          64MB     // Memory for sorts/joins
```

**Execution Configuration**:
```go
BatchSize:            10000    // Rows per batch
MaxRetries:           3        // Optimistic lock retries
CheckpointInterval:   10       // Checkpoint every N commits
```

**Storage Configuration**:
```go
TargetFileSize:       1GB      // Target Parquet file size
MinFileSize:          10MB     // Compaction threshold
CompressionCodec:     "snappy" // Compression algorithm
```

**Logging Configuration** (Environment Variables):
```bash
ENVIRONMENT=production   # production | development | test
LOG_LEVEL=info          # debug | info | warn | error
LOG_FILE=logs/minidb.log
LOG_MAX_SIZE=100        # MB
LOG_MAX_BACKUPS=3
LOG_MAX_AGE=30          # days
```

---

## Appendix C: Performance Benchmarks

**Hardware**: MacBook Pro M1, 16GB RAM, SSD

**Dataset**: 10M rows, 1GB data

**Query**: `SELECT category, COUNT(*), AVG(price) FROM products GROUP BY category`

| Executor | Time | Throughput |
|----------|------|------------|
| Regular | 15.2s | 658K rows/s |
| Vectorized | 1.8s | 5.5M rows/s |
| **Speedup** | **8.4x** | **8.4x** |

**Predicate Pushdown**:

Query: `SELECT * FROM logs WHERE date='2024-01-15'`

| Approach | Files Scanned | I/O | Time |
|----------|---------------|-----|------|
| No Pushdown | 365 files | 100GB | 250s |
| With Pushdown | 1 file | 274MB | 0.8s |
| **Improvement** | **365x** | **365x** | **312x** |

---

## Glossary

- **ACID**: Atomicity, Consistency, Isolation, Durability
- **AST**: Abstract Syntax Tree
- **CoW**: Copy-on-Write
- **MoR**: Merge-on-Read
- **MPP**: Massively Parallel Processing
- **MVCC**: Multi-Version Concurrency Control
- **OLAP**: Online Analytical Processing
- **OLTP**: Online Transaction Processing
- **SIMD**: Single Instruction, Multiple Data
- **SSTable**: Sorted String Table
- **WAL**: Write-Ahead Log

---

**Document Version**: 1.0
**Created**: 2025-10-06
**For**: MiniDB v2.0 (Lakehouse Architecture)

*This document represents the architectural design of MiniDB as implemented. For the latest updates, refer to the project repository and code comments.*
