package executor

import (
	"github.com/yyun543/minidb/internal/catalog"
	"github.com/yyun543/minidb/internal/session"
)

// Context 执行上下文
type Context struct {
	Session *session.Session
	catalog *catalog.Catalog // 元数据管理器
	// 可以添加更多上下文信息
}

// NewContext 创建执行上下文
func NewContext(cat *catalog.Catalog, sess *session.Session) *Context {
	return &Context{
		Session: sess,
		catalog: cat,
	}
}

// GetCatalog 获取元数据管理器
func (ctx *Context) GetCatalog() *catalog.Catalog {
	return ctx.catalog
}
