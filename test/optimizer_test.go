package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yyun543/minidb/internal/optimizer"
	"github.com/yyun543/minidb/internal/parser"
)

func TestOptimizer(t *testing.T) {
	// 测试优化器创建
	t.Run("TestNewOptimizer", func(t *testing.T) {
		opt := optimizer.NewOptimizer()
		assert.NotNil(t, opt)
	})

	// 测试SELECT语句优化
	t.Run("TestOptimizeSelect", func(t *testing.T) {
		// 基本SELECT
		t.Run("BasicSelect", func(t *testing.T) {
			sql := "SELECT id, name FROM users"
			stmt, err := parser.Parse(sql)
			assert.NoError(t, err)

			opt := optimizer.NewOptimizer()
			plan, err := opt.Optimize(stmt)
			assert.NoError(t, err)
			assert.NotNil(t, plan)

			// 验证计划类型
			assert.Equal(t, optimizer.SelectPlan, plan.Type)

			// 验证子计划
			children := plan.Children
			assert.GreaterOrEqual(t, len(children), 1)

			// 验证表扫描计划
			var scanPlan *optimizer.Plan
			for _, child := range children {
				if child.Type == optimizer.TableScanPlan {
					scanPlan = child
					break
				}
			}
			assert.NotNil(t, scanPlan)
			scanProps, ok := scanPlan.Properties.(*optimizer.TableScanProperties)
			assert.True(t, ok)
			assert.Equal(t, "users", scanProps.Table)
		})

		// 带WHERE子句的SELECT
		t.Run("SelectWithWhere", func(t *testing.T) {
			sql := "SELECT id, name FROM users WHERE age > 18"
			stmt, err := parser.Parse(sql)
			assert.NoError(t, err)

			opt := optimizer.NewOptimizer()
			plan, err := opt.Optimize(stmt)
			assert.NoError(t, err)
			assert.NotNil(t, plan)

			// 验证过滤计划
			var filterPlan *optimizer.Plan
			for _, child := range plan.Children {
				if child.Type == optimizer.FilterPlan {
					filterPlan = child
					break
				}
			}
			assert.NotNil(t, filterPlan)
			filterProps, ok := filterPlan.Properties.(*optimizer.FilterProperties)
			assert.True(t, ok)
			assert.NotNil(t, filterProps.Condition)
		})

		// 带JOIN的SELECT
		t.Run("SelectWithJoin", func(t *testing.T) {
			sql := "SELECT u.id, o.order_id FROM users u JOIN orders o ON u.id = o.user_id"
			stmt, err := parser.Parse(sql)
			assert.NoError(t, err)

			opt := optimizer.NewOptimizer()
			plan, err := opt.Optimize(stmt)
			assert.NoError(t, err)
			assert.NotNil(t, plan)

			// 验证JOIN计划
			var joinPlan *optimizer.Plan
			for _, child := range plan.Children {
				if child.Type == optimizer.JoinPlan {
					joinPlan = child
					break
				}
			}
			assert.NotNil(t, joinPlan)
			joinProps, ok := joinPlan.Properties.(*optimizer.JoinProperties)
			assert.True(t, ok)
			assert.Equal(t, "INNER", joinProps.JoinType)
			assert.Equal(t, "users", joinProps.Left)
			assert.Equal(t, "orders", joinProps.Right)
		})

		// 带ORDER BY的SELECT
		t.Run("SelectWithOrderBy", func(t *testing.T) {
			sql := "SELECT id, name FROM users ORDER BY name DESC"
			stmt, err := parser.Parse(sql)
			assert.NoError(t, err)

			opt := optimizer.NewOptimizer()
			plan, err := opt.Optimize(stmt)
			assert.NoError(t, err)
			assert.NotNil(t, plan)

			// 验证ORDER BY计划
			var orderPlan *optimizer.Plan
			for _, child := range plan.Children {
				if child.Type == optimizer.OrderPlan {
					orderPlan = child
					break
				}
			}
			assert.NotNil(t, orderPlan)
			orderProps, ok := orderPlan.Properties.(*optimizer.OrderByProperties)
			assert.True(t, ok)
			assert.Len(t, orderProps.OrderKeys, 1)
			assert.Equal(t, "name", orderProps.OrderKeys[0].Column)
			assert.Equal(t, "DESC", orderProps.OrderKeys[0].Direction)
		})

		// 带GROUP BY的SELECT
		t.Run("SelectWithGroupBy", func(t *testing.T) {
			sql := "SELECT department, COUNT(*) FROM employees GROUP BY department"
			stmt, err := parser.Parse(sql)
			assert.NoError(t, err)

			opt := optimizer.NewOptimizer()
			plan, err := opt.Optimize(stmt)
			assert.NoError(t, err)
			assert.NotNil(t, plan)

			// 验证GROUP BY计划
			var groupPlan *optimizer.Plan
			for _, child := range plan.Children {
				if child.Type == optimizer.GroupPlan {
					groupPlan = child
					break
				}
			}
			assert.NotNil(t, groupPlan)
			groupProps, ok := groupPlan.Properties.(*optimizer.GroupByProperties)
			assert.True(t, ok)
			assert.Contains(t, groupProps.GroupKeys, "department")
		})
	})

	// 测试INSERT语句优化
	t.Run("TestOptimizeInsert", func(t *testing.T) {
		sql := "INSERT INTO users (id, name) VALUES (1, 'test')"
		stmt, err := parser.Parse(sql)
		assert.NoError(t, err)

		opt := optimizer.NewOptimizer()
		plan, err := opt.Optimize(stmt)
		assert.NoError(t, err)
		assert.NotNil(t, plan)

		assert.Equal(t, optimizer.InsertPlan, plan.Type)
		props, ok := plan.Properties.(*optimizer.InsertProperties)
		assert.True(t, ok)
		assert.Equal(t, "users", props.Table)
		assert.Equal(t, []string{"id", "name"}, props.Columns)
	})

	// 测试UPDATE语句优化
	t.Run("TestOptimizeUpdate", func(t *testing.T) {
		sql := "UPDATE users SET name = 'updated' WHERE id = 1"
		stmt, err := parser.Parse(sql)
		assert.NoError(t, err)

		opt := optimizer.NewOptimizer()
		plan, err := opt.Optimize(stmt)
		assert.NoError(t, err)
		assert.NotNil(t, plan)

		assert.Equal(t, optimizer.UpdatePlan, plan.Type)
		props, ok := plan.Properties.(*optimizer.UpdateProperties)
		assert.True(t, ok)
		assert.Equal(t, "users", props.Table)
		assert.NotNil(t, props.Where)
	})

	// 测试DELETE语句优化
	t.Run("TestOptimizeDelete", func(t *testing.T) {
		sql := "DELETE FROM users WHERE id = 1"
		stmt, err := parser.Parse(sql)
		assert.NoError(t, err)

		opt := optimizer.NewOptimizer()
		plan, err := opt.Optimize(stmt)
		assert.NoError(t, err)
		assert.NotNil(t, plan)

		assert.Equal(t, optimizer.DeletePlan, plan.Type)
		props, ok := plan.Properties.(*optimizer.DeleteProperties)
		assert.True(t, ok)
		assert.Equal(t, "users", props.Table)
		assert.NotNil(t, props.Where)
	})
}
