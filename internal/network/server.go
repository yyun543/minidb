package network

import (
	"bufio"
	"fmt"
	"net"
	"strings"

	"github.com/yyun543/minidb/internal/parser"
	"github.com/yyun543/minidb/internal/executor"
)

type Server struct {
	listener net.Listener
	parser   *parser.Parser
	executor *executor.Executor
}

func NewServer(address string, parser *parser.Parser, executor *executor.Executor) (*Server, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	return &Server{
		listener: listener,
		parser:   parser,
		executor: executor,
	}, nil
}

func (s *Server) Start() error {
	fmt.Println("Server started. Listening on", s.listener.Addr())
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	for {
		query, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from connection:", err)
			return
		}
		
		// 清理输入，移除多余的空白字符
		query = strings.TrimSpace(query)
		if query == "exit" || query == "exit;" {
			return
		}
		
		// 如果查询为空，继续下一次循环
		if query == "" {
			continue
		}
		
		parsedQuery, err := s.parser.Parse(query)
		if err != nil {
			fmt.Fprintf(conn, "Error parsing query: %v\n", err)
			continue
		}
		
		result, err := s.executor.Execute(parsedQuery)
		if err != nil {
			fmt.Fprintf(conn, "Error executing query: %v\n", err)
			continue
		}
		
		fmt.Fprintf(conn, "%s\n", result)
	}
}
