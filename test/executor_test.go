package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yyun543/minidb/internal/catalog"
	"github.com/yyun543/minidb/internal/executor"
	"github.com/yyun543/minidb/internal/optimizer"
	"github.com/yyun543/minidb/internal/parser"
)

func TestExecutor(t *testing.T) {
	// 创建测试环境
	cat := catalog.NewCatalog()
	exec := executor.NewExecutor(cat)

	// 创建测试数据库和表
	setupTestDB(t, cat)

	// 测试SELECT查询
	t.Run("SelectQuery", func(t *testing.T) {
		// 构建查询计划
		plan := &optimizer.LogicalPlan{
			Type: optimizer.SelectPlan,
			Properties: &optimizer.SelectProperties{
				Columns: []*parser.ColumnItem{
					{Column: "id", Alias: ""},
					{Column: "name", Alias: ""},
				},
			},
			Children: []*optimizer.LogicalPlan{
				{
					Type: optimizer.TableScanPlan,
					Properties: &optimizer.TableScanProperties{
						Table: "users",
					},
				},
			},
		}

		// 执行查询
		result, err := exec.Execute(plan)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	// 测试INSERT语句
	t.Run("InsertQuery", func(t *testing.T) {
		plan := &optimizer.LogicalPlan{
			Type: optimizer.InsertPlan,
			Properties: &optimizer.InsertProperties{
				Table:   "users",
				Columns: []string{"id", "name"},
				Values: []parser.Node{
					&parser.IntLiteral{Value: 1},
					&parser.StringLiteral{Value: "test"},
				},
			},
		}

		result, err := exec.Execute(plan)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	// 测试UPDATE语句
	t.Run("UpdateQuery", func(t *testing.T) {
		plan := &optimizer.LogicalPlan{
			Type: optimizer.UpdatePlan,
			Properties: &optimizer.UpdateProperties{
				Table: "users",
				Assignments: []*parser.UpdateAssignment{
					{
						Column: "name",
						Value:  &parser.StringLiteral{Value: "updated"},
					},
				},
				Where: &parser.WhereClause{
					Condition: &parser.BinaryExpr{
						Left:     &parser.Identifier{Value: "id"},
						Operator: "=",
						Right:    &parser.IntLiteral{Value: 1},
					},
				},
			},
		}

		result, err := exec.Execute(plan)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	// 测试DELETE语句
	t.Run("DeleteQuery", func(t *testing.T) {
		plan := &optimizer.LogicalPlan{
			Type: optimizer.DeletePlan,
			Properties: &optimizer.DeleteProperties{
				Table: "users",
				Where: &parser.WhereClause{
					Condition: &parser.BinaryExpr{
						Left:     &parser.Identifier{Value: "id"},
						Operator: "=",
						Right:    &parser.IntLiteral{Value: 1},
					},
				},
			},
		}

		result, err := exec.Execute(plan)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})
}

// 创建测试数据库和表
func setupTestDB(t *testing.T, cat *catalog.Catalog) {
	err := cat.CreateDatabase("testdb")
	assert.NoError(t, err)

	err = cat.UseDatabase("testdb")
	assert.NoError(t, err)

	table := &catalog.TableMeta{
		Name: "users",
		Columns: []catalog.ColumnMeta{
			{Name: "id", Type: "INT64", NotNull: true},
			{Name: "name", Type: "STRING", NotNull: true},
		},
	}
	err = cat.CreateTable(table)
	assert.NoError(t, err)
}
