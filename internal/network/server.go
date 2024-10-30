package network

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/yyun543/minidb/internal/executor"
)

// Server 表示数据库服务器
type Server struct {
	listener    net.Listener       // TCP监听器
	executor    *executor.Executor // SQL执行器
	connections sync.WaitGroup     // 活动连接计数
	maxConns    int                // 最大连接数
	connCount   int32              // 当前连接数
	ctx         context.Context    // 用于优雅关闭
	cancel      context.CancelFunc // 取消函数
	addr        string             // 服务器地址
}

// NewServer 创建新的服务器实例
func NewServer(ctx context.Context, addr string, maxConns int, executor *executor.Executor) (*Server, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to start server: %v", err)
	}

	ctx, cancel := context.WithCancel(ctx)

	return &Server{
		listener: listener,
		executor: executor,
		ctx:      ctx,
		cancel:   cancel,
		maxConns: maxConns,
		addr:     addr,
	}, nil
}

// Start 启动服务器
func (s *Server) Start() error {
	defer s.listener.Close()

	// 启动连接清理goroutine
	go s.cleanupConnections()

	log.Printf("Server listening on %s", s.addr)

	for {
		select {
		case <-s.ctx.Done():
			return nil
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				if strings.Contains(err.Error(), "use of closed network connection") {
					return nil
				}
				log.Printf("Accept error: %v", err)
				continue
			}

			// 检查连接数限制
			if s.connCount >= int32(s.maxConns) {
				conn.Close()
				log.Printf("Connection rejected: max connections reached")
				continue
			}

			// 处理新连接
			atomic.AddInt32(&s.connCount, 1)
			s.connections.Add(1)
			go s.handleConnection(conn)
		}
	}
}

// Stop 优雅关闭服务器
func (s *Server) Stop(ctx context.Context) error {
	s.cancel()
	s.listener.Close()

	// 等待所有连接处理完成
	done := make(chan struct{})
	go func() {
		s.connections.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("shutdown timeout")
	}
}

// handleConnection 处理单个客户端连接
func (s *Server) handleConnection(conn net.Conn) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in connection handler: %v", r)
		}
		conn.Close()
		s.connections.Done()
		atomic.AddInt32(&s.connCount, -1)
	}()

	reader := bufio.NewReader(conn)

	for {
		// 设置读取超时
		conn.SetReadDeadline(time.Now().Add(time.Minute))

		// 读取命令
		command, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				log.Printf("Read error: %v", err)
			}
			return
		}

		// 处理命令
		command = strings.TrimSpace(command)
		if command == "" {
			continue
		}

		// 处理退出命令
		if strings.ToLower(command) == "quit" || strings.ToLower(command) == "exit" {
			return
		}

		// 执行SQL命令
		result, err := s.executor.Execute(command)
		if err != nil {
			result = fmt.Sprintf("Error: %v", err)
		}

		// 发送响应
		response := fmt.Sprintf("%s\n", result)
		if _, err := conn.Write([]byte(response)); err != nil {
			log.Printf("Write error: %v", err)
			return
		}
	}
}

// cleanupConnections 定期清理空闲连接
func (s *Server) cleanupConnections() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			// 这里可以添加空闲连接清理逻辑
			log.Printf("Active connections: %d", atomic.LoadInt32(&s.connCount))
		}
	}
}

// Client 表示数据库客户端
type Client struct {
	conn   net.Conn
	config ClientConfig
}

// ClientConfig 客户端配置
type ClientConfig struct {
	Address    string
	Timeout    time.Duration
	MaxRetries int
	RetryDelay time.Duration
}

// NewClient 创建新的客户端实例
func NewClient(config ClientConfig) *Client {
	return &Client{
		config: config,
	}
}

// Connect 连接到服务器
func (c *Client) Connect() error {
	var err error
	for i := 0; i < c.config.MaxRetries; i++ {
		c.conn, err = net.DialTimeout("tcp", c.config.Address, c.config.Timeout)
		if err == nil {
			return nil
		}
		time.Sleep(c.config.RetryDelay)
	}
	return fmt.Errorf("failed to connect after %d retries: %v", c.config.MaxRetries, err)
}

// Execute 执行SQL命令
func (c *Client) Execute(query string) (string, error) {
	if c.conn == nil {
		return "", fmt.Errorf("not connected")
	}

	// 设置读写超时
	c.conn.SetDeadline(time.Now().Add(c.config.Timeout))

	// 发送查询
	if _, err := fmt.Fprintf(c.conn, "%s\n", query); err != nil {
		return "", fmt.Errorf("failed to send query: %v", err)
	}

	// 读取响应
	reader := bufio.NewReader(c.conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	return strings.TrimSpace(response), nil
}

// Close 关闭客户端连接
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
