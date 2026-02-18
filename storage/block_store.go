package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"signet/core"
)

// BlockStore はブロックチェーンの永続化を担当する
type BlockStore struct {
	path string
}

// NewBlockStore は新しいBlockStoreを作成する
func NewBlockStore(path string) *BlockStore {
	return &BlockStore{path: path}
}

// LoadAll は全ブロックを読み込む
// ファイルが存在しない場合は空スライスを返す
func (s *BlockStore) LoadAll() ([]*core.Block, error) {
	// ファイルが存在しない場合は空スライスを返す
	_, err := os.Stat(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return []*core.Block{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	data, err := readFile(s.path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	if len(data) == 0 {
		return []*core.Block{}, nil
	}

	var blocks []*core.Block
	lines := splitLines(data)
	for i, line := range lines {
		if len(line) == 0 {
			continue
		}
		var block core.Block
		if err := json.Unmarshal(line, &block); err != nil {
			return nil, fmt.Errorf("failed to unmarshal block at line %d: %w", i+1, err)
		}
		blocks = append(blocks, &block)
	}

	return blocks, nil
}

// Append はブロックを1行追記する
func (s *BlockStore) Append(b *core.Block) error {
	data, err := json.Marshal(b)
	if err != nil {
		return fmt.Errorf("failed to marshal block: %w", err)
	}

	// 改行を追加して追記
	data = append(data, '\n')
	if err := appendFile(s.path, data); err != nil {
		return fmt.Errorf("failed to append to file: %w", err)
	}

	return nil
}

// ReplaceAll は全ブロックを書き直す（最長チェーンルール用）
// 一時ファイルに書いてrenameすることでアトミック性を確保
func (s *BlockStore) ReplaceAll(blocks []*core.Block) error {
	// 一時ファイルパス
	tmpPath := s.path + ".tmp"

	// 一時ファイルを開く
	f, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer f.Close()

	// 全ブロックを書き込み
	for _, b := range blocks {
		data, err := json.Marshal(b)
		if err != nil {
			return fmt.Errorf("failed to marshal block: %w", err)
		}
		if _, err := f.Write(data); err != nil {
			return fmt.Errorf("failed to write block: %w", err)
		}
		if _, err := f.Write([]byte("\n")); err != nil {
			return fmt.Errorf("failed to write newline: %w", err)
		}
	}

	// ファイルを閉じてディスクにフラッシュ
	if err := f.Sync(); err != nil {
		return fmt.Errorf("failed to sync file: %w", err)
	}

	// アトミックに置き換え
	if err := os.Rename(tmpPath, s.path); err != nil {
		return fmt.Errorf("failed to rename file: %w", err)
	}

	return nil
}
