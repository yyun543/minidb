package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/yyun543/minidb/internal/logger"
	"go.uber.org/zap"
)

var (
	host = flag.String("host", "localhost", "Host to bind to")
	port = flag.String("port", "7205", "Port to bind to")
	help = flag.Bool("h", false, "Show help")
)

func main() {
	flag.Parse()

	if *help {
		printUsage()
		return
	}

	// 初始化日志系统
	logConfig := logger.ConfigFromEnv()
	if err := logger.InitLogger(logConfig); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// 创建全局查询处理器
	handler, err := NewQueryHandler()
	if err != nil {
		logger.Fatal("Failed to create query handler", zap.Error(err))
	}
	defer handler.Close()

	// 启动TCP服务器
	address := *host + ":" + *port
	listener, err := net.Listen("tcp", address)
	if err != nil {
		logger.Fatal("Unable to start server",
			zap.String("address", address),
			zap.Error(err))
	}
	defer listener.Close()

	logger.LogServerEvent("server_starting",
		zap.String("version", "1.0 (MPP Optimized)"),
		zap.String("address", address),
		zap.Strings("features", []string{"Vectorized Execution", "Cost-based Optimization", "Statistics Collection"}))

	fmt.Printf("=== MiniDB Server ===\n")
	fmt.Printf("Version: 1.0 (MPP Optimized)\n")
	fmt.Printf("Listening on: %s\n", address)
	fmt.Printf("Features: Vectorized Execution, Cost-based Optimization, Statistics Collection\n")
	fmt.Printf("Ready for connections...\n\n")

	// 设置信号处理
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// 在新协程中处理连接
	go acceptConnections(listener, handler)

	// 等待停止信号
	<-signalChan
	logger.LogServerEvent("server_stopping")
	fmt.Println("\nShutting down server...")
}

// acceptConnections 接受客户端连接
func acceptConnections(listener net.Listener, handler *QueryHandler) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error("Failed to accept connection", zap.Error(err))
			continue
		}

		// 为每个连接创建一个新的goroutine
		go handleConnection(conn, handler)
	}
}

// printUsage 打印使用说明
func printUsage() {
	fmt.Printf("MiniDB - A lightweight MPP database system\n\n")
	fmt.Printf("Usage: %s [options]\n\n", os.Args[0])
	fmt.Printf("Options:\n")
	flag.PrintDefaults()
	fmt.Printf("\nExamples:\n")
	fmt.Printf("  %s                    # Start on default host:port (localhost:7205)\n", os.Args[0])
	fmt.Printf("  %s -port 8080         # Start on port 8080\n", os.Args[0])
	fmt.Printf("  %s -host 0.0.0.0      # Bind to all interfaces\n", os.Args[0])
}

func handleConnection(conn net.Conn, handler *QueryHandler) {
	defer conn.Close()

	// 获取客户端地址
	clientAddr := conn.RemoteAddr().String()

	// 创建新会话
	session := handler.sessionManager.CreateSession()
	sessionID := session.ID

	// 记录连接事件
	logger.LogConnectionEvent("client_connected", clientAddr, sessionID)

	// 在连接关闭时清理会话
	defer func() {
		handler.sessionManager.DeleteSession(sessionID)
		logger.LogConnectionEvent("client_disconnected", clientAddr, sessionID)
	}()

	// 发送欢迎消息
	welcomeMsg := fmt.Sprintf("Welcome to MiniDB v1.0!\n")
	welcomeMsg += fmt.Sprintf("Session ID: %d\n", sessionID)
	welcomeMsg += fmt.Sprintf("Type 'exit;' or 'quit;' to disconnect\n")
	welcomeMsg += fmt.Sprintf("------------------------------------\n")
	conn.Write([]byte(welcomeMsg))

	reader := bufio.NewReader(conn)

	// 读取客户端命令
	for {
		conn.Write([]byte("minidb> "))

		// 读取一行输入
		input, err := reader.ReadString(';')
		if err != nil {
			logger.WithClient(clientAddr).Error("Connection error",
				zap.Int64("session_id", sessionID),
				zap.Error(err))
			return
		}

		// 去除空白字符
		query := strings.TrimSpace(input)
		if query == "exit;" || query == "quit;" {
			conn.Write([]byte("Goodbye!\n"))
			return
		}

		// 空查询检查
		if query == "" || query == ";" {
			continue
		}

		// 记录查询（调试用）
		logger.WithClient(clientAddr).Debug("Query received",
			zap.Int64("session_id", sessionID),
			zap.String("query", query))

		// 使用会话ID处理查询
		result, err := handler.HandleQuery(sessionID, query)
		if err != nil {
			errorMsg := fmt.Sprintf("Error: %v\n", err)
			conn.Write([]byte(errorMsg))
			continue
		}

		// 发送结果
		conn.Write([]byte(result + "\n"))
	}
}
