package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yyun543/minidb/internal/optimizer"
	"github.com/yyun543/minidb/internal/parser"
)

// Optimizer 查询优化器单元测试

func TestOptimizer(t *testing.T) {
	// 测试基本的SELECT查询优化
	t.Run("OptimizeSimpleSelect", func(t *testing.T) {
		sql := "SELECT id, name FROM users WHERE age > 18;"
		stmt, err := parser.Parse(sql)
		assert.NoError(t, err)

		opt := optimizer.NewOptimizer()
		plan, err := opt.Optimize(stmt)

		assert.NoError(t, err)
		assert.NotNil(t, plan)
		assert.Equal(t, optimizer.SelectPlan, plan.Type)

		// 验证SELECT属性
		selectProps, ok := plan.Properties.(*optimizer.SelectProperties)
		assert.True(t, ok)
		assert.Equal(t, []string{"id", "name"}, selectProps.Columns)

		// 验证子节点
		assert.Len(t, plan.Children, 2) // 应该有TableScan和Filter两个子节点

		// 验证TableScan节点
		assert.Equal(t, optimizer.TableScanPlan, plan.Children[0].Type)
		scanProps, ok := plan.Children[0].Properties.(*optimizer.TableScanProperties)
		assert.True(t, ok)
		assert.Equal(t, "users", scanProps.Table)

		// 验证Filter节点
		assert.Equal(t, optimizer.FilterPlan, plan.Children[1].Type)
		filterProps, ok := plan.Children[1].Properties.(*optimizer.FilterProperties)
		assert.True(t, ok)
		assert.NotNil(t, filterProps.Condition)
	})

	// 测试带JOIN的查询优化
	t.Run("OptimizeJoinQuery", func(t *testing.T) {
		sql := `
			SELECT u.id, u.name, o.order_id 
			FROM users u 
			JOIN orders o ON u.id = o.user_id 
			WHERE u.age > 18;
		`
		stmt, err := parser.Parse(sql)
		assert.NoError(t, err)

		opt := optimizer.NewOptimizer()
		plan, err := opt.Optimize(stmt)

		assert.NoError(t, err)
		assert.NotNil(t, plan)
		assert.Equal(t, optimizer.SelectPlan, plan.Type)

		// 验证SELECT属性
		selectProps, ok := plan.Properties.(*optimizer.SelectProperties)
		assert.True(t, ok)
		assert.Equal(t, []string{"u.id", "u.name", "o.order_id"}, selectProps.Columns)

		// 验证子节点结构
		assert.Len(t, plan.Children, 3) // TableScan, Join, 和 Filter

		// 验证Join节点
		assert.Equal(t, optimizer.JoinPlan, plan.Children[1].Type)
		joinProps, ok := plan.Children[1].Properties.(*optimizer.JoinProperties)
		assert.True(t, ok)
		assert.Equal(t, "orders", joinProps.Table)
		assert.NotNil(t, joinProps.Condition)
	})

	// 测试INSERT语句优化
	t.Run("OptimizeInsert", func(t *testing.T) {
		sql := "INSERT INTO users (id, name, age) VALUES (1, 'test', 20);"
		stmt, err := parser.Parse(sql)
		assert.NoError(t, err)

		opt := optimizer.NewOptimizer()
		plan, err := opt.Optimize(stmt)

		assert.NoError(t, err)
		assert.NotNil(t, plan)
		assert.Equal(t, optimizer.InsertPlan, plan.Type)

		insertProps, ok := plan.Properties.(*optimizer.InsertProperties)
		assert.True(t, ok)
		assert.Equal(t, "users", insertProps.Table)
		assert.Equal(t, []string{"id", "name", "age"}, insertProps.Columns)
		assert.Len(t, insertProps.Values, 3)
	})

	// 测试UPDATE语句优化
	t.Run("OptimizeUpdate", func(t *testing.T) {
		sql := "UPDATE users SET name = 'updated' WHERE id = 1;"
		stmt, err := parser.Parse(sql)
		assert.NoError(t, err)

		opt := optimizer.NewOptimizer()
		plan, err := opt.Optimize(stmt)

		assert.NoError(t, err)
		assert.NotNil(t, plan)
		assert.Equal(t, optimizer.UpdatePlan, plan.Type)

		updateProps, ok := plan.Properties.(*optimizer.UpdateProperties)
		assert.True(t, ok)
		assert.Equal(t, "users", updateProps.Table)
		assert.Len(t, updateProps.Assignments, 1)
		assert.NotNil(t, updateProps.Where)
	})

	// 测试DELETE语句优化
	t.Run("OptimizeDelete", func(t *testing.T) {
		sql := "DELETE FROM users WHERE id = 1;"
		stmt, err := parser.Parse(sql)
		assert.NoError(t, err)

		opt := optimizer.NewOptimizer()
		plan, err := opt.Optimize(stmt)

		assert.NoError(t, err)
		assert.NotNil(t, plan)
		assert.Equal(t, optimizer.DeletePlan, plan.Type)

		deleteProps, ok := plan.Properties.(*optimizer.DeleteProperties)
		assert.True(t, ok)
		assert.Equal(t, "users", deleteProps.Table)
		assert.NotNil(t, deleteProps.Where)
	})

	// 测试JOIN重排序优化规则
	t.Run("TestJoinReorderRule", func(t *testing.T) {
		sql := `
			SELECT u.id, u.name, o.order_id, p.product_name
			FROM users u 
			JOIN orders o ON u.id = o.user_id
			JOIN products p ON o.product_id = p.id
			WHERE u.age > 18;
		`
		stmt, err := parser.Parse(sql)
		assert.NoError(t, err)

		opt := optimizer.NewOptimizer()
		plan, err := opt.Optimize(stmt)

		assert.NoError(t, err)
		assert.NotNil(t, plan)

		// 验证JOIN节点的顺序
		joinCount := 0
		for _, child := range plan.Children {
			if child.Type == optimizer.JoinPlan {
				joinCount++
				// 这里可以添加更多的验证逻辑，比如检查JOIN的顺序是否符合优化预期
			}
		}
		assert.Equal(t, 2, joinCount) // 应该有两个JOIN节点
	})
}
