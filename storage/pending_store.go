package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"signet/core"
)

// PendingStore は承認待ちトランザクションの永続化を担当する
type PendingStore struct {
	path string
}

// NewPendingStore は新しいPendingStoreを作成する
func NewPendingStore(path string) *PendingStore {
	return &PendingStore{path: path}
}

// Load は承認待ちトランザクションを読み込む
// ファイルが存在しない場合は空スライスを返す
func (s *PendingStore) Load() ([]*core.PendingTransaction, error) {
	// ファイルが存在しない場合は空スライスを返す
	_, err := os.Stat(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return []*core.PendingTransaction{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	data, err := readFile(s.path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	if len(data) == 0 {
		return []*core.PendingTransaction{}, nil
	}

	var items []*core.PendingTransaction
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, fmt.Errorf("failed to unmarshal pending transactions: %w", err)
	}

	return items, nil
}

// Save は承認待ちトランザクションをJSON配列として書き出す
func (s *PendingStore) Save(items []*core.PendingTransaction) error {
	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal pending transactions: %w", err)
	}

	// 改行で終わるようにする
	data = append(data, '\n')

	if err := writeFile(s.path, string(data)); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// Add は承認待ちトランザクションを1つ追加する
func (s *PendingStore) Add(item *core.PendingTransaction) error {
	items, err := s.Load()
	if err != nil {
		return err
	}

	items = append(items, item)
	return s.Save(items)
}

// Remove は指定されたインデックスの承認待ちトランザクションを削除する
func (s *PendingStore) Remove(index int) error {
	items, err := s.Load()
	if err != nil {
		return err
	}

	if index < 0 || index >= len(items) {
		return fmt.Errorf("index out of range: %d", index)
	}

	// スライスから要素を削除
	items = append(items[:index], items[index+1:]...)
	return s.Save(items)
}

// Clear は全ての承認待ちトランザクションを削除する
func (s *PendingStore) Clear() error {
	return s.Save([]*core.PendingTransaction{})
}
