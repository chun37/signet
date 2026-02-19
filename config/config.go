package config

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultRootDir  = "/etc/signet"
	DefaultPort     = "8080"
	defaultConfPath = "/etc/signet/signet.conf"
)

// Config はアプリケーションの設定を表す
type Config struct {
	RootDir  string
	Address  string
	NickName string
	NodeName string
	Port     string
}

// LoadConfig はデフォルトパスから設定を読み込む
func LoadConfig() (*Config, error) {
	return LoadConfigFrom(defaultConfPath)
}

// LoadConfigFrom は指定パスから設定を読み込む
func LoadConfigFrom(path string) (*Config, error) {
	cfg := &Config{
		RootDir: defaultRootDir,
		Port:    DefaultPort,
	}

	// 設定ファイルが存在しない場合はデフォルト値を返す
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil
	}

	values, err := ParseTOMLFile(path)
	if err != nil {
		return nil, err
	}

	if v, ok := values["RootDir"]; ok {
		cfg.RootDir = v
	}
	if v, ok := values["Address"]; ok {
		cfg.Address = v
	}
	if v, ok := values["NickName"]; ok {
		cfg.NickName = v
	}
	if v, ok := values["NodeName"]; ok {
		cfg.NodeName = v
	}
	if v, ok := values["Port"]; ok {
		cfg.Port = v
	}

	return cfg, nil
}

// PrivKeyPath は秘密鍵ファイルのパスを返す
func (c *Config) PrivKeyPath() string {
	return filepath.Join(c.RootDir, "ed25519.priv")
}

// BlockFilePath はブロックチェーンファイルのパスを返す
func (c *Config) BlockFilePath() string {
	return filepath.Join(c.RootDir, "block.jsonl")
}

// PendingFilePath は承認待ちトランザクションファイルのパスを返す
func (c *Config) PendingFilePath() string {
	return filepath.Join(c.RootDir, "pending_transaction.json")
}

// NodesDir はノード設定ディレクトリのパスを返す
func (c *Config) NodesDir() string {
	return filepath.Join(c.RootDir, "nodes")
}

// PIDFilePath はPIDファイルのパスを返す
func (c *Config) PIDFilePath() string {
	return filepath.Join(c.RootDir, "signet.pid")
}

// NodeFilePath は指定ノード名の設定ファイルパスを返す
func (c *Config) NodeFilePath(nodeName string) string {
	return filepath.Join(c.RootDir, "nodes", nodeName)
}

// ParseAddress はアドレス文字列からホストとポートをパースする
// 形式: "host:port" または "host" (デフォルトポート使用)
func ParseAddress(addr string) (host string, port string) {
	parts := strings.Split(addr, ":")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return addr, DefaultPort
}

// NormalizeAddress はアドレスにポートが含まれていなければデフォルトポートを付与する
// "192.168.1.1" → "192.168.1.1:8080", "192.168.1.1:9090" → "192.168.1.1:9090"
func NormalizeAddress(addr string) string {
	host, port := ParseAddress(addr)
	return host + ":" + port
}
