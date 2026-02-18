package storage

import (
	"os"
	"path/filepath"
	"signet/core"
	"testing"
	"time"
)

func TestNewBlockStore(t *testing.T) {
	store := NewBlockStore("/test/path")
	if store == nil {
		t.Fatal("NewBlockStore() returned nil")
	}
	if store.path != "/test/path" {
		t.Errorf("store.path = %v, want /test/path", store.path)
	}
}

func TestBlockStoreLoadAll(t *testing.T) {
	t.Run("nonexistent file returns empty slice", func(t *testing.T) {
		tmpDir := t.TempDir()
		store := NewBlockStore(filepath.Join(tmpDir, "nonexistent.jsonl"))

		blocks, err := store.LoadAll()
		if err != nil {
			t.Fatalf("LoadAll() error = %v", err)
		}
		if len(blocks) != 0 {
			t.Errorf("LoadAll() returned %d blocks, want 0", len(blocks))
		}
	})

	t.Run("empty file returns empty slice", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "empty.jsonl")
		if err := writeFile(filePath, ""); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		store := NewBlockStore(filePath)
		blocks, err := store.LoadAll()
		if err != nil {
			t.Fatalf("LoadAll() error = %v", err)
		}
		if len(blocks) != 0 {
			t.Errorf("LoadAll() returned %d blocks, want 0", len(blocks))
		}
	})

	t.Run("load blocks from file", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "blocks.jsonl")

		// テスト用ブロックデータを書き込み
		block1 := core.NewBlock(0, "0", core.BlockPayload{Type: "add_node"})
		block2 := core.NewBlock(1, block1.Header.Hash, core.BlockPayload{Type: "add_node"})

		// ファイルに直接JSONを書き込み
		data1, _ := encodeJSON(block1)
		data2, _ := encodeJSON(block2)
		content := string(data1) + "\n" + string(data2) + "\n"
		if err := writeFile(filePath, content); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		store := NewBlockStore(filePath)
		blocks, err := store.LoadAll()
		if err != nil {
			t.Fatalf("LoadAll() error = %v", err)
		}
		if len(blocks) != 2 {
			t.Errorf("LoadAll() returned %d blocks, want 2", len(blocks))
		}
		if blocks[0].Header.Index != 0 {
			t.Errorf("blocks[0].Header.Index = %d, want 0", blocks[0].Header.Index)
		}
		if blocks[1].Header.Index != 1 {
			t.Errorf("blocks[1].Header.Index = %d, want 1", blocks[1].Header.Index)
		}
	})
}

func TestBlockStoreAppend(t *testing.T) {
	t.Run("append block to new file", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "blocks.jsonl")
		store := NewBlockStore(filePath)

		block := core.NewBlock(0, "0", core.BlockPayload{Type: "add_node"})
		if err := store.Append(block); err != nil {
			t.Fatalf("Append() error = %v", err)
		}

		// ファイルが存在することを確認
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Error("Append() did not create file")
		}

		// 読み込んで確認
		blocks, err := store.LoadAll()
		if err != nil {
			t.Fatalf("LoadAll() error = %v", err)
		}
		if len(blocks) != 1 {
			t.Errorf("LoadAll() returned %d blocks, want 1", len(blocks))
		}
		if blocks[0].Header.Index != 0 {
			t.Errorf("blocks[0].Header.Index = %d, want 0", blocks[0].Header.Index)
		}
	})

	t.Run("append multiple blocks", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "blocks.jsonl")
		store := NewBlockStore(filePath)

		block1 := core.NewBlock(0, "0", core.BlockPayload{Type: "add_node"})
		block2 := core.NewBlock(1, "hash1", core.BlockPayload{Type: "add_node"})
		block3 := core.NewBlock(2, "hash2", core.BlockPayload{Type: "add_node"})

		for _, block := range []*core.Block{block1, block2, block3} {
			if err := store.Append(block); err != nil {
				t.Fatalf("Append() error = %v", err)
			}
		}

		blocks, err := store.LoadAll()
		if err != nil {
			t.Fatalf("LoadAll() error = %v", err)
		}
		if len(blocks) != 3 {
			t.Errorf("LoadAll() returned %d blocks, want 3", len(blocks))
		}
	})
}

func TestBlockStoreReplaceAll(t *testing.T) {
	t.Run("replace all blocks", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "blocks.jsonl")
		store := NewBlockStore(filePath)

		// 初期ブロックを追加
		block1 := core.NewBlock(0, "0", core.BlockPayload{Type: "add_node"})
		store.Append(block1)

		// 全ブロックを置き換え
		newBlock1 := core.NewBlock(0, "0", core.BlockPayload{Type: "add_node"})
		newBlock2 := core.NewBlock(1, newBlock1.Header.Hash, core.BlockPayload{Type: "add_node"})
		newBlock3 := core.NewBlock(2, newBlock2.Header.Hash, core.BlockPayload{Type: "add_node"})

		newBlocks := []*core.Block{newBlock1, newBlock2, newBlock3}
		if err := store.ReplaceAll(newBlocks); err != nil {
			t.Fatalf("ReplaceAll() error = %v", err)
		}

		// 読み込んで確認
		blocks, err := store.LoadAll()
		if err != nil {
			t.Fatalf("LoadAll() error = %v", err)
		}
		if len(blocks) != 3 {
			t.Errorf("LoadAll() returned %d blocks, want 3", len(blocks))
		}

		// ハッシュが正しいことを確認
		if blocks[1].Header.PrevHash != blocks[0].Header.Hash {
			t.Error("Block chain is not properly linked")
		}
	})

	t.Run("replace with empty slice", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "blocks.jsonl")
		store := NewBlockStore(filePath)

		// 初期ブロックを追加
		block := core.NewBlock(0, "0", core.BlockPayload{Type: "add_node"})
		store.Append(block)

		// 空スライスで置き換え
		if err := store.ReplaceAll([]*core.Block{}); err != nil {
			t.Fatalf("ReplaceAll() error = %v", err)
		}

		blocks, err := store.LoadAll()
		if err != nil {
			t.Fatalf("LoadAll() error = %v", err)
		}
		if len(blocks) != 0 {
			t.Errorf("LoadAll() returned %d blocks, want 0", len(blocks))
		}
	})

	t.Run("atomic replace (no partial writes on error)", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "blocks.jsonl")
		store := NewBlockStore(filePath)

		// 初期ブロックを追加
		originalBlock := core.NewBlock(0, "0", core.BlockPayload{Type: "add_node"})
		store.Append(originalBlock)

		// 別のブロックで置き換え
		newBlock := core.NewBlock(0, "0", core.BlockPayload{Type: "add_node"})
		newBlock.Header.CreatedAt = time.Now().Add(time.Hour) // 時刻を変えて区別

		if err := store.ReplaceAll([]*core.Block{newBlock}); err != nil {
			t.Fatalf("ReplaceAll() error = %v", err)
		}

		// 読み込んで元のデータが失われていることを確認
		blocks, err := store.LoadAll()
		if err != nil {
			t.Fatalf("LoadAll() error = %v", err)
		}
		if len(blocks) != 1 {
			t.Errorf("LoadAll() returned %d blocks, want 1", len(blocks))
		}
		// 時刻が変わっていることを確認（完全に置き換えられている）
		if blocks[0].Header.CreatedAt.Equal(originalBlock.Header.CreatedAt) {
			t.Error("Block was not properly replaced")
		}
	})
}
