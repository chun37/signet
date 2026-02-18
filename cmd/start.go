package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"signet/config"
	"signet/node"
	"signet/p2p"
	"signet/server"
	"syscall"
	"time"
)

// RunStart は `signet start` コマンドを実行する
func RunStart(args []string) {
	// 設定読み込み
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error: failed to load config: %v", err)
	}

	// Node 初期化
	n, err := node.NewNode(cfg)
	if err != nil {
		log.Fatalf("Error: failed to initialize node: %v", err)
	}

	// ピアからチェーン同期
	peers, err := n.NodeStore.LoadAll()
	if err != nil {
		log.Printf("Warning: failed to load peers for sync: %v", err)
	} else {
		if len(peers) > 0 {
			log.Println("Syncing chain with peers...")
			if err := p2p.SyncChain(n.Chain, peers); err != nil {
				log.Printf("Warning: chain sync failed: %v", err)
			}
		}
	}

	// HTTPサーバー起動
	addr := fmt.Sprintf("%s:%s", cfg.Address, cfg.Port)
	srv := server.NewServer(addr, n)

	// サーバーをgoroutineで起動
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- srv.Start()
	}()

	// PIDファイル書き込み
	pid := os.Getpid()
	pidPath := cfg.PIDFilePath()
	if err := os.WriteFile(pidPath, []byte(fmt.Sprintf("%d\n", pid)), 0644); err != nil {
		log.Printf("Warning: failed to write PID file: %v", err)
	}

	log.Printf("Signet node started (PID: %d)", pid)
	log.Printf("Listening on %s", addr)

	// シグナルハンドリング
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// シャットダウン処理
	select {
	case err := <-serverErr:
		if err != nil {
			log.Fatalf("Server error: %v", err)
		}
	case sig := <-sigCh:
		log.Printf("Received signal: %v", sig)
		// Graceful shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Stop(ctx); err != nil {
			log.Printf("Warning: server shutdown error: %v", err)
		}

		// PIDファイル削除
		if err := os.Remove(pidPath); err != nil && !os.IsNotExist(err) {
			log.Printf("Warning: failed to remove PID file: %v", err)
		}

		log.Println("Signet node stopped")
	}
}
