package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/yyun543/minidb/internal/cache"
	"github.com/yyun543/minidb/internal/executor"
	"github.com/yyun543/minidb/internal/network"
	"github.com/yyun543/minidb/internal/parser"
	"github.com/yyun543/minidb/internal/storage"
)

func main() {
	// Parse command line flags
	port := flag.String("port", "8086", "Port to listen on")
	flag.Parse()

	// Create components
	engine := storage.NewEngine()
	parser := parser.NewParser()
	executor := executor.NewExecutor(engine, cache.NewCache())

	// Create server
	server, err := network.NewServer(":"+*port, parser, executor)
	if err != nil {
		log.Fatal(err)
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down...")
		server.Stop()
	}()

	// Start server
	fmt.Printf("MiniDB server starting on port %s...\n", *port)
	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}
