package session

import (
	"sync"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/yyun543/minidb/internal/logger"
	"go.uber.org/zap"
)

// Session 表示一个数据库会话
type Session struct {
	ID           int64                  // 会话ID
	CurrentDB    string                 // 当前数据库
	CreatedAt    time.Time              // 创建时间
	LastAccessAt time.Time              // 最后访问时间
	Variables    map[string]interface{} // 会话变量
}

// SessionManager 会话管理器
type SessionManager struct {
	sessions sync.Map        // 会话存储
	node     *snowflake.Node // Snowflake节点
	mu       sync.Mutex      // 互斥锁
}

// NewSessionManager 创建新的会话管理器
func NewSessionManager() (*SessionManager, error) {
	logger.WithComponent("session").Info("Creating new session manager")

	// 创建Snowflake节点，节点ID使用1
	node, err := snowflake.NewNode(1)
	if err != nil {
		logger.WithComponent("session").Error("Failed to create Snowflake node for session manager",
			zap.Error(err))
		return nil, err
	}

	manager := &SessionManager{
		node: node,
	}

	logger.WithComponent("session").Info("Session manager created successfully",
		zap.Int64("snowflake_node_id", 1))

	return manager, nil
}

// CreateSession 创建新会话
func (m *SessionManager) CreateSession() *Session {
	session := &Session{
		ID:           m.node.Generate().Int64(),
		CreatedAt:    time.Now(),
		LastAccessAt: time.Now(),
		Variables:    make(map[string]interface{}),
	}
	m.sessions.Store(session.ID, session)

	logger.WithComponent("session").Info("New session created",
		zap.Int64("session_id", session.ID),
		zap.Time("created_at", session.CreatedAt))

	return session
}

// GetSession 获取会话
func (m *SessionManager) GetSession(id int64) (*Session, bool) {
	if value, ok := m.sessions.Load(id); ok {
		session := value.(*Session)
		session.LastAccessAt = time.Now()
		return session, true
	}
	return nil, false
}

// DeleteSession 删除会话
func (m *SessionManager) DeleteSession(id int64) {
	m.sessions.Delete(id)
}

// CleanupExpiredSessions 清理过期会话
func (m *SessionManager) CleanupExpiredSessions(timeout time.Duration) {
	m.sessions.Range(func(key, value interface{}) bool {
		session := value.(*Session)
		if time.Since(session.LastAccessAt) > timeout {
			m.sessions.Delete(key)
		}
		return true
	})
}
