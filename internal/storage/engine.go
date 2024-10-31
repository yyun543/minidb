package storage

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/yyun543/minidb/internal/parser"
)

// Row 表示表中的一行数据
type Row struct {
	ID      int                    // 行ID
	Data    map[string]interface{} // 列数据
	Created time.Time              // 创建时间
	Updated time.Time              // 更新时间
	Values  map[string]interface{}
}

// Schema 表示表结构
type Schema map[string]Column

// Table 表示数据表
type Table struct {
	Name    string
	Schema  Schema
	Rows    map[int]Row
	LastID  int
	mu      sync.RWMutex
	rowLock sync.Map
}

// 实现行级锁
func (t *Table) lockRow(id int) func() {
	value, _ := t.rowLock.LoadOrStore(id, &sync.Mutex{})
	mu := value.(*sync.Mutex)
	mu.Lock()
	return mu.Unlock
}

// 使用行级锁的更新操作
func (t *Table) UpdateRow(id int, values map[string]interface{}) error {
	unlock := t.lockRow(id)
	defer unlock()

	row, exists := t.Rows[id]
	if !exists {
		return fmt.Errorf("row %d does not exist", id)
	}

	// 更新数据
	for k, v := range values {
		row.Data[k] = v
	}
	row.Updated = time.Now()
	t.Rows[id] = row

	return nil
}

// 添加死锁检测
type DeadlockDetector struct {
	waitFor sync.Map // 记录等待关系
	mu      sync.Mutex
}

func (d *DeadlockDetector) CheckDeadlock(from, to int) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	visited := make(map[int]bool)
	return d.dfs(to, from, visited)
}

func (d *DeadlockDetector) dfs(current, target int, visited map[int]bool) bool {
	if current == target {
		return true
	}

	visited[current] = true
	if waitList, ok := d.waitFor.Load(current); ok {
		for _, next := range waitList.([]int) {
			if !visited[next] && d.dfs(next, target, visited) {
				return true
			}
		}
	}
	return false
}

// StorageEngine 定义存储引擎接口
type StorageEngine interface {
	CreateTable(name string, schema Schema) error
	DropTable(name string) error
	Insert(tableName string, values map[string]interface{}) error
	Select(tableName string, columns []string, where string) ([]Row, error)
	Update(tableName string, values map[string]interface{}, where string) (int, error)
	Delete(tableName string, where string) (int, error)
	Backup() ([]byte, error)
	Restore(data []byte) error
}

// Engine 混合存储引擎，同时支持行存储和列存储
type Engine struct {
	rowStore    *RowStore    // OLTP行存储
	columnStore *ColumnStore // OLAP列存储
	tables      map[string]*Table
	mu          sync.RWMutex
	metrics     Metrics
}

// NewEngine 创建新的存储引擎
func NewEngine() *Engine {
	return &Engine{
		tables:  make(map[string]*Table),
		metrics: Metrics{},
	}
}

// CreateTable 创建新表
func (e *Engine) CreateTable(name string, schema Schema) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.tables[name]; exists {
		return fmt.Errorf("table %s already exists", name)
	}

	e.tables[name] = &Table{
		Name:   name,
		Schema: schema,
		Rows:   make(map[int]Row),
	}
	return nil
}

// DropTable 删除表
func (e *Engine) DropTable(name string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.tables[name]; !exists {
		return fmt.Errorf("table %s does not exist", name)
	}

	delete(e.tables, name)
	return nil
}

// Insert 插入数据
func (e *Engine) Insert(tableName string, values map[string]interface{}) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	table, exists := e.tables[tableName]
	if !exists {
		return fmt.Errorf("table %s does not exist", tableName)
	}

	// 验证数据
	if err := validateValues(table.Schema, values); err != nil {
		return err
	}

	// 生成新的行ID
	table.LastID++
	rowID := table.LastID

	// 创建新行
	now := time.Now()
	row := Row{
		ID:      rowID,
		Data:    values,
		Created: now,
		Updated: now,
	}

	table.Rows[rowID] = row
	return nil
}

// Select 实现查询操作
func (e *Engine) Select(tableName string, columns []string, where string) ([]Row, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	table, exists := e.tables[tableName]
	if !exists {
		return nil, fmt.Errorf("table %s does not exist", tableName)
	}

	// 如果没有指定列，返回所有列
	if len(columns) == 0 || (len(columns) == 1 && columns[0] == "*") {
		columns = make([]string, 0, len(table.Schema))
		for col := range table.Schema {
			columns = append(columns, col)
		}
	}

	// 验证列是否存在
	for _, col := range columns {
		if _, exists := table.Schema[col]; !exists && col != "*" {
			return nil, fmt.Errorf("column %s does not exist", col)
		}
	}

	var result []Row
	// 简单的WHERE条件处理
	for _, row := range table.Rows {
		if evaluateWhereCondition(row, where) {
			// 只选择指定的列
			selectedRow := Row{
				ID:      row.ID,
				Data:    make(map[string]interface{}),
				Created: row.Created,
				Updated: row.Updated,
			}
			for _, col := range columns {
				if col == "*" {
					selectedRow.Data = row.Data
					break
				}
				selectedRow.Data[col] = row.Data[col]
			}
			result = append(result, selectedRow)
		}
	}

	return result, nil
}

// evaluateWhereCondition 评估WHERE条件
func evaluateWhereCondition(row Row, where string) bool {
	if where == "" {
		return true
	}
	// TODO: 实现WHERE条件解析和评估
	// 这里需要实现一个简单的条件解析器
	return true
}

// validateSchema 验证表结构
func validateSchema(schema Schema) error {
	if len(schema) == 0 {
		return fmt.Errorf("empty schema")
	}

	for colName, col := range schema {
		if colName == "" {
			return fmt.Errorf("empty column name")
		}
		if err := validateColumnType(col.Type); err != nil {
			return fmt.Errorf("invalid column %s: %v", colName, err)
		}
	}
	return nil
}

// validateColumnType 验证列类型
func validateColumnType(colType string) error {
	validTypes := map[string]bool{
		"string": true,
		"int":    true,
		"float":  true,
		"bool":   true,
		"date":   true,
	}

	if !validTypes[colType] {
		return fmt.Errorf("unsupported type: %s", colType)
	}
	return nil
}

// Update 更新数据
func (e *Engine) Update(tableName string, values map[string]interface{}, where string) (int, error) {
	e.mu.RLock()
	table, exists := e.tables[tableName]
	e.mu.RUnlock()

	if !exists {
		return 0, fmt.Errorf("table %s does not exist", tableName)
	}

	// 验证数据类型
	if err := validateValues(table.Schema, values); err != nil {
		return 0, err
	}

	table.mu.Lock()
	defer table.mu.Unlock()

	updateCount := 0
	now := time.Now()

	// 遍历所有行
	for id, row := range table.Rows {
		// 如果有WHERE条件，进行过滤
		if where != "" && !e.evaluateWhere(row, where) {
			continue
		}

		// 更新数据
		for key, value := range values {
			row.Data[key] = value
		}
		row.Updated = now
		table.Rows[id] = row
		updateCount++
	}

	return updateCount, nil
}

// evaluateWhere 评估WHERE条件
func (e *Engine) evaluateWhere(row Row, where string) bool {
	if where == "" {
		return true
	}

	// 解析WHERE条件
	p := parser.NewParser(where)
	expr, err := p.ParseWhereExpression()
	if err != nil {
		return false
	}

	return e.evaluateExpression(row, expr)
}

// evaluateExpression 评估表达式
func (e *Engine) evaluateExpression(row Row, expr parser.Expression) bool {
	switch e := expr.(type) {
	case *parser.ComparisonExpr:
		return e.evaluateComparison(row)
	case *parser.BinaryExpr:
		return e.evaluateBinary(row)
	default:
		return false
	}
}

// Delete 删除数据
func (e *Engine) Delete(tableName string, where string) (int, error) {
	e.mu.RLock()
	table, exists := e.tables[tableName]
	e.mu.RUnlock()

	if !exists {
		return 0, fmt.Errorf("table %s does not exist", tableName)
	}

	table.mu.Lock()
	defer table.mu.Unlock()

	deleteCount := 0

	// 遍历所有行
	for id, row := range table.Rows {
		// 如果有WHERE条件，进行过滤
		if where != "" && !e.evaluateWhere(row, where) {
			continue
		}

		delete(table.Rows, id)
		deleteCount++
	}

	return deleteCount, nil
}

// 辅助方法

// validateValues 验证数据值是否符合schema定义
func validateValues(schema Schema, values map[string]interface{}) error {
	// 1. 检查必需字段
	for colName, colDef := range schema {
		if !colDef.Nullable {
			if value, exists := values[colName]; !exists || value == nil {
				return fmt.Errorf("column '%s' cannot be null", colName)
			}
		}
	}

	// 2. 检查提供的值是否都在schema中定义
	for colName := range values {
		if _, exists := schema[colName]; !exists {
			return fmt.Errorf("column '%s' is not defined in schema", colName)
		}
	}

	// 3. 验证每个值的类型
	for colName, value := range values {
		colDef := schema[colName]

		// 如果值为nil且允许为空，则跳过类型检查
		if value == nil {
			if colDef.Nullable {
				continue
			}
			return fmt.Errorf("column '%s' cannot be null", colName)
		}

		if err := validateType(colDef.Type, value); err != nil {
			return fmt.Errorf("invalid value for column '%s': %v", colName, err)
		}
	}

	return nil
}

// validateType 验证单个值的类型
func validateType(expectedType string, value interface{}) error {
	switch DataType(expectedType) {
	case TypeString:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}

	case TypeInt:
		switch v := value.(type) {
		case int, int32, int64:
			// 这些类型都是可接受的
		case float64:
			// 检查是否为整数
			if v != float64(int64(v)) {
				return fmt.Errorf("expected integer, got float")
			}
		case string:
			// 尝试转换字符串为整数
			if _, err := strconv.ParseInt(v, 10, 64); err != nil {
				return fmt.Errorf("cannot convert string '%s' to integer", v)
			}
		default:
			return fmt.Errorf("expected integer, got %T", value)
		}

	case TypeFloat:
		switch v := value.(type) {
		case float32, float64:
			// 这些类型都是可接受的
		case int, int32, int64:
			// 整数可以自动转换为浮点数
		case string:
			// 尝试转换字符串为浮点数
			if _, err := strconv.ParseFloat(v, 64); err != nil {
				return fmt.Errorf("cannot convert string '%s' to float", v)
			}
		default:
			return fmt.Errorf("expected float, got %T", value)
		}

	case TypeBool:
		switch v := value.(type) {
		case bool:
			// 直接是布尔类型
		case string:
			// 尝试转换字符串为布尔值
			if _, err := strconv.ParseBool(v); err != nil {
				return fmt.Errorf("cannot convert string '%s' to boolean", v)
			}
		case int:
			// 只允许0和1
			if v != 0 && v != 1 {
				return fmt.Errorf("integer value for boolean must be 0 or 1")
			}
		default:
			return fmt.Errorf("expected boolean, got %T", value)
		}

	case TypeDateTime:
		switch v := value.(type) {
		case time.Time:
			// 直接是时间类型
		case string:
			// 尝试解析多种常见的时间格式
			layouts := []string{
				time.RFC3339,
				"2006-01-02 15:04:05",
				"2006-01-02",
			}
			valid := false
			for _, layout := range layouts {
				if _, err := time.Parse(layout, v); err == nil {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("cannot parse '%s' as datetime", v)
			}
		default:
			return fmt.Errorf("expected datetime, got %T", value)
		}

	default:
		return fmt.Errorf("unsupported data type: %s", expectedType)
	}

	return nil
}

// Backup 创建数据库备份
func (e *Engine) Backup() ([]byte, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return json.Marshal(e.tables)
}

// Restore 从备份恢复数据库
func (e *Engine) Restore(data []byte) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	return json.Unmarshal(data, &e.tables)
}

// RowStore 行存储引擎
type RowStore struct {
	path   string
	tables map[string]*Table

	// ... 其他字段
}

// 添加事务支持
type Transaction struct {
	id      int64
	changes map[string][]Row
	engine  *Engine
	mu      sync.RWMutex
}

// 开始事务
func (e *Engine) Begin() (*Transaction, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	txID := time.Now().UnixNano()
	return &Transaction{
		id:      txID,
		changes: make(map[string][]Row),
		engine:  e,
	}, nil
}

// 实现完整的CRUD操作
func (tx *Transaction) Insert(tableName string, values map[string]interface{}) error {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	table, exists := tx.engine.tables[tableName]
	if !exists {
		return fmt.Errorf("table %s does not exist", tableName)
	}

	// 验证数据
	if err := validateValues(table.Schema, values); err != nil {
		return err
	}

	// 生成新行
	row := Row{
		ID:      table.LastID + 1,
		Data:    values,
		Created: time.Now(),
		Updated: time.Now(),
	}

	// 记录变更
	tx.changes[tableName] = append(tx.changes[tableName], row)
	return nil
}

// 添加批量操作支持
type Batch struct {
	operations []operation
	engine     *Engine
}

type operation struct {
	op    string
	table string
	data  interface{}
}

func (e *Engine) NewBatch() *Batch {
	return &Batch{
		engine: e,
	}
}

func (b *Batch) Insert(table string, values map[string]interface{}) {
	b.operations = append(b.operations, operation{
		op:    "insert",
		table: table,
		data:  values,
	})
}

func (b *Batch) Execute() error {
	tx, err := b.engine.Begin()
	if err != nil {
		return err
	}

	for _, op := range b.operations {
		switch op.op {
		case "insert":
			if err := tx.Insert(op.table, op.data.(map[string]interface{})); err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return tx.Commit()
}

// Transaction 实现事务接口
func (tx *Transaction) Commit() error {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	// 将变更应用到存储引擎
	for tableName, rows := range tx.changes {
		table := tx.engine.tables[tableName]
		for _, row := range rows {
			table.Rows[row.ID] = row
		}
	}

	return nil
}

func (tx *Transaction) Rollback() error {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	// 清空未提交的变更
	tx.changes = make(map[string][]Row)
	return nil
}

// GetRow 获取指定行
func (e *Engine) GetRow(tableName string, id int) (Row, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	table, exists := e.tables[tableName]
	if !exists {
		return Row{}, fmt.Errorf("table %s does not exist", tableName)
	}

	row, exists := table.Rows[id]
	if !exists {
		return Row{}, fmt.Errorf("row %d does not exist", id)
	}

	return row, nil
}

// Error types
var (
	ErrTableNotFound    = fmt.Errorf("table not found")
	ErrColumnNotFound   = fmt.Errorf("column not found")
	ErrDuplicateTable   = fmt.Errorf("table already exists")
	ErrInvalidSchema    = fmt.Errorf("invalid schema")
	ErrTransactionError = fmt.Errorf("transaction error")
)

// logError 记录错误信息
func (e *Engine) logError(operation string, err error) {
	log.Printf("Error during %s: %v", operation, err)
}

// wrapError 包装错误信息
func wrapError(err error, msg string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", msg, err)
}

// Metrics 存储引擎指标
type Metrics struct {
	QueryCount       int64
	InsertCount      int64
	UpdateCount      int64
	DeleteCount      int64
	CacheHitCount    int64
	CacheMissCount   int64
	TransactionCount int64
	ErrorCount       int64
	mu               sync.RWMutex
}

// incrementMetric 增加指标计数
func (m *Metrics) incrementMetric(metric *int64) {
	atomic.AddInt64(metric, 1)
}

// GetMetrics 获取当前指标
func (e *Engine) GetMetrics() Metrics {
	e.metrics.mu.RLock()
	defer e.metrics.mu.RUnlock()
	return e.metrics
}
