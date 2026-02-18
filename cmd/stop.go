package cmd

import (
	"fmt"
	"os"
	"signet/config"
	"syscall"
)

// RunStop は `signet stop` コマンドを実行する
func RunStop(args []string) {
	// 設定読み込み
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load config: %v\n", err)
		os.Exit(1)
	}

	// PIDファイルパス
	pidPath := cfg.PIDFilePath()

	// PIDファイル読み込み
	pidData, err := os.ReadFile(pidPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, "Error: PID file not found. Is the node running?")
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Error: failed to read PID file: %v\n", err)
		os.Exit(1)
	}

	var pid int
	_, err = fmt.Sscanf(string(pidData), "%d", &pid)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid PID format: %v\n", err)
		os.Exit(1)
	}

	// プロセスが存在するか確認
	process, err := os.FindProcess(pid)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find process: %v\n", err)
		os.Exit(1)
	}

	// SIGTERM送信
	if err := process.Signal(syscall.SIGTERM); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to send SIGTERM: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Sent SIGTERM to process %d\n", pid)

	// PIDファイル削除
	if err := os.Remove(pidPath); err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Warning: failed to remove PID file: %v\n", err)
	}
}
