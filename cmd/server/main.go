package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

const (
	HOST = "localhost"
	PORT = "7205"
)

func main() {
	// 启动TCP服务器
	listener, err := net.Listen("tcp", HOST+":"+PORT)
	if err != nil {
		log.Fatalf("无法启动服务器: %v", err)
	}
	defer listener.Close()

	// 创建全局查询处理器
	handler, err := NewQueryHandler()
	if err != nil {
		log.Fatalf("创建查询处理器失败: %v", err)
	}

	fmt.Printf("MiniDB 服务器已启动，监听 %s:%s\n", HOST, PORT)

	// 接受客户端连接
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("接受连接失败: %v", err)
			continue
		}

		// 为每个连接创建一个新的goroutine
		go handleConnection(conn, handler)
	}
}

func handleConnection(conn net.Conn, handler *QueryHandler) {
	defer conn.Close()

	// 创建新会话
	session := handler.sessionManager.CreateSession()
	sessionID := session.ID

	// 在连接关闭时清理会话
	defer handler.sessionManager.DeleteSession(sessionID)

	// 发送欢迎消息
	conn.Write([]byte(fmt.Sprintf("欢迎使用 MiniDB! (会话ID: %d)\n", sessionID)))

	reader := bufio.NewReader(conn)

	// 读取客户端命令
	for {
		conn.Write([]byte("minidb> "))

		// 读取一行输入
		input, err := reader.ReadString(';')
		if err != nil {
			log.Printf("读取输入失败: %v", err)
			return
		}

		// 去除空白字符
		query := strings.TrimSpace(input)
		if query == "exit;" || query == "quit;" {
			conn.Write([]byte("再见!\n"))
			return
		}

		// 使用会话ID处理查询
		result, err := handler.HandleQuery(sessionID, query)
		if err != nil {
			conn.Write([]byte(fmt.Sprintf("错误: %v\n", err)))
			continue
		}

		// 发送结果
		conn.Write([]byte(result + "\n"))
	}
}
