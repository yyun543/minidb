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

			// 验证计划树结构：
			// Project(id, name)
			//   └─ TableScan(users)
			assert.Equal(t, optimizer.SelectPlan, plan.Type)
			assert.Len(t, plan.Children, 1)

			// 验证投影算子
			projectProps, ok := plan.Properties.(*optimizer.SelectProperties)
			assert.True(t, ok)
			assert.Len(t, projectProps.Columns, 2)
			assert.Equal(t, "id", projectProps.Columns[0].Column)
			assert.Equal(t, "name", projectProps.Columns[1].Column)

			// 验证表扫描算子
			scanPlan := plan.Children[0]
			assert.Equal(t, optimizer.TableScanPlan, scanPlan.Type)
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

			// 验证计划树结构：
			// Project(id, name)
			//   └─ Filter(age > 18)
			//        └─ TableScan(users)
			assert.Equal(t, optimizer.SelectPlan, plan.Type)
			assert.Len(t, plan.Children, 1)

			// 验证投影算子
			projectProps, ok := plan.Properties.(*optimizer.SelectProperties)
			assert.True(t, ok)
			assert.Len(t, projectProps.Columns, 2)

			// 验证过滤算子
			filterPlan := plan.Children[0]
			assert.Equal(t, optimizer.FilterPlan, filterPlan.Type)
			filterProps, ok := filterPlan.Properties.(*optimizer.FilterProperties)
			assert.True(t, ok)
			assert.NotNil(t, filterProps.Condition)

			// 验证表扫描算子
			scanPlan := filterPlan.Children[0]
			assert.Equal(t, optimizer.TableScanPlan, scanPlan.Type)
			scanProps, ok := scanPlan.Properties.(*optimizer.TableScanProperties)
			assert.True(t, ok)
			assert.Equal(t, "users", scanProps.Table)
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

			// 验证计划树结构：
			// Project(u.id, o.order_id)
			//   └─ Join(INNER, u.id = o.user_id)
			//        ├─ TableScan(users AS u)
			//        └─ TableScan(orders AS o)
			assert.Equal(t, optimizer.SelectPlan, plan.Type)
			assert.Len(t, plan.Children, 1)

			// 验证投影算子
			projectProps, ok := plan.Properties.(*optimizer.SelectProperties)
			assert.True(t, ok)
			assert.Len(t, projectProps.Columns, 2)

			// 验证JOIN算子
			joinPlan := plan.Children[0]
			assert.Equal(t, optimizer.JoinPlan, joinPlan.Type)
			joinProps, ok := joinPlan.Properties.(*optimizer.JoinProperties)
			assert.True(t, ok)
			assert.Equal(t, "INNER", joinProps.JoinType)
			assert.Equal(t, "users", joinProps.Left)
			assert.Equal(t, "u", joinProps.LeftAlias)
			assert.Equal(t, "orders", joinProps.Right)
			assert.Equal(t, "o", joinProps.RightAlias)
			assert.NotNil(t, joinProps.Condition)

			// 验证左表扫描
			leftScan := joinPlan.Children[0]
			assert.Equal(t, optimizer.TableScanPlan, leftScan.Type)
			leftProps, ok := leftScan.Properties.(*optimizer.TableScanProperties)
			assert.True(t, ok)
			assert.Equal(t, "users", leftProps.Table)
			assert.Equal(t, "u", leftProps.TableAlias)

			// 验证右表扫描
			rightScan := joinPlan.Children[1]
			assert.Equal(t, optimizer.TableScanPlan, rightScan.Type)
			rightProps, ok := rightScan.Properties.(*optimizer.TableScanProperties)
			assert.True(t, ok)
			assert.Equal(t, "orders", rightProps.Table)
			assert.Equal(t, "o", rightProps.TableAlias)
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

			// 验证计划树结构：
			// Project(id, name)
			//   └─ OrderBy(name DESC)
			//        └─ TableScan(users)
			assert.Equal(t, optimizer.SelectPlan, plan.Type)
			assert.Len(t, plan.Children, 1)

			// 验证投影算子
			projectProps, ok := plan.Properties.(*optimizer.SelectProperties)
			assert.True(t, ok)
			assert.Len(t, projectProps.Columns, 2)

			// 验证ORDER BY算子
			orderByPlan := plan.Children[0]
			assert.Equal(t, optimizer.OrderPlan, orderByPlan.Type)
			orderByProps, ok := orderByPlan.Properties.(*optimizer.OrderByProperties)
			assert.True(t, ok)
			assert.Len(t, orderByProps.OrderKeys, 1)
			assert.Equal(t, "name", orderByProps.OrderKeys[0].Column)
			assert.Equal(t, "DESC", orderByProps.OrderKeys[0].Direction)

			// 验证表扫描算子
			scanPlan := orderByPlan.Children[0]
			assert.Equal(t, optimizer.TableScanPlan, scanPlan.Type)
			scanProps, ok := scanPlan.Properties.(*optimizer.TableScanProperties)
			assert.True(t, ok)
			assert.Equal(t, "users", scanProps.Table)
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

			// 验证计划树结构：
			// Project(department, COUNT(*))
			//   └─ GroupBy(department)
			//        └─ TableScan(employees)
			assert.Equal(t, optimizer.SelectPlan, plan.Type)
			assert.Len(t, plan.Children, 1)

			// 验证投影算子
			projectProps, ok := plan.Properties.(*optimizer.SelectProperties)
			assert.True(t, ok)
			assert.Len(t, projectProps.Columns, 2)

			// 验证GROUP BY算子
			groupPlan := plan.Children[0]
			assert.Equal(t, optimizer.GroupPlan, groupPlan.Type)
			groupProps, ok := groupPlan.Properties.(*optimizer.GroupByProperties)
			assert.True(t, ok)
			assert.Contains(t, groupProps.GroupKeys[0].Column, "department")

			// 验证表扫描算子
			scanPlan := groupPlan.Children[0]
			assert.Equal(t, optimizer.TableScanPlan, scanPlan.Type)
			scanProps, ok := scanPlan.Properties.(*optimizer.TableScanProperties)
			assert.True(t, ok)
			assert.Equal(t, "employees", scanProps.Table)
		})
	})

	// 测试INSERT语句优化
	t.Run("TestOptimizeInsert", func(t *testing.T) {
		// 测试基本INSERT
		t.Run("BasicInsert", func(t *testing.T) {
			sql := "INSERT INTO users (id, name, age) VALUES (1, 'test', 25)"
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
			assert.Equal(t, []string{"id", "name", "age"}, props.Columns)
			assert.Len(t, props.Values, 3)
		})
	})

	// 测试UPDATE语句优化
	t.Run("TestOptimizeUpdate", func(t *testing.T) {
		// 测试带条件的UPDATE
		t.Run("UpdateWithCondition", func(t *testing.T) {
			sql := "UPDATE users SET name = 'updated', age = 30 WHERE id = 1 AND status = 'active'"
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
			assert.Len(t, props.Assignments, 2)
			assert.NotNil(t, props.Where)
		})
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

	// 测试复杂SELECT语句优化
	t.Run("TestOptimizeComplexSelect", func(t *testing.T) {
		// 带GROUP BY, HAVING, ORDER BY和LIMIT的复杂查询
		t.Run("SelectWithAllClauses", func(t *testing.T) {
			sql := `
				SELECT department, COUNT(*) as count 
				FROM employees 
				WHERE salary > 50000 
				GROUP BY department 
				HAVING COUNT(*) > 5 
				ORDER BY count DESC 
				LIMIT 10
			`
			stmt, err := parser.Parse(sql)
			assert.NoError(t, err)

			opt := optimizer.NewOptimizer()
			plan, err := opt.Optimize(stmt)
			assert.NoError(t, err)
			assert.NotNil(t, plan)

			// 验证计划树结构：
			// Project(department, COUNT(*))
			//   └─ Limit(10)
			//        └─ OrderBy(count DESC)
			//             └─ Having(COUNT(*) > 5)
			//                  └─ GroupBy(department)
			//                       └─ Filter(salary > 50000)
			//                            └─ TableScan(employees)
			assert.Equal(t, optimizer.SelectPlan, plan.Type)
			assert.Len(t, plan.Children, 1)

			// 验证投影算子
			projectProps, ok := plan.Properties.(*optimizer.SelectProperties)
			assert.True(t, ok)
			assert.Len(t, projectProps.Columns, 2)

			// 验证LIMIT算子
			limitPlan := plan.Children[0]
			assert.Equal(t, optimizer.LimitPlan, limitPlan.Type)
			limitProps, ok := limitPlan.Properties.(*optimizer.LimitProperties)
			assert.True(t, ok)
			assert.Equal(t, int64(10), limitProps.Limit)

			// 验证ORDER BY算子
			orderByPlan := limitPlan.Children[0]
			assert.Equal(t, optimizer.OrderPlan, orderByPlan.Type)
			orderByProps, ok := orderByPlan.Properties.(*optimizer.OrderByProperties)
			assert.True(t, ok)
			assert.Len(t, orderByProps.OrderKeys, 1)
			assert.Equal(t, "DESC", orderByProps.OrderKeys[0].Direction)

			// 验证HAVING算子
			havingPlan := orderByPlan.Children[0]
			assert.Equal(t, optimizer.HavingPlan, havingPlan.Type)
			havingProps, ok := havingPlan.Properties.(*optimizer.HavingProperties)
			assert.True(t, ok)
			assert.NotNil(t, havingProps.Condition)

			// 验证GROUP BY算子
			groupPlan := havingPlan.Children[0]
			assert.Equal(t, optimizer.GroupPlan, groupPlan.Type)
			groupProps, ok := groupPlan.Properties.(*optimizer.GroupByProperties)
			assert.True(t, ok)
			assert.Contains(t, groupProps.GroupKeys[0].Column, "department")

			// 验证Filter算子
			filterPlan := groupPlan.Children[0]
			assert.Equal(t, optimizer.FilterPlan, filterPlan.Type)
			filterProps, ok := filterPlan.Properties.(*optimizer.FilterProperties)
			assert.True(t, ok)
			assert.NotNil(t, filterProps.Condition)

			// 验证表扫描算子
			scanPlan := filterPlan.Children[0]
			assert.Equal(t, optimizer.TableScanPlan, scanPlan.Type)
			scanProps, ok := scanPlan.Properties.(*optimizer.TableScanProperties)
			assert.True(t, ok)
			assert.Equal(t, "employees", scanProps.Table)
		})

		// 测试多表JOIN和子查询
		t.Run("SelectWithMultipleJoinsAndSubquery", func(t *testing.T) {
			sql := `
				SELECT u.name, d.department_name, o.order_id 
				FROM users u 
				JOIN departments d ON u.dept_id = d.id 
				LEFT JOIN orders o ON u.id = o.user_id 
				WHERE u.status = 'active'
			`
			stmt, err := parser.Parse(sql)
			assert.NoError(t, err)

			opt := optimizer.NewOptimizer()
			plan, err := opt.Optimize(stmt)
			assert.NoError(t, err)
			assert.NotNil(t, plan)

			// 验证计划树结构：
			// Project(u.name, d.department_name, o.order_id)
			//   └─ Filter(u.status = 'active')
			//        └─ Join(LEFT, u.id = o.user_id)
			//             ├─ Join(INNER, u.dept_id = d.id)
			//             │    ├─ TableScan(users AS u)
			//             │    └─ TableScan(departments AS d)
			//             └─ TableScan(orders AS o)
			assert.Equal(t, optimizer.SelectPlan, plan.Type)
			assert.Len(t, plan.Children, 1)

			// 验证投影算子
			projectProps, ok := plan.Properties.(*optimizer.SelectProperties)
			assert.True(t, ok)
			assert.Len(t, projectProps.Columns, 3)

			// 验证Filter算子
			filterPlan := plan.Children[0]
			assert.Equal(t, optimizer.FilterPlan, filterPlan.Type)
			filterProps, ok := filterPlan.Properties.(*optimizer.FilterProperties)
			assert.True(t, ok)
			assert.NotNil(t, filterProps.Condition)

			// 验证第一个JOIN算子（LEFT JOIN）
			leftJoinPlan := filterPlan.Children[0]
			assert.Equal(t, optimizer.JoinPlan, leftJoinPlan.Type)
			leftJoinProps, ok := leftJoinPlan.Properties.(*optimizer.JoinProperties)
			assert.True(t, ok)
			assert.Equal(t, "LEFT", leftJoinProps.JoinType)

			// 验证第二个JOIN算子（INNER JOIN）
			innerJoinPlan := leftJoinPlan.Children[0]
			assert.Equal(t, optimizer.JoinPlan, innerJoinPlan.Type)
			innerJoinProps, ok := innerJoinPlan.Properties.(*optimizer.JoinProperties)
			assert.True(t, ok)
			assert.Equal(t, "INNER", innerJoinProps.JoinType)

			// 验证表扫描算子
			usersScan := innerJoinPlan.Children[0]
			assert.Equal(t, optimizer.TableScanPlan, usersScan.Type)
			usersProps, ok := usersScan.Properties.(*optimizer.TableScanProperties)
			assert.True(t, ok)
			assert.Equal(t, "users", usersProps.Table)
			assert.Equal(t, "u", usersProps.TableAlias)
		})

		// 测试DDL语句的优化
		t.Run("TestDDLStatements", func(t *testing.T) {
			// 测试CREATE DATABASE语句
			t.Run("CreateDatabase", func(t *testing.T) {
				sql := "CREATE DATABASE test_db"
				stmt, err := parser.Parse(sql)
				assert.NoError(t, err)

				opt := optimizer.NewOptimizer()
				plan, err := opt.Optimize(stmt)
				assert.NoError(t, err)
				assert.NotNil(t, plan)

				// 验证计划类型
				assert.Equal(t, optimizer.CreateDatabasePlan, plan.Type)
				props, ok := plan.Properties.(*optimizer.CreateDatabaseProperties)
				assert.True(t, ok)
				assert.Equal(t, "test_db", props.Database)
			})

			// 测试CREATE TABLE语句
			t.Run("CreateTable", func(t *testing.T) {
				sql := `CREATE TABLE users (
					id INTEGER PRIMARY KEY,
					name VARCHAR(255) NOT NULL,
					age INTEGER
				)`
				stmt, err := parser.Parse(sql)
				assert.NoError(t, err)

				opt := optimizer.NewOptimizer()
				plan, err := opt.Optimize(stmt)
				assert.NoError(t, err)
				assert.NotNil(t, plan)

				// 验证计划类型
				assert.Equal(t, optimizer.CreateTablePlan, plan.Type)
				props, ok := plan.Properties.(*optimizer.CreateTableProperties)
				assert.True(t, ok)
				assert.Equal(t, "users", props.Table)
				assert.Len(t, props.Columns, 3)
			})

			// 测试DROP DATABASE语句
			t.Run("DropDatabase", func(t *testing.T) {
				sql := "DROP DATABASE test_db"
				stmt, err := parser.Parse(sql)
				assert.NoError(t, err)

				opt := optimizer.NewOptimizer()
				plan, err := opt.Optimize(stmt)
				assert.NoError(t, err)
				assert.NotNil(t, plan)

				// 验证计划类型
				assert.Equal(t, optimizer.DropDatabasePlan, plan.Type)
				props, ok := plan.Properties.(*optimizer.DropDatabaseProperties)
				assert.True(t, ok)
				assert.Equal(t, "test_db", props.Database)
			})

			// 测试DROP TABLE语句
			t.Run("DropTable", func(t *testing.T) {
				sql := "DROP TABLE users"
				stmt, err := parser.Parse(sql)
				assert.NoError(t, err)

				opt := optimizer.NewOptimizer()
				plan, err := opt.Optimize(stmt)
				assert.NoError(t, err)
				assert.NotNil(t, plan)

				// 验证计划类型
				assert.Equal(t, optimizer.DropTablePlan, plan.Type)
				props, ok := plan.Properties.(*optimizer.DropTableProperties)
				assert.True(t, ok)
				assert.Equal(t, "users", props.Table)
			})
		})

		// 测试DCL语句的优化
		t.Run("TestDCLStatements", func(t *testing.T) {
			// 测试BEGIN TRANSACTION语句
			t.Run("BeginTransaction", func(t *testing.T) {
				sql := "START TRANSACTION"
				stmt, err := parser.Parse(sql)
				assert.NoError(t, err)

				opt := optimizer.NewOptimizer()
				plan, err := opt.Optimize(stmt)
				assert.NoError(t, err)
				assert.NotNil(t, plan)

				// 验证计划类型
				assert.Equal(t, optimizer.TransactionPlan, plan.Type)
				props, ok := plan.Properties.(*optimizer.TransactionProperties)
				assert.True(t, ok)
				assert.Equal(t, "BEGIN", props.Type)
			})

			// 测试COMMIT语句
			t.Run("CommitTransaction", func(t *testing.T) {
				sql := "COMMIT"
				stmt, err := parser.Parse(sql)
				assert.NoError(t, err)

				opt := optimizer.NewOptimizer()
				plan, err := opt.Optimize(stmt)
				assert.NoError(t, err)
				assert.NotNil(t, plan)

				// 验证计划类型
				assert.Equal(t, optimizer.TransactionPlan, plan.Type)
				props, ok := plan.Properties.(*optimizer.TransactionProperties)
				assert.True(t, ok)
				assert.Equal(t, "COMMIT", props.Type)
			})
		})

		// 测试Utility语句的优化
		t.Run("TestUtilityStatements", func(t *testing.T) {
			// 测试USE语句
			t.Run("UseDatabase", func(t *testing.T) {
				sql := "USE test_db"
				stmt, err := parser.Parse(sql)
				assert.NoError(t, err)

				opt := optimizer.NewOptimizer()
				plan, err := opt.Optimize(stmt)
				assert.NoError(t, err)
				assert.NotNil(t, plan)

				// 验证计划类型
				assert.Equal(t, optimizer.UsePlan, plan.Type)
				props, ok := plan.Properties.(*optimizer.UseProperties)
				assert.True(t, ok)
				assert.Equal(t, "test_db", props.Database)
			})

			// 测试SHOW DATABASES语句
			t.Run("ShowDatabases", func(t *testing.T) {
				sql := "SHOW DATABASES"
				stmt, err := parser.Parse(sql)
				assert.NoError(t, err)

				opt := optimizer.NewOptimizer()
				plan, err := opt.Optimize(stmt)
				assert.NoError(t, err)
				assert.NotNil(t, plan)

				// 验证计划类型
				assert.Equal(t, optimizer.ShowPlan, plan.Type)
				props, ok := plan.Properties.(*optimizer.ShowProperties)
				assert.True(t, ok)
				assert.Equal(t, "DATABASES", props.Type)
			})

			// 测试EXPLAIN语句
			t.Run("ExplainStatement", func(t *testing.T) {
				sql := "EXPLAIN SELECT * FROM users"
				stmt, err := parser.Parse(sql)
				assert.NoError(t, err)

				opt := optimizer.NewOptimizer()
				plan, err := opt.Optimize(stmt)
				assert.NoError(t, err)
				assert.NotNil(t, plan)

				// 验证计划类型
				assert.Equal(t, optimizer.ExplainPlan, plan.Type)
				props, ok := plan.Properties.(*optimizer.ExplainProperties)
				assert.True(t, ok)
				assert.NotNil(t, props.Query)

				// 验证被解释的查询计划
				queryPlan := props.Query
				assert.Equal(t, optimizer.SelectPlan, queryPlan.Type)
			})
		})
	})
}
