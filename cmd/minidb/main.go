package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yyun543/minidb/internal/cache"
	"github.com/yyun543/minidb/internal/executor"
	"github.com/yyun543/minidb/internal/index"
	"github.com/yyun543/minidb/internal/network"
	"github.com/yyun543/minidb/internal/storage"
)

// 配置选项
type Config struct {
	Port     string
	MaxConns int
	CacheTTL time.Duration
}

func main() {
	// 1. 解析命令行参数
	config := parseFlags()

	// 2. 创建上下文用于优雅关闭
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 3. 初始化组件
	db, err := initializeDB(config)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 4. 启动服务器
	if err := startServer(ctx, db, config); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func parseFlags() *Config {
	config := &Config{}

	// 定义命令行参数
	flag.StringVar(&config.Port, "port", "8086", "Port to listen on")
	flag.IntVar(&config.MaxConns, "max-connections", 100, "Maximum number of concurrent connections")
	flag.DurationVar(&config.CacheTTL, "cache-ttl", 5*time.Minute, "Cache TTL duration")

	flag.Parse()
	return config
}

// 初始化数据库组件
func initializeDB(config *Config) (*executor.Executor, error) {
	// 创建存储引擎
	store := storage.NewEngine()

	// 创建索引管理器
	indexMgr := index.NewManager()

	// 创建查询缓存
	queryCache := cache.New(config.CacheTTL)

	// 创建执行器
	return executor.New(store, indexMgr, queryCache), nil
}

// 启动服务器
func startServer(ctx context.Context, db *executor.Executor, config *Config) error {
	// 创建服务器实例
	server, err := network.NewServer(
		ctx,
		":"+config.Port,
		config.MaxConns,
		db,
	)
	if err != nil {
		return fmt.Errorf("failed to create server: %v", err)
	}

	// 处理优雅关闭
	go handleShutdown(ctx, server)

	// 启动服务器
	fmt.Printf("MiniDB server starting on port %s...\n", config.Port)
	return server.Start()
}

// 处理优雅关闭
func handleShutdown(ctx context.Context, server *network.Server) {
	// 监听系统信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigChan:
		fmt.Println("\nInitiating graceful shutdown...")
	case <-ctx.Done():
		fmt.Println("\nContext cancelled, shutting down...")
	}

	// 给服务器一些时间完成当前请求
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Stop(shutdownCtx); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}
}
