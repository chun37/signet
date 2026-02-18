package cmd

import (
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"signet/config"
	"signet/core"
	"signet/crypto"
	"signet/storage"
)

// RunInit は `signet init` コマンドを実行する
func RunInit(args []string) {
	// フラグ定義
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	addr := fs.String("address", "", "ノードのアドレス (例: localhost:8080)")
	nickname := fs.String("nickname", "", "ニックネーム")
	nodename := fs.String("nodename", "", "ノード名")

	if err := fs.Parse(args); err != nil {
		fs.Usage()
		os.Exit(1)
	}

	// 必須フラグチェック
	if *addr == "" || *nickname == "" || *nodename == "" {
		fmt.Fprintln(os.Stderr, "Error: --address, --nickname, --nodename are required")
		fs.Usage()
		os.Exit(1)
	}

	// 設定読み込み（デフォルト値でOK）
	cfg := &config.Config{
		RootDir:  "/etc/signet",
		Address:  *addr,
		NickName: *nickname,
		NodeName: *nodename,
		Port:     "8080",
	}

	// RootDir 作成
	if err := os.MkdirAll(cfg.RootDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create root directory: %v\n", err)
		os.Exit(1)
	}

	// nodes ディレクトリ作成
	if err := os.MkdirAll(cfg.NodesDir(), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create nodes directory: %v\n", err)
		os.Exit(1)
	}

	// Ed25519鍵ペア生成
	pubKey, privKey, err := crypto.GenerateKeyPair()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to generate key pair: %v\n", err)
		os.Exit(1)
	}

	// 秘密鍵を保存
	if err := crypto.SavePrivateKey(cfg.PrivKeyPath(), privKey); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to save private key: %v\n", err)
		os.Exit(1)
	}

	// ジェネシスブロック生成
	genesis := core.NewGenesisBlock()

	// block.jsonl に書き込み
	blockStore := storage.NewBlockStore(cfg.BlockFilePath())
	if err := blockStore.Append(genesis); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to write genesis block: %v\n", err)
		os.Exit(1)
	}

	// 自ノード情報をnodesディレクトリに保存
	nodeStore := storage.NewNodeStore(cfg.NodesDir())
	pubKeyHex := hex.EncodeToString(pubKey)
	nodeInfo := &storage.NodeInfo{
		Name:      *nodename,
		NickName:  *nickname,
		Address:   *addr,
		PublicKey: pubKeyHex,
	}
	if err := nodeStore.Save(*nodename, nodeInfo); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to save node info: %v\n", err)
		os.Exit(1)
	}

	// 設定ファイル保存
	if err := saveConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to save config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Signet node initialized successfully!")
	fmt.Printf("  Node Name: %s\n", *nodename)
	fmt.Printf("  Nick Name: %s\n", *nickname)
	fmt.Printf("  Address: %s\n", *addr)
	fmt.Printf("  Public Key: %s\n", pubKeyHex)
	fmt.Printf("  Config: %s\n", defaultConfigPath())
}

// saveConfig は設定をファイルに保存する
func saveConfig(cfg *config.Config) error {
	path := defaultConfigPath()
	content := fmt.Sprintf("RootDir = %s\n", cfg.RootDir)
	content += fmt.Sprintf("Address = %s\n", cfg.Address)
	content += fmt.Sprintf("NickName = %s\n", cfg.NickName)
	content += fmt.Sprintf("NodeName = %s\n", cfg.NodeName)
	content += fmt.Sprintf("Port = %s\n", cfg.Port)
	return os.WriteFile(path, []byte(content), 0644)
}

// defaultConfigPath はデフォルトの設定ファイルパスを返す
func defaultConfigPath() string {
	if path := os.Getenv("SIGNET_CONFIG"); path != "" {
		return path
	}
	return "/etc/signet/signet.conf"
}
