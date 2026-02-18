package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"signet/config"
)

// NodeStore はノード情報の永続化を担当する
type NodeStore struct {
	dir string // nodesディレクトリパス
}

// NewNodeStore は新しいNodeStoreを作成する
func NewNodeStore(dir string) *NodeStore {
	return &NodeStore{dir: dir}
}

// Save はノード情報をファイルに保存する
func (s *NodeStore) Save(nodeName string, info *NodeInfo) error {
	// ディレクトリが存在しない場合は作成
	if err := os.MkdirAll(s.dir, 0755); err != nil {
		return fmt.Errorf("failed to create nodes directory: %w", err)
	}

	// TOML形式で保存
	content := fmt.Sprintf("NickName = %s\n", info.NickName)
	content += fmt.Sprintf("Address = %s\n", info.Address)
	content += fmt.Sprintf("Ed25519PublicKey = %s\n", info.PublicKey)

	filePath := filepath.Join(s.dir, nodeName)
	if err := writeFile(filePath, content); err != nil {
		return fmt.Errorf("failed to write node file: %w", err)
	}

	return nil
}

// Load は指定されたノード名の情報を読み込む
func (s *NodeStore) Load(nodeName string) (*NodeInfo, error) {
	filePath := filepath.Join(s.dir, nodeName)

	values, err := config.ParseTOMLFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse node file: %w", err)
	}

	info := &NodeInfo{
		Name:      nodeName,
		NickName:  values["NickName"],
		Address:   values["Address"],
		PublicKey: values["Ed25519PublicKey"],
	}

	return info, nil
}

// LoadAll はディレクトリ内の全ノードファイルを読み込む
func (s *NodeStore) LoadAll() (map[string]*NodeInfo, error) {
	// ディレクトリが存在しない場合は空マップを返す
	if _, err := os.Stat(s.dir); os.IsNotExist(err) {
		return make(map[string]*NodeInfo), nil
	}

	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read nodes directory: %w", err)
	}

	result := make(map[string]*NodeInfo)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		nodeName := entry.Name()
		info, err := s.Load(nodeName)
		if err != nil {
			return nil, fmt.Errorf("failed to load node %s: %w", nodeName, err)
		}
		result[nodeName] = info
	}

	return result, nil
}

// Delete は指定されたノード名の情報を削除する
func (s *NodeStore) Delete(nodeName string) error {
	filePath := filepath.Join(s.dir, nodeName)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete node file: %w", err)
	}
	return nil
}

// Exists は指定されたノードが存在するかを確認する
func (s *NodeStore) Exists(nodeName string) bool {
	filePath := filepath.Join(s.dir, nodeName)
	_, err := os.Stat(filePath)
	return err == nil
}
