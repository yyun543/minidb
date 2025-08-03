package catalog

import (
	"fmt"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/yyun543/minidb/internal/storage"
)

// Catalog 基于SQL统一管理的catalog实现
// 遵循用户要求：完全通过SQL管理，不实现额外的代码管理机制
type Catalog struct {
	*SimpleSQLCatalog
}

// NewCatalog 创建新的Catalog实例
// 注意：这个实例需要在有SQL执行器可用后才能完全初始化
func NewCatalog(engine storage.Engine) *Catalog {
	// 创建一个基础实例，稍后通过SetSQLRunner设置SQL执行器
	simpleCatalog := NewSimpleSQLCatalog(engine)

	return &Catalog{
		SimpleSQLCatalog: simpleCatalog,
	}
}

// SetSQLRunner 设置SQL执行器
// 这允许catalog在SQL执行器创建后进行完整初始化
func (c *Catalog) SetSQLRunner(sqlRunner SQLRunner) {
	c.SimpleSQLCatalog.SetSQLRunner(sqlRunner)
}

// InitWithSQLRunner 创建带有SQL执行器的catalog并初始化
func InitWithSQLRunner(engine storage.Engine, sqlRunner SQLRunner) (*Catalog, error) {
	simpleCatalog := NewSimpleSQLCatalog(engine)
	simpleCatalog.SetSQLRunner(sqlRunner)

	catalog := &Catalog{
		SimpleSQLCatalog: simpleCatalog,
	}

	if err := catalog.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize SQL-based catalog: %w", err)
	}

	return catalog, nil
}

// CreateCatalogWithExecutor 使用执行器创建catalog的便利方法（暂时禁用）
// func CreateCatalogWithExecutor(engine storage.Engine, executor Executor, sessionManager SessionManager) (*Catalog, error) {
//     // 简化：直接创建不带SQL执行器的catalog
//     catalog := NewCatalog(engine)
//     return catalog, catalog.Init()
// }

// LegacyInit 传统初始化方法（用于向后兼容）
// 当SQL执行器不可用时的临时初始化
func (c *Catalog) LegacyInit() error {
	// 如果没有SQL执行器，使用最简单的系统表初始化
	if c.sqlRunner == nil {
		return c.initSystemTablesDirectly()
	}
	return c.Init()
}

// initSystemTablesDirectly 直接初始化系统表（无SQL执行器时的备选方案）
func (c *Catalog) initSystemTablesDirectly() error {
	// 这里可以调用原来的系统表初始化代码
	// 但是按照用户要求，我们应该尽量避免这种代码管理机制
	return fmt.Errorf("SQL runner not set - cannot initialize catalog without SQL execution capability")
}

// 确保向后兼容的类型定义
type DatabaseMeta struct {
	Name string
}

type TableMeta struct {
	Database   string
	Table      string
	ChunkCount int64
	Schema     *arrow.Schema
}

// SessionManager 接口定义（用于适配器）
type SessionManager interface {
	CreateSession() Session
	CloseSession(id string)
}

// Session 接口定义
type Session interface {
	GetID() string
}

// SimpleSession 简单会话实现
type SimpleSession struct {
	ID string
}

func (s *SimpleSession) GetID() string {
	return s.ID
}

// SimpleSessionManager 简单会话管理器实现
type SimpleSessionManager struct {
	counter int64
}

func NewSimpleSessionManager() *SimpleSessionManager {
	return &SimpleSessionManager{}
}

func (sm *SimpleSessionManager) CreateSession() Session {
	sm.counter++
	return &SimpleSession{ID: fmt.Sprintf("session-%d", sm.counter)}
}

func (sm *SimpleSessionManager) CloseSession(id string) {
	// 简单实现，实际可以添加清理逻辑
}

// NullSQLRunner 空的SQL执行器（用于测试或临时使用）
type NullSQLRunner struct{}

func (n *NullSQLRunner) ExecuteSQL(sql string) (arrow.Record, error) {
	return nil, fmt.Errorf("null SQL runner - no actual execution capability")
}

// CreateTemporaryCatalog 创建临时catalog（用于测试）
func CreateTemporaryCatalog(engine storage.Engine) *Catalog {
	catalog := NewCatalog(engine)
	catalog.SetSQLRunner(&NullSQLRunner{})

	// 初始化catalog（包含WAL恢复）
	if err := catalog.Init(); err != nil {
		// 如果初始化失败，记录错误但继续使用
		fmt.Printf("Warning: catalog initialization failed: %v\n", err)
	}

	return catalog
}
