package storage

import (
	"fmt"
	"sync"

	"github.com/yyun543/minidb/internal/parser"
)

// ColumnStore 实现基于列的存储引擎
type ColumnStore struct {
	store  *FileStore
	tables map[string]*ColumnTable
	mu     sync.RWMutex
}

// ColumnTable 列式存储的表结构
type ColumnTable struct {
	Name    string
	Schema  Schema
	Columns map[string]*Column
}

// Column 列数据
type Column struct {
	Name     string
	Type     string
	Values   []interface{}
	Nullable bool
}

// NewColumnStore 创建新的列存储引擎
func NewColumnStore(path string) (*ColumnStore, error) {
	store, err := NewFileStore(path)
	if err != nil {
		return nil, err
	}

	return &ColumnStore{
		store:  store,
		tables: make(map[string]*ColumnTable),
	}, nil
}

// Insert 插入列数据
func (cs *ColumnStore) Insert(tableName string, values map[string]interface{}) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	table, exists := cs.tables[tableName]
	if !exists {
		return fmt.Errorf("table %s does not exist", tableName)
	}

	// 验证数据
	if err := validateValues(table.Schema, values); err != nil {
		return err
	}

	// 将数据添加到对应的列
	for colName, value := range values {
		col := table.Columns[colName]
		col.Values = append(col.Values, value)
	}

	return nil
}

// Select 查询列数据
func (cs *ColumnStore) Select(tableName string, columns []string, where string) ([]Row, error) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	table, exists := cs.tables[tableName]
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

	// 获取满足条件的行索引
	rowIndices, err := cs.evaluateWhere(table, where)
	if err != nil {
		return nil, err
	}

	// 构建结果集
	var result []Row
	for _, rowIdx := range rowIndices {
		row := Row{
			ID:   rowIdx,
			Data: make(map[string]interface{}),
		}
		for _, colName := range columns {
			if col, ok := table.Columns[colName]; ok {
				row.Data[colName] = col.Values[rowIdx]
			}
		}
		result = append(result, row)
	}

	return result, nil
}

// 其他方法实现...

// evaluateWhere 的替代实现
func (cs *ColumnStore) evaluateWhere(table *ColumnTable, where string) ([]int, error) {
	if where == "" {
		// 处理空WHERE条件的情况
		return cs.getAllRowIndices(table), nil
	}

	// 直接解析WHERE表达式
	p := parser.NewParser(where)
	expr, err := p.ParseWhereExpression()
	if err != nil {
		return nil, fmt.Errorf("failed to parse WHERE condition: %v", err)
	}

	return cs.evaluateExpression(table, expr)
}

// getAllRowIndices 返回所有行的索引
func (cs *ColumnStore) getAllRowIndices(table *ColumnTable) []int {
	var rowCount int
	// 获取第一个列的长度作为行数
	for _, col := range table.Columns {
		rowCount = len(col.Values)
		break
	}
	return makeSequence(rowCount)
}

// 辅助函数：生成序列
func makeSequence(n int) []int {
	result := make([]int, n)
	for i := 0; i < n; i++ {
		result[i] = i
	}
	return result
}

func (cs *ColumnStore) evaluateExpression(table *ColumnTable, expr parser.Expression) ([]int, error) {
	switch e := expr.(type) {
	case *parser.ComparisonExpr:
		return cs.evaluateComparison(table, e)
	case *parser.BinaryExpr:
		return cs.evaluateBinary(table, e)
	default:
		return nil, fmt.Errorf("unsupported expression type: %T", expr)
	}
}

func (cs *ColumnStore) evaluateComparison(table *ColumnTable, expr *parser.ComparisonExpr) ([]int, error) {
	// TODO 实现比较逻辑
	return nil, nil
}

func (cs *ColumnStore) evaluateBinary(table *ColumnTable, expr *parser.BinaryExpr) ([]int, error) {
	// TODO 实现二元运算逻辑
	return nil, nil
}
