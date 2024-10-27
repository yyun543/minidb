package main

import (
	"fmt"
	"log"

	"github.com/yyun543/minidb/internal/executor"
	"github.com/yyun543/minidb/internal/network"
	"github.com/yyun543/minidb/internal/parser"
	"github.com/yyun543/minidb/internal/storage"
)

func main() {
	// 创建存储引擎
	engine := storage.NewEngine()

	// 创建一个示例表
	err := engine.CreateTable("users")
	if err != nil {
		log.Fatal(err)
	}

	// 创建解析器
	parser := parser.NewParser()

	// 创建执行器
	executor := executor.NewExecutor(engine)

	// 创建服务器
	server, err := network.NewServer(":8080", parser, executor)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("MiniDB server starting...")
	err = server.Start()
	if err != nil {
		log.Fatal(err)
	}
}

