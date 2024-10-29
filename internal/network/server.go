package network

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/yyun543/minidb/internal/executor"
)

type Server struct {
	listener    net.Listener
	executor    *executor.Executor
	connections sync.WaitGroup
	maxConns    int
	connCount   int32
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewServer(address string, executor *executor.Executor) (*Server, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to start server: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Server{
		listener: listener,
		executor: executor,
		ctx:      ctx,
		cancel:   cancel,
	}, nil
}

func (s *Server) Start() error {
	defer s.listener.Close()

	fmt.Printf("MiniDB server listening on %s\n", s.listener.Addr())

	go s.acceptConnections()

	<-s.ctx.Done()
	return nil
}

func (s *Server) acceptConnections() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.ctx.Done():
				return
			default:
				fmt.Printf("Accept error: %v\n", err)
				continue
			}
		}

		s.connections.Add(1)
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	if atomic.LoadInt32(&s.connCount) >= int32(s.maxConns) {
		conn.Close()
		return
	}
	atomic.AddInt32(&s.connCount, 1)
	defer atomic.AddInt32(&s.connCount, -1)

	defer func() {
		conn.Close()
		s.connections.Done()
	}()

	// 设置连接超时
	conn.SetDeadline(time.Now().Add(time.Hour))

	reader := bufio.NewReader(conn)
	fmt.Fprintf(conn, "Welcome to MiniDB! Enter your SQL query or type 'exit' to quit.\n")

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			fmt.Fprintf(conn, "minidb> ")
			query, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					fmt.Printf("Read error: %v\n", err)
				}
				return
			}

			query = strings.TrimSpace(query)
			if query == "" {
				continue
			}

			if strings.ToLower(query) == "exit" || strings.ToLower(query) == "quit" {
				fmt.Fprintf(conn, "Bye!\n")
				return
			}

			// 重置连接超时
			conn.SetDeadline(time.Now().Add(time.Hour))

			// 执行查询
			result, err := s.executor.Execute(query)
			if err != nil {
				fmt.Fprintf(conn, "Error: %v\n", err)
				continue
			}

			fmt.Fprintf(conn, "%s\n", result)
		}
	}
}

func (s *Server) Stop() error {
	s.cancel()

	// 关闭监听器
	if err := s.listener.Close(); err != nil {
		return fmt.Errorf("failed to close listener: %v", err)
	}

	// 等待所有连接处理完成
	done := make(chan struct{})
	go func() {
		s.connections.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-time.After(10 * time.Second):
		return fmt.Errorf("timeout waiting for connections to close")
	}
}

type ClientConfig struct {
	Address    string
	Timeout    time.Duration
	MaxRetries int
	RetryDelay time.Duration
}

type Client struct {
	conn   net.Conn
	config ClientConfig
}

func NewClient(config ClientConfig) *Client {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = time.Second
	}
	return &Client{config: config}
}

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

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
