package network

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
	"io"

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
	
	// Set connection timeout
	conn.SetDeadline(time.Now().Add(60 * time.Second))
	
	reader := bufio.NewReader(conn)
	fmt.Fprintf(conn, "Welcome to MiniDB! Enter your SQL query or type 'exit' to quit.\n")
	
	for {
		fmt.Fprintf(conn, "minidb> ")
		query, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				fmt.Println("Error reading from connection:", err)
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
		
		// Reset connection timeout
		conn.SetDeadline(time.Now().Add(30 * time.Second))
		
		parsedQuery, err := s.parser.Parse(query)
		if err != nil {
			fmt.Fprintf(conn, "Error: %v\n", err)
			continue
		}
		
		result, err := s.executor.Execute(parsedQuery)
		if err != nil {
			fmt.Fprintf(conn, "Error: %v\n", err)
			continue
		}
		
		fmt.Fprintf(conn, "%s\n", result)
	}
}
