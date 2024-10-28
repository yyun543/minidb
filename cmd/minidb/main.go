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
	// Create storage engine
	engine := storage.NewEngine()

	// Create parser
	parser := parser.NewParser()

	// Create executor
	executor := executor.NewExecutor(engine)

	// Create server
	server, err := network.NewServer(":8086", parser, executor)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("MiniDB server starting...")
	err = server.Start()
	if err != nil {
		log.Fatal(err)
	}
}
