# MiniDB Lakehouse v2.0 测试文档

本文档描述了 MiniDB v2.0 Lakehouse 架构的完整测试覆盖。

## 测试概览

MiniDB v2.0 包含五大类 Lakehouse 特性测试：

1. **Delta Lake ACID 事务测试** - 验证 ACID 属性
2. **时间旅行查询测试** - 验证版本控制和快照隔离
3. **谓词下推性能测试** - 验证查询优化和数据跳过
4. **Parquet 统计信息测试** - 验证完整的统计信息收集
5. **Arrow IPC 序列化测试** - 验证 schema 序列化的正确性

## 快速开始

### 运行所有 Lakehouse 测试

```bash
cd test
./run_lakehouse_tests.sh
```

### 运行特定测试类别

```bash
# Delta Lake ACID 测试
go test -v ./test/delta_acid_test.go -timeout 3m

# 时间旅行测试
go test -v ./test/time_travel_test.go -timeout 3m

# 谓词下推测试
go test -v ./test/predicate_pushdown_test.go -timeout 3m

# Parquet 统计信息测试
go test -v ./test/parquet_statistics_test.go -timeout 2m

# Arrow IPC 序列化测试
go test -v ./test/arrow_ipc_test.go -timeout 2m
```

## 详细测试说明

### 1. Delta Lake ACID 事务测试

**文件**: `delta_acid_test.go`

**测试内容**:
- **Atomicity (原子性)**: 验证所有操作要么全部成功，要么全部失败
- **Consistency (一致性)**: 验证数据完整性始终保持
- **Isolation (隔离性)**: 验证并发操作的快照隔离
- **Durability (持久性)**: 验证数据持久化到 Parquet 文件
- **Version Control (版本控制)**: 验证 Delta Log 版本跟踪
- **Snapshot Isolation (快照隔离)**: 验证一致性快照视图

**关键验证点**:
- 多个 INSERT 操作的原子性
- 数据一致性在多次操作后保持
- 不同会话的隔离性
- 数据立即可读性（持久性）
- Delta Log 版本递增
- 快照一致性

### 2. 时间旅行查询测试

**文件**: `time_travel_test.go`

**测试内容**:
- **Version-Based Time Travel**: 基于版本号的时间旅行查询
- **Snapshot Isolation**: 快照隔离验证
- **Delta Log Version Tracking**: Delta Log 版本跟踪
- **Snapshot Retrieval**: 快照检索和文件列表
- **File Tracking**: ADD 和 REMOVE 操作跟踪
- **Timestamp-Based Queries**: 基于时间戳的查询

**关键验证点**:
- 能够查询历史版本的数据
- 每个版本都有独立的快照
- 版本号单调递增
- 快照包含正确的文件列表
- ADD/REMOVE 操作正确记录
- 根据时间戳查找版本

### 3. 谓词下推性能测试

**文件**: `predicate_pushdown_test.go`

**测试内容**:
- **Integer Predicate Pushdown**: 整数类型谓词下推（=, >, <, >=, <=）
- **String Predicate Pushdown**: 字符串类型谓词下推
- **Float Predicate Pushdown**: 浮点类型谓词下推
- **Complex Predicates**: 复杂谓词（AND, OR）下推
- **Data Skipping**: 基于统计信息的数据跳过
- **Range-Based Skipping**: 范围基础的文件跳过
- **Performance Benchmarks**: 性能基准测试

**关键验证点**:
- 各种数据类型的谓词正确下推
- 范围查询能够跳过不相关的数据文件
- 谓词下推提供显著的性能提升
- 高选择性查询比全表扫描快
- Null 值正确处理

### 4. Parquet 统计信息测试

**文件**: `parquet_statistics_test.go`

**测试内容**:
- **Int64 Statistics**: INT64 类型的 min/max/null count
- **Int32/Int16/Int8 Statistics**: 其他整数类型统计
- **Float64/Float32 Statistics**: 浮点类型统计
- **String Statistics**: 字符串类型统计
- **Boolean Statistics**: 布尔类型统计
- **Mixed Type Statistics**: 混合类型表统计
- **Statistics Roundtrip**: 统计信息写入和读取

**关键验证点**:
- 所有 Arrow 数据类型都收集统计信息
- Min/Max 值正确计算
- Null count 正确统计
- 统计信息持久化到 Parquet 文件
- 多种数据类型同时存在时统计正确

**支持的数据类型**:
- INT8, INT16, INT32, INT64
- UINT8, UINT16, UINT32, UINT64
- FLOAT32, FLOAT64
- BOOLEAN
- STRING, BINARY
- DATE32, DATE64
- TIMESTAMP (s/ms/us/ns)

### 5. Arrow IPC 序列化测试

**文件**: `arrow_ipc_test.go`

**测试内容**:
- **Basic Schema Roundtrip**: 基础 schema 序列化和反序列化
- **Complex Schema**: 复杂 schema（所有类型）序列化
- **Schema with Field Metadata**: 字段级元数据序列化
- **Schema with Table Metadata**: 表级元数据序列化
- **Timestamp Types**: 时间戳类型序列化
- **Date Types**: 日期类型序列化
- **Multiple Schema Versions**: 多版本 schema 演化
- **Empty Schema**: 空 schema 边界情况
- **Performance Tests**: 大规模 schema 性能测试

**关键验证点**:
- Schema 完整往返（序列化 -> 反序列化）
- 所有字段类型正确保留
- Nullable 属性正确保留
- 字段和表级元数据正确保留
- 大规模 schema（100+ 列）性能良好
- 重复序列化性能稳定

## 测试数据目录

所有测试数据存储在 `test_data/` 目录下：

```
test_data/
├── delta_acid_test/         # ACID 测试数据
├── time_travel_test/        # 时间旅行测试数据
├── predicate_pushdown_test/ # 谓词下推测试数据
├── data_skipping_test/      # 数据跳过测试数据
└── *.parquet               # 各种统计信息测试文件
```

## 测试覆盖率

| 特性类别 | 测试数量 | 覆盖范围 |
|---------|---------|---------|
| ACID 事务 | 6 | 完整 (Atomicity, Consistency, Isolation, Durability) |
| 时间旅行 | 5 | 完整 (版本查询, 时间戳查询, 快照隔离) |
| 谓词下推 | 7 | 完整 (所有比较运算符, 所有数据类型) |
| Parquet 统计 | 8 | 完整 (所有 Arrow 类型, Roundtrip) |
| Arrow IPC | 9 | 完整 (所有类型, 元数据, 性能) |
| **总计** | **35** | **Lakehouse 核心特性全覆盖** |

## 性能基准

测试脚本包含性能基准测试，用于衡量：

1. **谓词下推加速比**: 通常为 2-10x，取决于选择性
2. **数据跳过效率**: 能够跳过 50-90% 的无关文件
3. **统计信息开销**: 写入时 < 5% 性能影响
4. **IPC 序列化性能**: 大 schema (100 列) < 10ms
5. **快照检索性能**: 版本查询 < 1ms

## CI/CD 集成

将测试脚本集成到 CI/CD 流程：

```yaml
# .github/workflows/lakehouse-tests.yml 示例
name: Lakehouse Tests

on: [push, pull_request]

jobs:
  lakehouse-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: '1.21'
      - name: Run Lakehouse Tests
        run: |
          cd test
          ./run_lakehouse_tests.sh
```

## 故障排查

### 测试失败常见原因

1. **内存不足**: Lakehouse 测试需要足够内存处理 Arrow 数据
   - 解决: 增加系统内存或减少测试数据量

2. **磁盘空间不足**: Parquet 文件需要磁盘空间
   - 解决: 清理 `test_data/` 目录

3. **并发冲突**: 多个测试同时运行可能产生冲突
   - 解决: 使用 `-p 1` 参数串行运行测试

4. **超时**: 某些性能测试可能超时
   - 解决: 增加 `-timeout` 参数值

### 调试建议

1. 运行单个测试以隔离问题：
   ```bash
   go test -v ./test/delta_acid_test.go -run TestDeltaLakeACID/Atomicity
   ```

2. 启用详细日志：
   ```bash
   ENVIRONMENT=development go test -v ./test/...
   ```

3. 检查测试数据：
   ```bash
   ls -lh test_data/*/
   ```

## 贡献指南

添加新的 Lakehouse 测试时：

1. 在相应的测试文件中添加测试函数
2. 使用描述性的测试名称（驼峰命名）
3. 包含充分的注释说明测试目的
4. 更新 `run_lakehouse_tests.sh` 脚本
5. 更新本文档的测试覆盖表

## 参考资料

- [Delta Lake Protocol](https://github.com/delta-io/delta/blob/master/PROTOCOL.md)
- [Apache Arrow Format](https://arrow.apache.org/docs/format/Columnar.html)
- [Parquet Format](https://parquet.apache.org/docs/)
- [ACID Properties](https://en.wikipedia.org/wiki/ACID)
